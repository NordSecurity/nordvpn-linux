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
	labelUnidentified    = "Unidentified"
	labelRecommended     = "Recommended"
)

type VPNConnection struct {
	Connection   Model  `json:"model"` // Represents the connection
	DisplayLabel string `json:"label"` // Human-friendly name for UI
}

func makeDisplayLabel(conn Model) string {
	switch conn.ConnectionType {
	case config.ServerSelectionRuleRecommended:
		return labelRecommended

	case config.ServerSelectionRuleCity:
		if conn.Country != "" && conn.City != "" {
			return fmt.Sprintf("%s, %s", conn.Country, conn.City)
		}
		return labelUnidentified

	case config.ServerSelectionRuleCountry:
		return conn.Country

	case config.ServerSelectionRuleSpecificServer:
		return conn.SpecificServerName

	case config.ServerSelectionRuleGroup:
		return conn.Group

	case config.ServerSelectionRuleCountryWithGroup:
		if conn.Group != "" && conn.Country != "" {
			return fmt.Sprintf("%s (%s)", conn.Group, conn.Country)
		}
		return labelUnidentified

	case config.ServerSelectionRuleSpecificServerWithGroup:
		if conn.Group != "" {
			if conn.Country != "" && conn.City != "" {
				return fmt.Sprintf("%s (%s, %s)", conn.Group, conn.Country, conn.City)
			} else if conn.Country != "" {
				return fmt.Sprintf("%s (%s)", conn.Group, conn.Country)
			}
		}
		return labelUnidentified

	default:
		return labelUnidentified
	}
}

// NewVPNConnection creates new a VPN connection
func NewVPNConnection(conn Model) VPNConnection {
	return VPNConnection{
		Connection:   conn,
		DisplayLabel: makeDisplayLabel(conn),
	}
}

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
func (r *RecentConnectionsStore) Get() ([]VPNConnection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkExistence(); err != nil {
		return nil, err
	}

	return r.loadLocked()
}

func (r *RecentConnectionsStore) find(conn VPNConnection, list []VPNConnection) int {
	return slices.IndexFunc(list, func(c VPNConnection) bool {
		// Compare by display label to handle cases where different servers
		// lead to the same destination (city, country, group, etc.)
		// This ensures we don't have duplicate entries with identical labels
		return c.DisplayLabel == conn.DisplayLabel
	})
}

// Add adds a new VPN connection to store if it does not exist yet
// New connections are placed at the beginning of the store
func (r *RecentConnectionsStore) Add(conn VPNConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log.Println("RECENTS:", conn)

	if err := r.checkExistence(); err != nil {
		return err
	}

	connections, err := r.loadLocked()
	if err != nil {
		return fmt.Errorf("adding new recent vpn connection: %w", err)
	}

	index := r.find(conn, connections)
	if index != -1 {
		connections = slices.Delete(connections, index, index+1)
	}

	connections = slices.Insert(connections, 0, conn)
	if len(connections) > maxRecentConnections {
		connections = connections[:maxRecentConnections]
	}

	return r.saveLocked(connections)
}

// Clean removes all stored connection information
func (r *RecentConnectionsStore) Clean() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.saveLocked([]VPNConnection{})
}

func (r *RecentConnectionsStore) saveLocked(values []VPNConnection) error {
	data, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("marshaling vpn connections store: %w", err)
	}

	if err := r.fsHandle.WriteFile(r.path, data, internal.PermUserRW); err != nil {
		return fmt.Errorf("writing vpn connections store: %w", err)
	}
	return nil
}

func (r *RecentConnectionsStore) loadLocked() ([]VPNConnection, error) {
	data, err := r.fsHandle.ReadFile(r.path)
	if err != nil {
		return nil, fmt.Errorf("reading recent connections store: %w", err)
	}

	var connections []VPNConnection
	if err := json.Unmarshal(data, &connections); err != nil {
		return nil, fmt.Errorf("unmarshaling vpn connections store: %w", err)
	}

	return connections, nil
}

func (r *RecentConnectionsStore) checkExistence() error {
	if !r.fsHandle.FileExists(r.path) {
		if err := r.saveLocked([]VPNConnection{}); err != nil {
			return fmt.Errorf("creating new recent vpn connections store: %w", err)
		}
	}
	return nil
}
