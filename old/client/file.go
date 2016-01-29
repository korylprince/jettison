package main

import (
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	pathlib "path"
)

//Hash takes the path to a file and returns the crc32 hash,
//or an error if one occurred
func Hash(path string) (uint32, error) {
	h := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	_, err = io.Copy(h, f)
	if err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

//Walk recurses the given directory, caching file names and hashes
//in the global cache
func Walk(prefix, path string) {
	root := prefix + path
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.Printf("Error occurred trying to read contents of %s: %s\n", root, err)
		return
	}
	for _, f := range files {
		mode := f.Mode()
		rootName := root + f.Name()
		switch {
		case mode.IsDir():
			Walk(prefix, path+f.Name()+"/")
		case mode.IsRegular():
			h, err := Hash(rootName)
			if err != nil {
				log.Printf("Error hashing file %s: %s\n", rootName, err)
			}
			cache.SetFile(path+f.Name(), h)
		default:
			log.Println("Skipping non-regular file:", rootName)
		}
	}
}

//Write writes the data in r to path and verifies that its
//crc32 hash is the same as crc, returning an error if any
//of this fails
func Write(path string, crc uint32, r io.ReadCloser) error {
	defer r.Close()
	dir := config.Podbay + pathlib.Dir(path)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}
	f, err := os.Create(config.Podbay + path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	h, err := Hash(config.Podbay + path)
	if err != nil {
		return err
	}
	if h != crc {
		return fmt.Errorf("CRC mismatch on: %s; Got: %#v; Should be: %#v", path, h, crc)
	}
	return nil
}
