package recents

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// maxRecentConnections defines the maximum number of recent connections to store
	maxRecentConnections = 10
	logTag               = "[recents]"
)

type RecentConnectionsStore struct {
	path     string
	fsHandle config.FilesystemHandle
	mu       sync.Mutex
}

// NewRecentConnectionsStore creates a recent VPN connection store
func NewRecentConnectionsStore(
	path string,
	fsHandle config.FilesystemHandle,
) *RecentConnectionsStore {
	return &RecentConnectionsStore{
		path:     path,
		fsHandle: fsHandle,
	}
}

// Get retrieves all stored VPN connections from the store
func (r *RecentConnectionsStore) Get() ([]Model, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkExistence(); err != nil {
		return nil, fmt.Errorf("%s %s getting recent connections: %w\n", logTag, internal.ErrorPrefix, err)
	}

	conns, err := r.load()
	if err != nil {
		return nil, fmt.Errorf("%s %s getting recent vpn connections: %w", logTag, internal.ErrorPrefix, err)
	}

	return conns, nil
}

func (r *RecentConnectionsStore) find(model Model, list []Model) int {
	return slices.IndexFunc(list, func(m Model) bool {
		return m.Country == model.Country &&
			m.City == model.City &&
			m.Group == model.Group &&
			m.CountryCode == model.CountryCode &&
			m.SpecificServerName == model.SpecificServerName &&
			m.SpecificServer == model.SpecificServer &&
			m.ConnectionType == model.ConnectionType
	})
}

// Add adds a new VPN connection to store if it does not exist yet
// New connections are placed at the beginning of the store
func (r *RecentConnectionsStore) Add(model Model) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkExistence(); err != nil {
		return fmt.Errorf("%s %s adding new vpn connection: %w\n", logTag, internal.ErrorPrefix, err)
	}

	connections, err := r.load()
	if err != nil {
		return fmt.Errorf("%s %s adding new recent vpn connection: %w\n", logTag, internal.ErrorPrefix, err)
	}

	index := r.find(model, connections)
	if index != -1 {
		connections = slices.Delete(connections, index, index+1)
	}

	connections = slices.Insert(connections, 0, model)
	if len(connections) > maxRecentConnections {
		connections = connections[:maxRecentConnections]
	}

	if err := r.save(connections); err != nil {
		return fmt.Errorf("%s %s adding new recent vpn connection: %w\n", logTag, internal.ErrorPrefix, err)
	}

	return nil
}

// Clean removes all stored connection information
func (r *RecentConnectionsStore) Clean() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.save([]Model{}); err != nil {
		return fmt.Errorf("%s %s cleaning existing recent vpn connections: %w\n", logTag, internal.ErrorPrefix, err)
	}

	return nil
}

func (r *RecentConnectionsStore) save(values []Model) error {
	data, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("marshaling vpn connections store: %w", err)
	}

	if err := r.fsHandle.WriteFile(r.path, data, internal.PermUserRW); err != nil {
		return fmt.Errorf("writing vpn connections store: %w", err)
	}
	defer log.Println("saved recent conns:", values)
	return nil
}

func (r *RecentConnectionsStore) load() ([]Model, error) {
	data, err := r.fsHandle.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("reading recent connections store: %w", err)
	}

	var connections []Model
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, fmt.Errorf("unmarshaling vpn connections store: %w", err)
	}

	return connections, nil
}

func (r *RecentConnectionsStore) checkExistence() error {
	if !r.fsHandle.FileExists(r.path) {
		if err := r.save([]Model{}); err != nil {
			return fmt.Errorf("creating new recent vpn connections store: %w", err)
		}
	}
	return nil
}
