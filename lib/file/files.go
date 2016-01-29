package file

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"golang.org/x/net/context"

	"github.com/korylprince/jettison/lib/cache"
)

//Files is a thread-safe access to file sets
type Files struct {
	cache cache.Cache
	sets  map[string]*VersionedSet //group:VersionedSet
	mu    *sync.RWMutex
}

//FilesFromDefinition returns a new Files with the given definition and cache paths or an error if one occurred
func FilesFromDefinition(defPath, cachePath string) (*Files, error) {
	c, err := cache.NewBoltCache(cachePath)
	if err != nil {
		return nil, err
	}
	f := &Files{cache: c, mu: new(sync.RWMutex)}
	_, err = f.CheckDefinition(defPath)
	return f, err
}

//Origin returns the origin path for the given path, if ok is true
func (f *Files) Origin(hash uint64) (path string, ok bool) {
	f.mu.RLock()
	if _, ok = f.sets["_origin"]; ok {
		path, ok = f.sets["_origin"].Set[hash]
	} else {
		panic(fmt.Errorf("f.sets[\"_origin\"] accessed but doesn't exist"))
	}
	f.mu.RUnlock()
	return path, ok
}

//Sets returns VersionedSets for the given groups. The caller should not modify the result
func (f *Files) Sets(groups ...string) map[string]*VersionedSet {
	sets := make(map[string]*VersionedSet)
	f.mu.RLock()
	for _, group := range groups {
		if vs, ok := f.sets[group]; ok {
			sets[group] = vs
		}
	}
	f.mu.RUnlock()
	return sets
}

//CheckDefinition causes f to reread the definition and filesystem for changes
//CheckDefinition returns changed, a map[group]version of any groups that changed versions
//CheckDefinition blocks until finished or returns an error if one occurred
func (f *Files) CheckDefinition(defPath string) (changed map[string]uint64, err error) {
	def, err := Parse(defPath)
	if err != nil {
		return nil, err
	}
	all, mapped, err := WalkDefinition(context.Background(), def, f.cache, 10)
	if err != nil {
		return nil, err
	}

	f.mu.Lock()

	changed = make(map[string]uint64)
	//check if versior changed
	for group, new := range mapped {
		if old, ok := f.sets[group]; ok {
			if group != "_origin" && new.Version != old.Version {
				changed[group] = new.Version
			}
		} else {
			changed[group] = new.Version
		}
	}
	mapped["_origin"] = &VersionedSet{Set: all}
	f.sets = mapped

	f.mu.Unlock()
	return changed, nil
}

//Open statisfies http.FileSystem
func (f *Files) Open(hash string) (http.File, error) {
	h, err := strconv.ParseUint(hash[1:], 10, 64)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: hash[1:], Err: os.ErrNotExist}
	}
	path, ok := f.Origin(h)
	if !ok {
		return nil, &os.PathError{Op: "open", Path: hash[1:], Err: os.ErrNotExist}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

//Close closes the underlying cache
func (f *Files) Close() error {
	return f.cache.Close()
}

//ServeHTTP satisfies http.Handler, returning the underlying sets in JSON or an error if one occurred
func (f *Files) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	e := json.NewEncoder(w)
	f.mu.RLock()
	err := e.Encode(f.sets)
	f.mu.RUnlock()
	if err != nil {
		log.Println("Files: Error encoding JSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error":%d,"msg":"%s"}`, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))))
	}
}
