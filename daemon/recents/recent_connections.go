package recents

import (
	"encoding/json"
	"errors"
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

	logTag = "[recents]"
)

type RecentConnectionsStore struct {
	path              string
	fsHandle          config.FilesystemHandle
	mu                sync.Mutex
	pendingConnection Model
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
		return nil, fmt.Errorf("getting recent connections: %w", err)
	}

	conns, err := r.load()
	if err != nil {
		log.Printf("%s %s Getting recent VPN connections: %s\n", logTag, internal.WarningPrefix, err)
		if saveErr := r.save([]Model{}); saveErr != nil {
			return nil, errors.Join(
				fmt.Errorf("getting recent vpn connections: %w", err),
				fmt.Errorf("recreating recent connections file: %w", saveErr))
		}
		return []Model{}, nil
	}

	return conns, nil
}

// SearchOptions defines which fields to exclude from the search comparison.
// By default (all false), all fields are compared. Set a field to true to exclude it from matching.
type SearchOptions struct {
	ExcludeCountry            bool
	ExcludeCity               bool
	ExcludeGroup              bool
	ExcludeCountryCode        bool
	ExcludeSpecificServerName bool
	ExcludeSpecificServer     bool
	ExcludeConnectionType     bool
	ExcludeServerTechnologies bool
	ExcludeIsVirtual          bool
}

// searchOptionsForConnectionType returns SearchOptions based on the connection type.
func searchOptionsForConnectionType(connType config.ServerSelectionRule) SearchOptions {
	opts := SearchOptions{}

	switch connType {
	case config.ServerSelectionRule_SPECIFIC_SERVER,
		config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		// No exclusions
	case config.ServerSelectionRule_COUNTRY,
		config.ServerSelectionRule_COUNTRY_WITH_GROUP,
		config.ServerSelectionRule_GROUP,
		config.ServerSelectionRule_NONE,
		config.ServerSelectionRule_RECOMMENDED,
		config.ServerSelectionRule_CITY:
		// For non-specific server connections, exclude specific server fields
		opts.ExcludeSpecificServer = true
		opts.ExcludeSpecificServerName = true
	}

	return opts
}

// find searches for a model in the list using configurable field matching.
// Fields marked as true in SearchOptions will be excluded from comparison.
// Returns the index of the first matching model, or -1 if not found.
func (r *RecentConnectionsStore) find(model Model, list []Model, opts SearchOptions) int {
	return slices.IndexFunc(list, func(m Model) bool {
		if !opts.ExcludeCountry && m.Country != model.Country {
			return false
		}
		if !opts.ExcludeCity && m.City != model.City {
			return false
		}
		if !opts.ExcludeGroup && m.Group != model.Group {
			return false
		}
		if !opts.ExcludeCountryCode && m.CountryCode != model.CountryCode {
			return false
		}
		if !opts.ExcludeSpecificServerName && m.SpecificServerName != model.SpecificServerName {
			return false
		}
		if !opts.ExcludeSpecificServer && m.SpecificServer != model.SpecificServer {
			return false
		}
		if !opts.ExcludeConnectionType && m.ConnectionType != model.ConnectionType {
			return false
		}
		if !opts.ExcludeServerTechnologies && !slices.Equal(m.ServerTechnologies, model.ServerTechnologies) {
			return false
		}
		if !opts.ExcludeIsVirtual && m.IsVirtual != model.IsVirtual {
			return false
		}

		return true
	})
}

// Add adds a new VPN connection to store if it does not exist yet
// New connections are placed at the beginning of the store
func (r *RecentConnectionsStore) Add(model Model) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkExistence(); err != nil {
		return fmt.Errorf("adding new vpn connection: %w", err)
	}

	connections, err := r.load()
	if err != nil {
		log.Printf("%s %s Adding new recent VPN connection: %s\n", logTag, internal.WarningPrefix, err)
		if saveErr := r.save([]Model{}); saveErr != nil {
			return errors.Join(
				fmt.Errorf("adding new recent vpn connection: %w", err),
				fmt.Errorf("recreating recent connections file: %w", saveErr))
		}
		connections = []Model{}
	}

	// Sort server technologies, so that the order does not affect equality checks
	slices.Sort(model.ServerTechnologies)
	opts := searchOptionsForConnectionType(model.ConnectionType)
	index := r.find(model, connections, opts)
	if index != -1 {
		connections = slices.Delete(connections, index, index+1)
	}

	connections = slices.Insert(connections, 0, model)
	if len(connections) > maxRecentConnections {
		connections = connections[:maxRecentConnections]
	}

	if err := r.save(connections); err != nil {
		return fmt.Errorf("adding new recent vpn connection: %w", err)
	}

	return nil
}

// AddPending stores a recent connection model as pending to be added later
func (r *RecentConnectionsStore) AddPending(model Model) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pendingConnection = model
}

// PopPending retrieves and clears the pending recent connection model
func (r *RecentConnectionsStore) PopPending() (bool, Model) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.pendingConnection.IsEmpty() {
		return false, Model{}
	}

	connection := r.pendingConnection.Clone()
	r.pendingConnection = Model{}

	return true, connection
}

// Clean removes all stored connection information
func (r *RecentConnectionsStore) Clean() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.save([]Model{}); err != nil {
		return fmt.Errorf("cleaning existing recent vpn connections: %w", err)
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
