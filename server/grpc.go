package main

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/korylprince/jettison/lib/file"
	"github.com/korylprince/jettison/lib/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type fileService struct {
	Files *file.Files
}

func (s fileService) Get(ctx context.Context, r *rpc.FileSetRequest) (*rpc.FileSetResponse, error) {
	sets := s.Files.Sets(r.Groups...)
	var grps sort.StringSlice
	resp := &rpc.FileSetResponse{Sets: make(map[string]*rpc.FileSetResponse_VersionedSet)}
	for group, set := range sets {
		resp.Sets[group] = &rpc.FileSetResponse_VersionedSet{Set: set.Set, Version: set.Version}
		grps = append(grps, fmt.Sprintf("%s:%d", group, set.Version))
	}
	grps.Sort()
	LogGRPC(ctx, "FileSetRequest", strings.Join(grps, ", "))
	return resp, nil
}

type eventService struct {
	NotifyService *NotifyService
}

func (s eventService) Stream(stream rpc.Events_StreamServer) error {
	//register for notifications
	if md, ok := metadata.FromContext(stream.Context()); ok {
		if groups, ok := md["groups"]; ok {
			s.NotifyService.Register(stream, groups...)
			LogGRPC(stream.Context(), "Register", fmt.Sprintf("Groups: %s", strings.Join(groups, ", ")))
			defer func() {
				s.NotifyService.Unregister(stream, groups...)
				LogGRPC(stream.Context(), "Unregister", fmt.Sprintf("Groups: %s", strings.Join(groups, ", ")))
			}()
		}
	}
	for {
		rpt, err := stream.Recv()
		if err != nil {
			LogGRPC(stream.Context(), "Report", fmt.Sprintf("Error: %v", err))
			return err
		}
		Report(rpt)
		LogGRPC(stream.Context(), "Report", fmt.Sprintf("Serial: %s, MAC Address: %s, Location: %s, Version: %v",
			rpt.Serial, net.HardwareAddr(rpt.MacAddress), rpt.Location, rpt.Version))
	}
}

func Report(rpt *rpc.Report) {
	//TODO
}
