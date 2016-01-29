package cache

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

//ErrorInvalidCacheEntry signals that the given path has an invalid or empty cache entry
var ErrorInvalidCacheEntry = fmt.Errorf("invalid cache entry")

//Cache is an interface for storing file metadata
type Cache interface {
	Get(path string) (hash uint64, mtime time.Time, err error)
	Put(path string, hash uint64, mtime time.Time) error
	Close() error
}

//BoltCache is an implementation of Cache on top of boltdb
type BoltCache struct {
	db *bolt.DB
}

//NewBoltCache returns a new Cache
func NewBoltCache(path string) (Cache, error) {
	db, err := bolt.Open(path, 0644, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, txErr := tx.CreateBucketIfNotExists([]byte("files"))
		return txErr
	})
	if err != nil {
		return nil, err
	}
	return Cache(&BoltCache{db: db}), nil
}

//Get returns the hash and mtime for the given path, or an error if one occurred
func (c *BoltCache) Get(path string) (hash uint64, mtime time.Time, err error) {
	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		if b == nil {
			return fmt.Errorf("invalid bucket: files")
		}
		v := b.Get([]byte(path))
		if v == nil || len(v) != 23 {
			return ErrorInvalidCacheEntry
		}
		err = mtime.UnmarshalBinary(v[0:15]) //length of binary encoded time.Time
		if err != nil {
			return err
		}
		hash = binary.BigEndian.Uint64(v[15:23])
		return nil
	})
	return hash, mtime, err
}

//Put sets the hash and mtime for the given path or will return an error if one occurred
func (c *BoltCache) Put(path string, hash uint64, mtime time.Time) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		t, err := mtime.MarshalBinary()
		if err != nil {
			return err
		}

		h := make([]byte, 8)
		binary.BigEndian.PutUint64(h, hash)

		b := tx.Bucket([]byte("files"))
		if b == nil {
			return fmt.Errorf("invalid bucket: files")
		}
		return b.Put([]byte(path), append(t, h...))
	})
	return err
}

//Close closes the underlying boltdb database
func (c *BoltCache) Close() error {
	return c.db.Close()
}
