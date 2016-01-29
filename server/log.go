package main

import (
	"log"

	"golang.org/x/net/context"

	"google.golang.org/grpc/peer"
)

//LogGRPC writes a log message with a GRPC context
func LogGRPC(ctx context.Context, src, msg string) {
	p, ok := peer.FromContext(ctx)
	if ok {
		log.Printf("%s (%s): %s", src, p.Addr, msg)
	} else {
		log.Printf("%s (No Address): %s\n", src, msg)
	}
}
