package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

//Config stores configuration from the environment
type Config struct {
	Podbay    string
	Room      string
	ConfigURL string
	APIURL    string
	Interval  int //minutes
}

type jsonConfig struct {
	ListenAddr string `json:"listenaddr"`
}

const apiPrefix = "/api/v1"

var config = &Config{}

func cacheFallback(err error) (string, error) {
	log.Println("Unable to fetch Config:", err)
	addr, err := cache.ListenAddr()
	if err != nil {
		return "", fmt.Errorf("Error checking Cache for Config: %s", err)
	}
	if addr == "" {
		return "", fmt.Errorf("Config is not Cached; Cannot Continue")
	}
	return addr, nil
}

func init() {
	log.SetOutput(os.Stdout)

	err := envconfig.Process("JETTISON", config)
	if err != nil {
		log.Fatalln("Error reading configuration from environment:", err)
	}
	if config.Podbay == "" {
		log.Fatalln("JETTISON_PODBAY must be configured")
	}
	if config.Podbay[len(config.Podbay)-1:] != "/" {
		config.Podbay += "/"
	}
	if config.Room == "" {
		log.Fatalln("JETTISON_ROOM must be configured")
	}
	if config.ConfigURL == "" {
		log.Fatalln("JETTISON_CONFIGURL must be configured")
	}
	if config.Interval == 0 {
		config.Interval = 10
		log.Println("JETTISON_INTERVAL not set; Defaulted to 10 minutes")
	}

	cacheInit()

	for {
		addr, err := pollServer()
		if err != nil {
			log.Println("Error getting Listen Address:", err)
		} else if addr == "" {
			log.Println("Error getting Listen Address: Address is Empty")
		} else {
			err = cache.SetListenAddr(addr)
			if err == nil {
				config.APIURL = "http://" + addr + apiPrefix
				break
			}
			log.Println("Error updating cache:", err)
		}
		time.Sleep(NewInterval())
	}

	log.Printf("Config: %#v\n", *config)
}

//getServer tries to update the Listen Address from the config url
//or falls back to the cache if possible
func pollServer() (string, error) {
	log.Println("Fetching Config from", config.ConfigURL)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(config.ConfigURL)
	if err != nil {
		return cacheFallback(err)
	}
	defer resp.Body.Close()

	j := jsonConfig{}
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&j)
	if err != nil {
		return cacheFallback(err)
	}

	if j.ListenAddr == "" {
		return cacheFallback(fmt.Errorf("ListenAddr not present is Config"))
	}

	return j.ListenAddr, nil
}
