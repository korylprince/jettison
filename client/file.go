package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/korylprince/jettison/lib/cache"
	"github.com/korylprince/jettison/lib/file"
	"github.com/korylprince/jettison/lib/rpc"
)

//FileService manages the local files for the jettison client
type FileService struct {
	config *Config
	cache  cache.Cache
	client rpc.FileSetClient
	sets   map[string]*file.VersionedSet //group:VersionedSet
	mu     *sync.RWMutex

	scan chan []string //chan groups
}

//NewFileService returns a new FileService
func NewFileService(config *Config, c cache.Cache, client rpc.FileSetClient) *FileService {
	f := &FileService{
		config: config,
		cache:  c,
		client: client,
		sets:   make(map[string]*file.VersionedSet),
		mu:     new(sync.RWMutex),
		scan:   make(chan []string, len(config.Groups)),
	}
	go f.timer()
	return f
}

//Scan causes the FileService to rescan the groups
func (s *FileService) Scan(groups ...string) {
	s.scan <- groups
}

//Versions returns the current FileSet versions
func (s *FileService) Versions() map[string]uint64 {
	//map[group]version
	v := make(map[string]uint64)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for group, vs := range s.sets {
		v[group] = vs.Version
	}
	return v
}

func (s *FileService) timer() {
	groups := s.config.Groups
	for {
		err := s.check(groups...)
		if err != nil {
			log.Println("FileService: Error downloading files:", err)
		}
		select {
		case <-time.After(s.config.CheckInterval * time.Second):
			groups = s.config.Groups
		case groups = <-s.scan:
		}
	}
}

func (s *FileService) check(groups ...string) error {
	resp, err := s.client.Get(context.Background(), &rpc.FileSetRequest{Groups: groups})
	if err != nil {
		return fmt.Errorf("FileSetRequest error: %v", err)
	}

	//convert fileset
	var grps sort.StringSlice
	sets := make(map[string]*file.VersionedSet)
	for group, set := range resp.Sets {
		sets[group] = &file.VersionedSet{Set: set.Set, Version: set.Version}
		grps = append(grps, fmt.Sprintf("{Group: %s, Len: %d, Version: %d}", group, len(set.Set), set.Version))
	}
	grps.Sort()
	log.Printf("FileSetResponse: %s\n", strings.Join(grps, ", "))

	//walk and download
	return s.walk(sets)
}

func (s *FileService) walk(sets map[string]*file.VersionedSet) error {
	for group, vs := range sets {
		for hash, path := range vs.Set {
			_, _, err := s.cache.Get(path)
			if err == cache.ErrorInvalidCacheEntry {
				err = Download(fmt.Sprintf("http://%s/file/%d", s.config.HTTPServerAddr, hash), path, hash)
				if err != nil {
					return fmt.Errorf("Download: Error: %v", err)
				}
				log.Printf("Download: Path: %s, Hash: %d\n", path, hash)

				err = s.cache.Put(path, hash, time.Now())
				if err != nil {
					return fmt.Errorf("Cache.Put error: %v", err)
				}
			} else if err != nil {
				return fmt.Errorf("Cache.Get error: %v", err)
			}
		}

		//everything has been downloaded and cached so update version
		s.mu.Lock()
		s.sets[group] = vs
		s.mu.Unlock()
	}
	return nil
}

//Download downloads url to path, verifing that the file's hash matches hash
func Download(url, path string, hash uint64) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Error getting %s: %v", url, err)
	}
	defer resp.Body.Close()

	err = os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		return fmt.Errorf("Error creating directory %s: %v", filepath.Dir(path), err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Error creating file %s: %v", path, err)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		return fmt.Errorf("Error writing to file %s: %v", path, err)
	}
	f.Close()

	h, err := file.Hash(path)
	if err != nil {
		return fmt.Errorf("Error hashing file %s: %v", path, err)
	}

	if hash != h {
		return fmt.Errorf("Hash mismatch on file %s: Expected %d, Result: %d", path, hash, h)
	}

	return nil
}
