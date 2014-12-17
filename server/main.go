package main

import (
	"log"
	"net/http"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	envInit()
	fileInit()
	routesInit()
}

func main() {
	log.Println("Starting web server on", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, router))
}
