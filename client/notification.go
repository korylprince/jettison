package main

import (
	"log"
	"os"

	"github.com/korylprince/jettison/lib/rpc"
)

//NotificationService is a GRPC NotificationService
func NotificationService(fileService *FileService, stream rpc.Events_StreamClient) {
	log.Println("Notification: Service Started")
	for {
		n, err := stream.Recv()
		if err != nil {
			log.Printf("Notification: EXITING, Error: %v\n", err)
			os.Exit(1)
		}
		log.Printf("Notification: Group: %s, Version: %d\n", n.GetGroup(), n.GetVersion())
		fileService.Scan(n.GetGroup())
	}
}
