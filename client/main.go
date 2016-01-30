package main

import (
	"log"

	"github.com/korylprince/jettison/lib/cache"
	"github.com/korylprince/jettison/lib/rpc"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	config, err := ParseEnv()
	if err != nil {
		log.Fatalln("Config parse error:", err)
	}
	log.Printf("Config: %#v\n", *config)

	conn, err := grpc.Dial(config.RPCServerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("GRPC connection error: %v", err)
	}
	defer conn.Close()

	var md metadata.MD = map[string][]string{"groups": config.Groups}
	ctx := metadata.NewContext(context.Background(), md)

	eventsClient := rpc.NewEventsClient(conn)
	fileClient := rpc.NewFileSetClient(conn)

	stream, err := eventsClient.Stream(ctx)
	if err != nil {
		log.Fatalf("stream create error: %v", err)
	}

	c, err := cache.NewBoltCache(config.CachePath)
	if err != nil {
		log.Fatalf("cache create error: %v", err)
	}

	fileService := NewFileService(config, c, fileClient)

	go NotificationService(fileService, stream)
	ReportService(config, stream, fileService)
}
