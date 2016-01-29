package file

import (
	"time"

	"github.com/korylprince/jettison/lib/cache"
)

func _download(hash uint64, path string) error {
	return nil
}

func download(sets map[string]*VersionedSet, c cache.Cache) error {
	for _, vs := range sets {
		for hash, path := range vs.Set {
			_, _, err := c.Get(path)
			if err == cache.ErrorInvalidCacheEntry {
				err = _download(hash, path)
				if err != nil {
					return err
				}
				err = c.Put(path, hash, time.Now())
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
		}
	}
	return nil
}
