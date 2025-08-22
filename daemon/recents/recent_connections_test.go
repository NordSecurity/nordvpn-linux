package recents

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockconfig "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVPNConnection_Recommended(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		ConnectionType: config.ServerSelectionRuleRecommended,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "Recommended", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestNewVPNConnection_City(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		Country:        "Germany",
		City:           "Berlin",
		ConnectionType: config.ServerSelectionRuleCity,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "Germany, Berlin", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestNewVPNConnection_Country(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRuleCountry,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "France", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestNewVPNConnection_SpecificServer(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		SpecificServerName: "us1234",
		ConnectionType:     config.ServerSelectionRuleSpecificServer,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "us1234", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestNewVPNConnection_Group(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		Group:          "P2P",
		ConnectionType: config.ServerSelectionRuleGroup,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "P2P", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestNewVPNConnection_CountryWithGroup(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		Country:        "United States",
		Group:          "Double VPN",
		ConnectionType: config.ServerSelectionRuleCountryWithGroup,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "Double VPN (United States)", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestNewVPNConnection_SpecificServerWithGroup(t *testing.T) {
	category.Set(t, category.Unit)

	conn := Model{
		Country:        "Canada",
		City:           "Toronto",
		Group:          "Dedicated IP",
		ConnectionType: config.ServerSelectionRuleSpecificServerWithGroup,
	}

	result := NewVPNConnection(conn)

	assert.Equal(t, "Dedicated IP (Canada, Toronto)", result.DisplayLabel)
	assert.Equal(t, conn, result.Connection)
}

func TestRecentConnectionsStore_Get_EmptyStore(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	require.NoError(t, err)
	assert.Empty(t, connections)
	assert.True(t, fs.FileExists("/test/path"))
}

func TestRecentConnectionsStore_Get_ExistingConnections(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	existingConnections := []VPNConnection{
		{
			Connection:   Model{Country: "Germany", ConnectionType: config.ServerSelectionRuleCountry},
			DisplayLabel: "Germany",
		},
		{
			Connection:   Model{ConnectionType: config.ServerSelectionRuleRecommended},
			DisplayLabel: "Recommended",
		},
	}
	data, _ := json.Marshal(existingConnections)
	fs.AddFile("/test/path", data)

	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	require.NoError(t, err)
	assert.Equal(t, existingConnections, connections)
}

func TestRecentConnectionsStore_Add_NewConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	newConn := NewVPNConnection(Model{
		Country:        "Spain",
		ConnectionType: config.ServerSelectionRuleCountry,
	})

	err := store.Add(newConn)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, newConn, connections[0])
}

func TestRecentConnectionsStore_Add_DuplicateConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn1 := NewVPNConnection(Model{
		Country:        "Italy",
		ConnectionType: config.ServerSelectionRuleCountry,
	})
	conn2 := NewVPNConnection(Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRuleCountry,
	})

	// Add first two connections
	err := store.Add(conn1)
	require.NoError(t, err)
	err = store.Add(conn2)
	require.NoError(t, err)

	// Add duplicate of first connection
	err = store.Add(conn1)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 2)
	assert.Equal(t, conn1, connections[0]) // Should be at the front
	assert.Equal(t, conn2, connections[1])
}

func TestRecentConnectionsStore_Add_RecommendedDuplicate(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	recommended := NewVPNConnection(Model{
		ConnectionType: config.ServerSelectionRuleRecommended,
	})

	// Add recommended connection multiple times
	for i := 0; i < 3; i++ {
		err := store.Add(recommended)
		require.NoError(t, err)
	}

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1) // Should only have one entry
	assert.Equal(t, recommended, connections[0])
}

func TestRecentConnectionsStore_Add_CapacityLimit(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Add more than MaxRecentConnections
	for i := 0; i < maxRecentConnections+5; i++ {
		conn := NewVPNConnection(Model{
			Country:        string(rune('A' + i)),
			ConnectionType: config.ServerSelectionRuleCountry,
		})
		err := store.Add(conn)
		require.NoError(t, err)
	}

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, maxRecentConnections)

	// Verify the most recent connections are kept
	assert.Equal(t, string(rune('A'+maxRecentConnections+4)), connections[0].Connection.Country)
}

func TestRecentConnectionsStore_Clean(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Add some connections
	conn := NewVPNConnection(Model{
		Country:        "Norway",
		ConnectionType: config.ServerSelectionRuleCountry,
	})
	err := store.Add(conn)
	require.NoError(t, err)

	// Clean the store
	err = store.Clean()
	require.NoError(t, err)

	// Verify store is empty
	connections, err := store.Get()
	require.NoError(t, err)
	assert.Empty(t, connections)
}

func TestRecentConnectionsStore_Get_ReadError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.AddFile("/test/path", []byte("invalid json"))
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	assert.Error(t, err)
	assert.Nil(t, connections)
}

func TestRecentConnectionsStore_Add_WriteError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.WriteErr = errors.New("write error")
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn := NewVPNConnection(Model{
		Country:        "Sweden",
		ConnectionType: config.ServerSelectionRuleCountry,
	})

	err := store.Add(conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}

func TestRecentConnectionsStore_ConcurrentAccess(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Run multiple goroutines adding connections
	done := make(chan bool, 3)

	go func() {
		conn := NewVPNConnection(Model{
			Country:        "Denmark",
			ConnectionType: config.ServerSelectionRuleCountry,
		})
		_ = store.Add(conn)
		done <- true
	}()

	go func() {
		conn := NewVPNConnection(Model{
			Country:        "Finland",
			ConnectionType: config.ServerSelectionRuleCountry,
		})
		_ = store.Add(conn)
		done <- true
	}()

	go func() {
		_, _ = store.Get()
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify the store is in a consistent state
	connections, err := store.Get()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(connections), 0)
	assert.LessOrEqual(t, len(connections), 2)
}

func TestMakeDisplayLabel_EmptyFields(t *testing.T) {
	category.Set(t, category.Unit)

	testCases := []struct {
		name     string
		conn     Model
		expected string
	}{
		{
			name: "City without country",
			conn: Model{
				City:           "Berlin",
				ConnectionType: config.ServerSelectionRuleCity,
			},
			expected: labelUnidentified,
		},
		{
			name: "Country with group - missing country",
			conn: Model{
				Group:          "P2P",
				ConnectionType: config.ServerSelectionRuleCountryWithGroup,
			},
			expected: labelUnidentified,
		},
		{
			name: "Country with group - missing group",
			conn: Model{
				Country:        "USA",
				ConnectionType: config.ServerSelectionRuleCountryWithGroup,
			},
			expected: labelUnidentified,
		},
		{
			name: "Unknown connection type",
			conn: Model{
				Country:        "Unknown",
				ConnectionType: config.ServerSelectionRule(999),
			},
			expected: labelUnidentified,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := makeDisplayLabel(tc.conn)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRecentConnectionsStore_LoadSaveCycle(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Create test connections
	connections := []VPNConnection{
		NewVPNConnection(Model{
			Country:        "Netherlands",
			City:           "Amsterdam",
			ConnectionType: config.ServerSelectionRuleCity,
		}),
		NewVPNConnection(Model{
			ConnectionType: config.ServerSelectionRuleRecommended,
		}),
		NewVPNConnection(Model{
			SpecificServerName: "uk1234",
			ConnectionType:     config.ServerSelectionRuleSpecificServer,
		}),
	}

	// Add all connections
	for _, conn := range connections {
		err := store.Add(conn)
		require.NoError(t, err)
	}

	// Create a new store instance with the same filesystem
	store2 := NewRecentConnectionsStore("/test/path", &fs)

	// Load connections from the new store
	loadedConnections, err := store2.Get()
	require.NoError(t, err)

	// Verify all connections are preserved in reverse order (most recent first)
	require.Len(t, loadedConnections, len(connections))
	for i := range connections {
		assert.Equal(t, connections[len(connections)-1-i], loadedConnections[i])
	}
}

func TestRecentConnectionsStore_Add_DuplicatesByLabel(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Test case 1: Different servers for the same city should be treated as duplicates
	cityConn1 := NewVPNConnection(Model{
		Country:        "Germany",
		City:           "Berlin",
		SpecificServer: "de123",
		ConnectionType: config.ServerSelectionRuleCity,
	})
	cityConn2 := NewVPNConnection(Model{
		Country:        "Germany",
		City:           "Berlin",
		SpecificServer: "de456", // Different server
		ConnectionType: config.ServerSelectionRuleCity,
	})

	err := store.Add(cityConn1)
	require.NoError(t, err)
	err = store.Add(cityConn2)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1, "Should only have one Berlin connection")
	assert.Equal(t, "Germany, Berlin", connections[0].DisplayLabel)
	assert.Equal(t, cityConn2.Connection, connections[0].Connection, "Should keep the most recent connection")

	// Test case 2: Different servers for the same country should be treated as duplicates
	countryConn1 := NewVPNConnection(Model{
		Country:        "France",
		SpecificServer: "fr789",
		ConnectionType: config.ServerSelectionRuleCountry,
	})
	countryConn2 := NewVPNConnection(Model{
		Country:        "France",
		SpecificServer: "fr012", // Different server
		ConnectionType: config.ServerSelectionRuleCountry,
	})

	err = store.Add(countryConn1)
	require.NoError(t, err)
	err = store.Add(countryConn2)
	require.NoError(t, err)

	connections, err = store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 2, "Should have France and Berlin")
	assert.Equal(t, "France", connections[0].DisplayLabel)
	assert.Equal(t, countryConn2.Connection, connections[0].Connection, "Should keep the most recent France connection")

	// Test case 3: Multiple recommended connections should be treated as one
	recommended1 := NewVPNConnection(Model{
		ConnectionType: config.ServerSelectionRuleRecommended,
		SpecificServer: "us111",
	})
	recommended2 := NewVPNConnection(Model{
		ConnectionType: config.ServerSelectionRuleRecommended,
		SpecificServer: "ca222", // Different server
	})

	err = store.Add(recommended1)
	require.NoError(t, err)
	err = store.Add(recommended2)
	require.NoError(t, err)

	connections, err = store.Get()
	require.NoError(t, err)
	// Should still have 3 unique entries: Recommended, France, Berlin
	require.Len(t, connections, 3, "Should have Recommended, France, and Berlin")
	assert.Equal(t, "Recommended", connections[0].DisplayLabel)

	// Test case 4: Same group but different servers should be treated as duplicates
	groupConn1 := NewVPNConnection(Model{
		Group:          "P2P",
		SpecificServer: "nl333",
		ConnectionType: config.ServerSelectionRuleGroup,
	})
	groupConn2 := NewVPNConnection(Model{
		Group:          "P2P",
		SpecificServer: "se444", // Different server
		ConnectionType: config.ServerSelectionRuleGroup,
	})

	err = store.Add(groupConn1)
	require.NoError(t, err)
	err = store.Add(groupConn2)
	require.NoError(t, err)

	connections, err = store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 4, "Should have P2P, Recommended, France, and Berlin")
	assert.Equal(t, "P2P", connections[0].DisplayLabel)
	assert.Equal(t, groupConn2.Connection, connections[0].Connection, "Should keep the most recent P2P connection")
}

func TestRecentConnectionsStore_Add_DifferentLabelsNotDuplicates(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Add connections with different labels
	connections := []VPNConnection{
		NewVPNConnection(Model{
			Country:        "Germany",
			City:           "Berlin",
			ConnectionType: config.ServerSelectionRuleCity,
		}),
		NewVPNConnection(Model{
			Country:        "Germany",
			City:           "Munich",
			ConnectionType: config.ServerSelectionRuleCity,
		}),
		NewVPNConnection(Model{
			Country:        "Germany",
			ConnectionType: config.ServerSelectionRuleCountry,
		}),
		NewVPNConnection(Model{
			SpecificServerName: "de123",
			ConnectionType:     config.ServerSelectionRuleSpecificServer,
		}),
		NewVPNConnection(Model{
			Group:          "P2P",
			Country:        "Germany",
			ConnectionType: config.ServerSelectionRuleCountryWithGroup,
		}),
	}

	// Add all connections
	for _, conn := range connections {
		err := store.Add(conn)
		require.NoError(t, err)
	}

	// Get all connections
	storedConnections, err := store.Get()
	require.NoError(t, err)

	// All connections should be present as they have different labels
	require.Len(t, storedConnections, len(connections), "All connections should be stored as they have different labels")

	// Verify the labels are all different
	labels := make(map[string]bool)
	for _, conn := range storedConnections {
		assert.False(t, labels[conn.DisplayLabel], "Should not have duplicate labels")
		labels[conn.DisplayLabel] = true
	}
}
