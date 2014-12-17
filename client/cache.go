package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/boltdb/bolt"
)

//Cache holds both the file cache and the listen address
//of the server
type Cache struct {
	db *bolt.DB
}

//ListenAddr returns the cached Listen Address
func (c *Cache) ListenAddr() (string, error) {
	var addr string
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("listen"))
		if b == nil {
			return fmt.Errorf("ListenAddr not set")
		}
		val := b.Get([]byte("addr"))
		if val == nil {
			return fmt.Errorf("ListenAddr not set")
		}
		addr = string(val)
		return nil
	})
	if err != nil {
		return "", err
	}
	return addr, nil
}

//SetListenAddr sets the Listen Address in the cache
func (c *Cache) SetListenAddr(addr string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("listen"))
		if err != nil {
			return err
		}
		err = b.Put([]byte("addr"), []byte(addr))
		if err != nil {
			return err
		}
		return nil
	})
}

//File returns the crc32 hash of a file from the cache if it exists
func (c *Cache) File(path string) (crc32 uint32, ok bool, err error) {
	ok = false
	err = c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("files"))
		if b == nil {
			return nil
		}
		val := b.Get([]byte(path))
		if val == nil {
			return nil
		}
		i, err := strconv.ParseUint(string(val), 10, 32)
		if err != nil {
			return err
		}
		ok = true
		crc32 = uint32(i)
		return nil
	})
	if err != nil {
		return 0, false, err
	}
	return crc32, ok, nil
}

//SetFile sets the crc32 hash for a file in the cache
func (c *Cache) SetFile(path string, crc32 uint32) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("files"))
		if err != nil {
			return err
		}
		err = b.Put([]byte(path), []byte(strconv.FormatUint(uint64(crc32), 10)))
		if err != nil {
			return err
		}
		return nil
	})
}

var cache = Cache{}

func cacheInit() {
	db, err := bolt.Open(config.Podbay+".cache", 0666, nil)

	//if error opening cache, naively just create one
	if err != nil {
		log.Println("Error reading", config.Podbay+".cache:", err)
		log.Println("Creating new Cache at", config.Podbay+".cache")
		err = os.MkdirAll(config.Podbay, 0777)
		if err != nil {
			log.Fatalln("Not able to create Podbay:", config.Podbay+":", err)
		}
		err = os.Remove(config.Podbay + ".cache")
		if err != nil && err.(*os.PathError).Err != syscall.ENOENT { // ignore no such file or directory
			log.Fatalln("Not able to create Cache", config.Podbay+".cache:", err)
		}
		db, err = bolt.Open(config.Podbay+".cache", 0666, nil)
		if err != nil {
			log.Fatalln("Not able to create Cache", config.Podbay+".cache:", err)
		}
	}
	cache.db = db

	if err != nil {
		log.Println("Caching files in", config.Podbay)
		start := time.Now()
		Walk(config.Podbay, "/")
		log.Println("Cache generated in", time.Since(start))
	}

	err = hide(config.Podbay + ".cache")
	if err != nil {
		log.Println("Unable to hide cache file:", err)
	}
}
