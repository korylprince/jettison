package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/korylprince/jettison/lib/rpc"
)

//NotifyService registers event streams for notifications of group version updates
type NotifyService struct {
	config   *Config
	files    *FileService
	registry map[string]map[rpc.Events_StreamServer]struct{} //group:set{connections}
	mu       *sync.RWMutex
}

//NewNotifyService returns a new NotifyService
func NewNotifyService(config *Config, files *FileService) *NotifyService {
	return &NotifyService{
		config:   config,
		files:    files,
		registry: make(map[string]map[rpc.Events_StreamServer]struct{}),
		mu:       new(sync.RWMutex),
	}
}

//Register registers stream to receive notifications for the given groups
func (s *NotifyService) Register(stream rpc.Events_StreamServer, groups ...string) {
	s.mu.Lock()
	for _, g := range groups {
		if _, ok := s.registry[g]; !ok {
			s.registry[g] = make(map[rpc.Events_StreamServer]struct{})
		}
		s.registry[g][stream] = struct{}{}
	}
	s.mu.Unlock()
}

//Unregister unregisters stream to receive notifications for the given groups
func (s *NotifyService) Unregister(stream rpc.Events_StreamServer, groups ...string) {
	s.mu.Lock()
	for _, g := range groups {
		if _, ok := s.registry[g]; ok {
			delete(s.registry[g], stream)
		}
	}
	s.mu.Unlock()
}

//Notify notifies the streams (if any) registered to the given groups of version changes
func (s *NotifyService) Notify(groups map[string]uint64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for group, ver := range groups {
		if _, ok := s.registry[group]; ok {
			for stream := range s.registry[group] {
				err := stream.Send(&rpc.Notification{Group: group, Version: ver})
				if err != nil {
					LogGRPC(stream.Context(), "Notification", fmt.Sprintf("Error: %v", err))
					return err
				}
				LogGRPC(stream.Context(), "Notification", fmt.Sprintf("Group: %s, Version: %d", group, ver))
			}
		}
	}
	return nil
}

//ServeHTTP satisfies http.Handler, reloading the underlying Definition and Files,
//notifying registered streams of changed versions, and logging and returning any errors encountered
func (s *NotifyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	groups, err := s.files.CheckDefinition(s.config.DefinitionPath)
	if err != nil {
		log.Println("Error reloading definition:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = s.Notify(groups)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
