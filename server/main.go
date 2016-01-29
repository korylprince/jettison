package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/korylprince/jettison/lib/file"
	"github.com/korylprince/jettison/lib/rpc"

	"google.golang.org/grpc"
)

func main() {
	config, err := ParseEnv()
	if err != nil {
		log.Fatalln("Error parsing Config from env:", err)
	}
	log.Printf("Config: %#v\n", *config)

	files, err := file.FilesFromDefinition(config.DefinitionPath, config.CachePath)
	if err != nil {
		log.Fatalln("Error creating Files:", err)
	}
	defer files.Close()

	notifyService := NewNotifyService(config, files)

	mux := mux.NewRouter()
	mux.Methods("GET").PathPrefix("/file/").Handler(http.StripPrefix("/file/", http.FileServer(files)))
	mux.Methods("GET").Path("/sets").Handler(files)
	mux.Methods("POST").Path("/reload").Handler(notifyService)
	server := &http.Server{Addr: config.HTTPListenAddr, Handler: handlers.CombinedLoggingHandler(os.Stdout, mux)}

	go server.ListenAndServe()

	s := grpc.NewServer()
	rpc.RegisterFileSetServer(s, &fileService{Files: files})
	rpc.RegisterEventsServer(s, &eventService{NotifyService: notifyService})

	lis, err := net.Listen("tcp", config.RPCListenAddr)
	if err != nil {
		log.Fatalf("Error listening on %s: %v", config.RPCListenAddr, err)
	}
	s.Serve(lis)
}
