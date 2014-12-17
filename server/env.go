package main

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

//Config stores configuration from the environment
type Config struct {
	ListenAddr string
	Podbay     string
}

var config = &Config{}

func envInit() {
	err := envconfig.Process("JETTISON", config)
	if err != nil {
		log.Panicln("Error reading configuration from environment:", err)
	}
	if config.ListenAddr == "" {
		log.Fatalln("JETTISON_LISTENADDR must be configured")
	}
	if config.Podbay == "" {
		log.Fatalln("JETTISON_PODBAY must be configured")
	}

	if config.Podbay[len(config.Podbay)-1:] != "/" {
		config.Podbay += "/"
	}

	// check that directories exist
	if _, err := os.Stat(config.Podbay); os.IsNotExist(err) {
		log.Fatalln(err)
	}
	log.Printf("Config: %#v\n", *config)
}
