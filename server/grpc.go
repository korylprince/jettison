package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/korylprince/jettison/lib/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

//FileSetServer is a GRPC FileSetService
type FileSetServer struct {
	Files *FileService
}

//Get returns FileSets for the groups requested
func (s FileSetServer) Get(ctx context.Context, r *rpc.FileSetRequest) (*rpc.FileSetResponse, error) {
	sets := s.Files.Sets(r.GetGroups()...)
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

//EventServer is a GRPC EventService
type EventServer struct {
	NotifyService *NotifyService
}

//Stream registers the stream for the groups included in metadata and saves reports to the database
func (s EventServer) Stream(stream rpc.Events_StreamServer) error {
	//register for notifications
	if md, ok := metadata.FromContext(stream.Context()); ok {
		if groups, ok := md["groups"]; ok && groups != nil {
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
		LogGRPC(stream.Context(), "Report", fmt.Sprintf("HardwareAddr: %s, Location: %s, Version: %v",
			rpt.GetHardwareAddr(), rpt.GetLocation(), rpt.GetVersion()))
	}
}

//Report saves the report to the database
func Report(rpt *rpc.Report) {
	//TODO
}
