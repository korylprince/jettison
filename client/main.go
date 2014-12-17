package main

import (
	"log"
	"time"
)

func updateServer() {
	for {
		time.Sleep(NewInterval() * 3)
		log.Println("Updating Config")
		addr, err := pollServer()
		if err != nil {
			log.Println("Error getting Listen Address:", err)
		} else if addr == "" {
			log.Println("Error getting Listen Address: Address is Empty")
		} else {
			err = cache.SetListenAddr(addr)
			if err != nil {
				log.Println("Error updating cache:", err)
			} else {
				config.APIURL = "http://" + addr + apiPrefix
				log.Println("Updating API URL to:", config.APIURL)
			}
		}
	}
}

func main() {
	log.Println("Starting jettison client")
	go updateServer()
	for {
		Fetch()
		time.Sleep(NewInterval())
	}
}
