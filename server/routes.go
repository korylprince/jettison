package main

import (
	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

var router http.Handler

const apiPrefix = "/api/v1"

//FileListHandler is an http.HandlerFunc that will return a json representation
//of all the files of the given room as well as the "all" room along with
//the crc32 hash of the file for verification
func FileListHandler(rw http.ResponseWriter, r *http.Request) {
	room := mux.Vars(r)["room"]
	files := GetFiles(room)
	e := json.NewEncoder(rw)
	err := e.Encode(files)
	if err != nil {
		log.Panicln("Error encoding json:", err)
	}
}

func routesInit() {
	r := mux.NewRouter()
	r.Handle(apiPrefix+"/list/{room}", http.HandlerFunc(FileListHandler)).Methods("GET")
	fileServer := http.StripPrefix(apiPrefix+"/file", http.FileServer(http.Dir(config.Podbay)))
	r.PathPrefix(apiPrefix + "/file").Handler(fileServer).Methods("GET")
	router = handlers.CompressHandler(handlers.CombinedLoggingHandler(os.Stdout, r))
}
