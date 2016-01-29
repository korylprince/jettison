package main

import (
	"encoding/json"
	"log"
	"net/http"
)

//Files is a map of filename to hash
type Files map[string]uint32

func getList() (Files, error) {
	resp, err := http.Get(config.APIURL + "/list/" + config.Room)
	if err != nil {
		return nil, err
	}
	list := make(Files)
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func fetchFile(path string, crc32 uint32) error {
	resp, err := http.Get(config.APIURL + "/file/" + path)
	if err != nil {
		return err
	}
	err = Write(path, crc32, resp.Body)
	if err != nil {
		return err
	}
	return cache.SetFile(path, crc32)
}

//Fetch gets a list of files and hashes from the server
//and attempts to download them
func Fetch() {
	files, err := getList()
	if err != nil {
		log.Println("Unable to get file list:", err)
		return
	}
	for path, crc32 := range files {
		cachedCRC32, ok, err := cache.File(path)
		if err != nil {
			log.Println("Error checking cache:", err)
		}
		if !(ok && cachedCRC32 == crc32) {
			err := fetchFile(path, crc32)
			if err != nil {
				log.Println("Error downloading file:", path, ":", err)
			} else {
				log.Println("File downloaded:", path)
			}
		}
	}
}
