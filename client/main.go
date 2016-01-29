package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/context"

	"github.com/korylprince/jettison/lib/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var groups = []string{"all", "rand"}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial("localhost:50081", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := rpc.NewFileSetClient(conn)
	resp, err := c.Get(context.Background(), &rpc.FileSetRequest{Groups: groups})
	if err != nil {
		log.Fatalf("request error: %v", err)
	}

	for group, vs := range resp.Sets {
		for hash, path := range vs.Set {
			log.Println(path)
			f, err := http.Get(fmt.Sprintf("http://localhost:50080/file/%d", hash))
			if err != nil {
				log.Fatalf("file request error: %v", err)
			}
			defer f.Body.Close()
			io.Copy(ioutil.Discard, f.Body)
		}
		log.Printf("Group: %s, Version: %d, Entries: %d\n", group, vs.Version, len(vs.Set))
	}

	var md metadata.MD = map[string][]string{"groups": groups}
	ctx := metadata.NewContext(context.Background(), md)

	events := rpc.NewEventsClient(conn)
	stream, err := events.Stream(ctx)
	if err != nil {
		log.Fatalf("stream create error: %v", err)
	}

	iface, err := net.InterfaceByName("enp3s0")
	if err != nil {
		log.Fatalf("interface get error: %v", err)
	}

	rpt := &rpc.Report{Serial: "12345", MacAddress: iface.HardwareAddr, Location: "Here", Version: nil}
	err = stream.Send(rpt)
	if err != nil {
		log.Fatalf("stream send error: %v", err)
	}

	for i := 0; i < 10; i++ {
		n, err := stream.Recv()
		if err != nil {
			log.Fatalf("stream receive error: %v", err)
		}
		log.Printf("Notification: %#v\n", n)
	}
	err = stream.CloseSend()
	if err != nil {
		log.Fatalf("stream close error: %v", err)
	}
}

/*

case: Path in List but not in Cache:
	Download Path, verify hash, put in cache
case: Path in List, Path in Cache, hashes match
	Done
case: Path in List, Path in Cache, hashes don't match (hash in cache doesn't match list)
	list modtime newer than cache mtime


*/

/*

1. startup
2. fetch filelist from server
3. walk files on disk
4. if path from list in cache, skip
	* if not in cache, download

*/
