package main

import (
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

//Files represents a mapping of filenames to hashes of those files
type Files map[string]uint32

//Cache is a global file cache
//map[room][file path]hash
type Cache map[string]Files

var cache = make(Cache)

func fileInit() {
	log.Println("Caching files in", config.Podbay)
	start := time.Now()

	rooms, err := ioutil.ReadDir(config.Podbay)
	if err != nil {
		log.Printf("Error occurred trying to read contents of %s: %s\n", config.Podbay, err)
		return
	}
	for _, r := range rooms {
		if r.Mode().IsDir() {
			room := r.Name()
			cache[room] = make(Files)
			walk(config.Podbay, room+"/", room)
		}
	}
	log.Println("Cache generated in", time.Since(start))
}

//hash takes the path to a file and returns the crc32 hash,
//or an error if one occurred
func hash(path string) (uint32, error) {
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

//walk recurses the given directory, caching file names and hashes
//under the given room in the global cache
func walk(prefix, path, room string) {
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
			walk(prefix, path+f.Name()+"/", room)
		case mode.IsRegular():
			h, err := hash(rootName)
			if err != nil {
				log.Printf("Error hashing file %s: %s\n", rootName, err)
			}
			cache[room][path+f.Name()] = h
		default:
			log.Println("Skipping non-regular file:", rootName)
		}
	}
}

//GetFiles returns a recursive list of files and their hashes in
//config.Podbay/all/ and config.Podbay/room/
func GetFiles(room string) Files {
	f := make(Files)
	if c, ok := cache[room]; ok {
		for k, v := range c {
			f[k] = v
		}
	}
	if room == "all" {
		return f
	}
	if c, ok := cache["all"]; ok {
		for k, v := range c {
			f[k] = v
		}
	}
	return f
}
