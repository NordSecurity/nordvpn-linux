package tray

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
)

const (
	recentConnectionsTimeout = 2 * time.Second
	maxRecentConnections     = 3
)

// RecentConnection represents a recent VPN connection
type RecentConnection struct {
	Country            string
	City               string
	Group              config.ServerGroup
	CountryCode        string
	SpecificServerName string
	SpecificServer     string
	ConnectionType     config.ServerSelectionRule
}

var groupTitles = map[config.ServerGroup]string{
	config.ServerGroup_DOUBLE_VPN:                       "Double VPN",
	config.ServerGroup_ONION_OVER_VPN:                   "Onion Over VPN",
	config.ServerGroup_STANDARD_VPN_SERVERS:             "Standard VPN Servers",
	config.ServerGroup_P2P:                              "P2P",
	config.ServerGroup_OBFUSCATED:                       "Obfuscated",
	config.ServerGroup_DEDICATED_IP:                     "Dedicated IP",
	config.ServerGroup_ULTRA_FAST_TV:                    "Ultra Fast TV",
	config.ServerGroup_ANTI_DDOS:                        "Anti DDOS",
	config.ServerGroup_NETFLIX_USA:                      "Netflix USA",
	config.ServerGroup_EUROPE:                           "Europe",
	config.ServerGroup_THE_AMERICAS:                     "The Americas",
	config.ServerGroup_ASIA_PACIFIC:                     "Asia Pacific",
	config.ServerGroup_AFRICA_THE_MIDDLE_EAST_AND_INDIA: "Africa The Middle East and India",
}

func formatGroupTitle(group config.ServerGroup) string {
	value, ok := groupTitles[group]
	if !ok {
		return ""
	}
	return value
}

func makeDisplayLabel(conn *RecentConnection) string {
	switch conn.ConnectionType {
	case config.ServerSelectionRule_CITY:
		if conn.Country != "" && conn.City != "" {
			return fmt.Sprintf("%s, %s", conn.Country, conn.City)
		}
		return ""

	case config.ServerSelectionRule_COUNTRY:
		return conn.Country

	case config.ServerSelectionRule_SPECIFIC_SERVER:
		return conn.SpecificServerName

	case config.ServerSelectionRule_GROUP:
		return formatGroupTitle(conn.Group)

	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		group := formatGroupTitle(conn.Group)
		if group == "" || conn.Country == "" {
			return ""
		}
		return fmt.Sprintf("%s (%s)", group, conn.Country)

	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		if conn.Group != config.ServerGroup_UNDEFINED {
			group := formatGroupTitle(conn.Group)
			if conn.Country != "" && conn.City != "" {
				return fmt.Sprintf("%s (%s, %s)", group, conn.Country, conn.City)
			} else if conn.Country != "" {
				return fmt.Sprintf("%s (%s)", group, conn.Country)
			}
		}
		return ""

	case config.ServerSelectionRule_NONE:
		log.Println(internal.WarningPrefix, "cannot make a proper label with server selection rule 'none'")
		fallthrough
	case config.ServerSelectionRule_RECOMMENDED:
		return ""
	}

	return ""
}

func connectByConnectionModel(ti *Instance, model *RecentConnection) bool {
	if model == nil {
		return false
	}

	normalizeForAPI := func(text string) string {
		return strings.ReplaceAll(text, " ", "_")
	}

	switch model.ConnectionType {
	case config.ServerSelectionRule_RECOMMENDED:
		return ti.connect("", "")

	case config.ServerSelectionRule_CITY:
		if model.City != "" {
			cityString := normalizeForAPI(model.City)
			return ti.connect(cityString, "")
		}

	case config.ServerSelectionRule_COUNTRY:
		if model.CountryCode != "" {
			return ti.connect(model.CountryCode, "")
		}

	case config.ServerSelectionRule_SPECIFIC_SERVER:
		if model.SpecificServer != "" {
			return ti.connect(model.SpecificServer, "")
		}

	case config.ServerSelectionRule_GROUP:
		if model.Group != config.ServerGroup_UNDEFINED {
			group := normalizeForAPI(model.Group.String())
			return ti.connect("", group)
		}

	case config.ServerSelectionRule_COUNTRY_WITH_GROUP:
		if model.CountryCode != "" && model.Group != config.ServerGroup_UNDEFINED {
			group := normalizeForAPI(model.Group.String())
			return ti.connect(model.CountryCode, group)
		}

	case config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP:
		if model.SpecificServer != "" && model.Group != config.ServerGroup_UNDEFINED {
			group := normalizeForAPI(model.Group.String())
			return ti.connect(model.SpecificServer, group)
		}

	case config.ServerSelectionRule_NONE:
	}
	return false
}

type recentConnectionsManager struct {
	mu          sync.RWMutex
	connections []RecentConnection
	client      pb.DaemonClient
}

// newRecentConnectionsManager creates a new recent VPN connection manager
func newRecentConnectionsManager(client pb.DaemonClient) *recentConnectionsManager {
	return &recentConnectionsManager{
		connections: make([]RecentConnection, 0),
		client:      client,
	}
}

// UpdateRecentConnections updates local list of recent VPN connections
func (m *recentConnectionsManager) UpdateRecentConnections() error {
	ctx, cancel := context.WithTimeout(context.Background(), recentConnectionsTimeout)
	defer cancel()

	limit := int64(maxRecentConnections)
	resp, err := m.client.GetRecentConnections(
		ctx,
		&pb.RecentConnectionsRequest{Limit: &limit},
		grpc.WaitForReady(true),
	)

	if err != nil || resp == nil {
		return err
	}

	// Convert gRPC models to tray models
	connections := make([]RecentConnection, 0, len(resp.Connections))
	for _, conn := range resp.Connections {
		connections = append(connections, RecentConnection{
			Country:            conn.Country,
			City:               conn.City,
			Group:              conn.Group,
			CountryCode:        conn.CountryCode,
			SpecificServerName: conn.SpecificServerName,
			SpecificServer:     conn.SpecificServer,
			ConnectionType:     conn.ConnectionType,
		})
	}

	m.mu.Lock()
	m.connections = connections
	m.mu.Unlock()
	return nil
}

// GetRecentConnections returns recent VPN connections
func (m *recentConnectionsManager) GetRecentConnections() []RecentConnection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return slices.Clone(m.connections)
}
