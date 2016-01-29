package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/korylprince/jettison/lib/cache"
)

//fileInfo represents metadata about a file
type fileInfo struct {
	Hash    uint64
	ModTime time.Time
	Path    string
}

//renamePath rewrites path, substituting dest for origin.
//Any error causes a panic as any path should exist that is passed
func renamePath(path, origin, dest string) string {
	p, err := filepath.Rel(origin, path)
	if err != nil {
		panic(err) //we got here with filepath.Walk, should never give error
	}
	return filepath.Join(dest, p)
}

//WalkDefinition walks d, returning all, a Set with origin paths, mapped a map of Sets with destination paths split by groups,
//or an error if one occurred. WalkDefinition will use cache as hash cache and workers for the number of workers.
func WalkDefinition(ctx context.Context, d Definition, c cache.Cache, workers int) (all Set, mapped map[string]*VersionedSet, err error) {
	m := make(map[string]*VersionedSet)
	all = make(Set)
	for group, mapping := range d {
		m[group] = &VersionedSet{Set: make(Set), Version: 0}
		for origin, dest := range mapping {
			s := make(Set)

			err := walkRoot(ctx, c, s, origin, workers)
			if err != nil {
				return nil, nil, err
			}

			//copy to all
			for hash, path := range s {
				all[hash] = path
			}

			//rewrite paths
			for hash, path := range s {
				//origin is file
				if origin == path {
					m[group].Set[hash] = path
				} else {
					m[group].Set[hash] = renamePath(path, origin, dest)
				}
				m[group].Version += hash
			}
		}
	}
	return all, m, nil
}

func walkRoot(ctx context.Context, c cache.Cache, s Set, root string, workers int) error {
	rootctx, rootCancel := context.WithCancel(ctx)
	accctx, accCancel := context.WithCancel(ctx)
	infos := make(chan *fileInfo)
	wg := new(sync.WaitGroup)
	var rerr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		rerr = rootWalker(rootctx, root, infos)
		if rerr != nil {
			accCancel()
		}
	}()

	err := accumulator(accctx, s, c, infos, workers)

	if err != nil {
		rootCancel()
		return err
	}

	wg.Wait()
	return rerr
}

//rootWalker passes an *Info for every path under root to out, returning the
//first error encountered, if any. If ctx is cancelled, rootWalker returns at earliest opportunity
func rootWalker(ctx context.Context, root string, out chan<- *fileInfo) error {
	defer close(out)
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Error walking path %s: %v", path, err)
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		select {
		case <-ctx.Done(): //cancelled
			return ctx.Err()
		case out <- &fileInfo{Path: path, ModTime: info.ModTime()}:
			return nil
		}
	})
}

//pathWalker passes an *Info for every path coming from in to out, returning the
//first error encountered, if any. If ctx is cancelled, pathWalker returns at earliest opportunity
func pathWalker(ctx context.Context, in <-chan string, out chan<- *fileInfo) error {
	defer close(out)
	for {
		path, ok := <-in
		if !ok {
			return nil
		}

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("Error stating path %s: %v", path, err)
		}

		if !info.Mode().IsRegular() {
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- &fileInfo{Path: path, ModTime: info.ModTime()}:
			return nil
		}
	}
}

//Accumulator adds hashed paths given on in to set.
func accumulator(ctx context.Context, set Set, c cache.Cache, in <-chan *fileInfo, workers int) error {

	wg := new(sync.WaitGroup)
	out := make(chan *fileInfo, workers)
	errors := make(chan error)
	var err error
	sub, cancel := context.WithCancel(ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case err = <-errors:
				cancel()
				return
			case info, ok := <-out:
				if !ok {
					return
				}
				set[info.Hash] = info.Path
			}
		}
	}()
	concurrentHasher(sub, c, in, out, errors, workers)
	wg.Wait()

	return err
}

//sendError is a helper function to send an error on errors if ctx has not been cancelled
func sendError(ctx context.Context, errors chan<- error, err error) {
	select {
	case <-ctx.Done(): //cancelled
		return
	case errors <- err:
		return
	}
}

//hasher takes *Infos from in and outputs them to out after computing the hash or getting it from the cache
//errors are sent on errors. If ctx is cancelled, hasher will exit at the earliest opportunity
func hasher(ctx context.Context, wg *sync.WaitGroup, c cache.Cache, in <-chan *fileInfo, out chan<- *fileInfo, errors chan<- error) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case info, ok := <-in:

			if !ok {
				//no more paths
				return
			}

			//check cache
			hash, mtime, err := c.Get(info.Path)
			if err == nil && !info.ModTime.After(mtime) {
				goto sendHash
			}
			if err != nil && err != cache.ErrorInvalidCacheEntry {
				sendError(ctx, errors, fmt.Errorf("Error getting cache entry %s: %v", info.Path, err))
				return
			}

			//compute hash
			hash, err = Hash(info.Path)
			if err != nil {
				sendError(ctx, errors, fmt.Errorf("Error hashing %s: %v", info.Path, err))
				return
			}

			//store hash in cache
			err = c.Put(info.Path, hash, info.ModTime)
			if err != nil {
				sendError(ctx, errors, fmt.Errorf("Error putting cache entry %s: %v", info.Path, err))
				return
			}

		sendHash:

			info.Hash = hash

			select {
			case <-ctx.Done(): //cancelled
				return
			case out <- info:
			}
		}
	}
}

//concurrentHasher takes *Infos from in and outputs them to out after computing the hash or getting it from the cache
//workers specifies how many hasher goroutines are run at once.
//errors are sent on errors. If ctx is cancelled, concurrentHasher will exit at the earliest opportunity
func concurrentHasher(ctx context.Context, c cache.Cache, in <-chan *fileInfo, out chan<- *fileInfo, errors chan<- error, workers int) {
	defer close(out)

	wg := new(sync.WaitGroup)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go hasher(ctx, wg, c, in, out, errors)
	}

	wg.Wait() //wait for hashers to exit
}
