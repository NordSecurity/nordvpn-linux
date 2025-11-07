package recents

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockconfig "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	existingConnections := []Model{
		{
			Country:        "Germany",
			ConnectionType: config.ServerSelectionRule_COUNTRY,
		},
		{
			ConnectionType: config.ServerSelectionRule_RECOMMENDED,
		},
	}
	data, _ := json.Marshal(existingConnections)
	fs.AddFile("/test/path", data)

	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	require.NoError(t, err)
	assert.Equal(t, existingConnections, connections)
}

func TestRecentConnectionsStore_Get_InvalidJSON(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.AddFile("/test/path", []byte("invalid json"))
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	assert.NoError(t, err)
	assert.NotNil(t, connections)
	assert.Len(t, connections, 0)
}

func TestRecentConnectionsStore_Add_SingleConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	newConn := Model{
		Country:        "Spain",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err := store.Add(newConn)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, newConn, connections[0])
}

func TestRecentConnectionsStore_Add_MovesExistingToFront(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn1 := Model{
		Country:        "Italy",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	conn2 := Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err := store.Add(conn1)
	require.NoError(t, err)
	err = store.Add(conn2)
	require.NoError(t, err)

	err = store.Add(conn1)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 2)
	assert.Equal(t, conn1, connections[0])
	assert.Equal(t, conn2, connections[1])
}

func TestRecentConnectionsStore_Add_RespectsCapacityLimit(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	for i := 0; i < maxRecentConnections+5; i++ {
		conn := Model{
			Country:        string(rune('A' + i)),
			ConnectionType: config.ServerSelectionRule_COUNTRY,
		}
		err := store.Add(conn)
		require.NoError(t, err)
	}

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, maxRecentConnections)
	assert.Equal(t, string(rune('A'+maxRecentConnections+4)), connections[0].Country)
}

func TestRecentConnectionsStore_Add_WriteError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.WriteErr = errors.New("write error")
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn := Model{
		Country:        "Sweden",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err := store.Add(conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "write error")
}

func TestRecentConnectionsStore_Clean(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn := Model{
		Country:        "Norway",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	err := store.Add(conn)
	require.NoError(t, err)

	err = store.Clean()
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Empty(t, connections)
}

func TestRecentConnectionsStore_Find_ExactMatch(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	models := []Model{
		{Country: "USA", City: "New York", ConnectionType: config.ServerSelectionRule_CITY},
		{Country: "Germany", ConnectionType: config.ServerSelectionRule_COUNTRY},
		{ConnectionType: config.ServerSelectionRule_RECOMMENDED},
		{SpecificServerName: "uk1234", ConnectionType: config.ServerSelectionRule_SPECIFIC_SERVER},
	}

	assert.Equal(t, 0, store.find(models[0], models, searchOptionsForConnectionType(models[0].ConnectionType)))
	assert.Equal(t, 1, store.find(models[1], models, searchOptionsForConnectionType(models[1].ConnectionType)))
	assert.Equal(t, 2, store.find(models[2], models, searchOptionsForConnectionType(models[2].ConnectionType)))
	assert.Equal(t, 3, store.find(models[3], models, searchOptionsForConnectionType(models[3].ConnectionType)))

	nonExisting := Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	assert.Equal(t, -1, store.find(nonExisting, models, searchOptionsForConnectionType(nonExisting.ConnectionType)))
}

func TestRecentConnectionsStore_Find_DifferentServersAreDifferent(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// For SPECIFIC_SERVER connection type, different servers should be different
	specificConn1 := Model{
		Country:            "Germany",
		City:               "Berlin",
		SpecificServerName: "de123",
		SpecificServer:     "de123",
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
	}
	specificConn2 := Model{
		Country:            "Germany",
		City:               "Berlin",
		SpecificServerName: "de456",
		SpecificServer:     "de456",
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
	}

	err := store.Add(specificConn1)
	require.NoError(t, err)
	err = store.Add(specificConn2)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, 2, "Different specific servers should create separate entries")
}

func TestRecentConnectionsStore_Find_AllFieldsMustMatch(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Use SPECIFIC_SERVER connection type so all fields are compared
	base := Model{
		Country:            "USA",
		City:               "New York",
		Group:              config.ServerGroup_P2P,
		CountryCode:        "US",
		SpecificServerName: "US #1234",
		SpecificServer:     "us1234",
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
		IsVirtual:          false,
	}

	variations := []Model{
		{Country: "Canada", City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: "Los Angeles", Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: base.City, Group: config.ServerGroup_DOUBLE_VPN, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: "CA", SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: "US #5678", SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: "us5678", ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: config.ServerSelectionRule_COUNTRY, IsVirtual: base.IsVirtual},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: !base.IsVirtual},
	}

	err := store.Add(base)
	require.NoError(t, err)

	for _, variation := range variations {
		err = store.Add(variation)
		require.NoError(t, err)
	}

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, len(variations)+1, "All variations should create separate entries when using SPECIFIC_SERVER type")
}

func TestRecentConnectionsStore_Persistence(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections := []Model{
		{
			Country:        "Netherlands",
			City:           "Amsterdam",
			ConnectionType: config.ServerSelectionRule_CITY,
		},
		{
			ConnectionType: config.ServerSelectionRule_RECOMMENDED,
		},
		{
			SpecificServerName: "uk1234",
			ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
		},
	}

	for _, conn := range connections {
		err := store.Add(conn)
		require.NoError(t, err)
	}

	store2 := NewRecentConnectionsStore("/test/path", &fs)

	loadedConnections, err := store2.Get()
	require.NoError(t, err)

	require.Len(t, loadedConnections, len(connections))
	for i := range connections {
		assert.Equal(t, connections[len(connections)-1-i], loadedConnections[i])
	}
}
func TestRecentConnectionsStore_ConcurrentAccess(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	initial := Model{
		Country:        "Sweden",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	require.NoError(t, store.Add(initial))

	denmark := Model{
		Country:        "Denmark",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	finland := Model{
		Country:        "Finland",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	norway := Model{
		Country:        "Norway",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := store.Add(denmark); err != nil {
				errChan <- fmt.Errorf("add Denmark: %w", err)
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := store.Add(finland); err != nil {
				errChan <- fmt.Errorf("add Finland: %w", err)
			}
		}()
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := store.Add(norway); err != nil {
				errChan <- fmt.Errorf("add Norway: %w", err)
			}
		}()
	}

	wg.Wait()

	err := store.Clean()
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		country := Model{
			Country:        fmt.Sprintf("Country%d", i),
			ConnectionType: config.ServerSelectionRule_COUNTRY,
		}
		err := store.Add(country)
		require.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			connections, err := store.Get()
			if err != nil {
				errChan <- fmt.Errorf("get connections: %w", err)
			}
			if len(connections) != 3 {
				errChan <- fmt.Errorf("expected 3 connections, got %d", len(connections))
			}
		}()
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		t.Fatalf("Concurrent operations failed with %d errors. First error: %v", len(errors), errors[0])
	}

	connections, err := store.Get()
	require.NoError(t, err)

	assert.Len(t, connections, 3, "Should have exactly 3 connections after Clean and subsequent Adds")

	require.Equal(t, "Country2", connections[0].Country)
	require.Equal(t, "Country1", connections[1].Country)
	require.Equal(t, "Country0", connections[2].Country)
}

func TestRecentConnectionsStore_RaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	var wg sync.WaitGroup
	const iterations = 100

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		go func(idx int) {
			defer wg.Done()
			model := Model{
				Country:        fmt.Sprintf("Country%d", idx%10),
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			}
			_ = store.Add(model)
		}(i)

		go func() {
			defer wg.Done()
			_, _ = store.Get()
		}()

		go func(idx int) {
			defer wg.Done()
			if idx%20 == 0 {
				_ = store.Clean()
			}
		}(i)
	}

	wg.Wait()

	_, err := store.Get()
	assert.NoError(t, err, "Store should still be functional after concurrent access")
}

func TestRecentConnectionsStore_ConcurrentAdd_OrderingGuarantee(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	denmark := Model{
		Country:        "Denmark",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := store.Add(denmark)
			assert.NoError(t, err)
		}()
	}

	wg.Wait()

	connections, err := store.Get()
	require.NoError(t, err)

	denmarkCount := 0
	for _, conn := range connections {
		if conn.Country == "Denmark" {
			denmarkCount++
			assert.Equal(t, denmark, conn)
		}
	}

	assert.Equal(t, 1, denmarkCount, "Denmark should appear exactly once despite concurrent adds")
}

func TestRecentConnectionsStore_CheckExistence_CreatesFile(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	assert.False(t, fs.FileExists("/test/path"))

	err := store.checkExistence()
	require.NoError(t, err)
	assert.True(t, fs.FileExists("/test/path"))

	data, err := fs.ReadFile("/test/path")
	require.NoError(t, err)
	assert.Equal(t, []byte("[]"), data)
}

func TestRecentConnectionsStore_Save_Error(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.WriteErr = errors.New("permission denied")
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections := []Model{
		{
			Country:        "Test",
			ConnectionType: config.ServerSelectionRule_COUNTRY,
		},
	}

	err := store.save(connections)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "writing vpn connections store")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRecentConnectionsStore_Load_Error(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.ReadErr = errors.New("file not found")
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading recent connections store")
	assert.Contains(t, err.Error(), "file not found")
	assert.Nil(t, connections)
}

func TestRecentConnectionsStore_SearchOptionsForConnectionType(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		connType     config.ServerSelectionRule
		wantExcluded SearchOptions
	}{
		{
			name:     "SPECIFIC_SERVER - no exclusions",
			connType: config.ServerSelectionRule_SPECIFIC_SERVER,
			wantExcluded: SearchOptions{
				ExcludeCountry:            false,
				ExcludeCity:               false,
				ExcludeGroup:              false,
				ExcludeCountryCode:        false,
				ExcludeSpecificServerName: false,
				ExcludeSpecificServer:     false,
				ExcludeConnectionType:     false,
				ExcludeServerTechnologies: false,
				ExcludeIsVirtual:          false,
			},
		},
		{
			name:     "SPECIFIC_SERVER_WITH_GROUP - no exclusions",
			connType: config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
			wantExcluded: SearchOptions{
				ExcludeCountry:            false,
				ExcludeCity:               false,
				ExcludeGroup:              false,
				ExcludeCountryCode:        false,
				ExcludeSpecificServerName: false,
				ExcludeSpecificServer:     false,
				ExcludeConnectionType:     false,
				ExcludeServerTechnologies: false,
				ExcludeIsVirtual:          false,
			},
		},
		{
			name:     "COUNTRY - excludes specific server fields",
			connType: config.ServerSelectionRule_COUNTRY,
			wantExcluded: SearchOptions{
				ExcludeCountry:            false,
				ExcludeCity:               false,
				ExcludeGroup:              false,
				ExcludeCountryCode:        false,
				ExcludeSpecificServerName: true,
				ExcludeSpecificServer:     true,
				ExcludeConnectionType:     false,
				ExcludeServerTechnologies: false,
				ExcludeIsVirtual:          false,
			},
		},
		{
			name:     "CITY - excludes specific server fields",
			connType: config.ServerSelectionRule_CITY,
			wantExcluded: SearchOptions{
				ExcludeCountry:            false,
				ExcludeCity:               false,
				ExcludeGroup:              false,
				ExcludeCountryCode:        false,
				ExcludeSpecificServerName: true,
				ExcludeSpecificServer:     true,
				ExcludeConnectionType:     false,
				ExcludeServerTechnologies: false,
				ExcludeIsVirtual:          false,
			},
		},
		{
			name:     "RECOMMENDED - excludes specific server fields",
			connType: config.ServerSelectionRule_RECOMMENDED,
			wantExcluded: SearchOptions{
				ExcludeCountry:            false,
				ExcludeCity:               false,
				ExcludeGroup:              false,
				ExcludeCountryCode:        false,
				ExcludeSpecificServerName: true,
				ExcludeSpecificServer:     true,
				ExcludeConnectionType:     false,
				ExcludeServerTechnologies: false,
				ExcludeIsVirtual:          false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchOptionsForConnectionType(tt.connType)
			assert.Equal(t, tt.wantExcluded, got)
		})
	}
}

func TestRecentConnectionsStore_Find_WithSearchOptions(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	base := Model{
		Country:            "USA",
		City:               "New York",
		Group:              config.ServerGroup_P2P,
		CountryCode:        "US",
		SpecificServerName: "US #1234",
		SpecificServer:     "us1234",
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER,
		IsVirtual:          false,
	}

	tests := []struct {
		name        string
		searchModel Model
		opts        SearchOptions
		wantFound   bool
	}{
		{
			name:        "ExcludeCountry - matches despite different country",
			searchModel: Model{Country: "Canada", City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeCountry: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeCity - matches despite different city",
			searchModel: Model{Country: base.Country, City: "Los Angeles", Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeCity: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeGroup - matches despite different group",
			searchModel: Model{Country: base.Country, City: base.City, Group: config.ServerGroup_DOUBLE_VPN, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeGroup: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeCountryCode - matches despite different country code",
			searchModel: Model{Country: base.Country, City: base.City, Group: base.Group, CountryCode: "CA", SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeCountryCode: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeSpecificServerName - matches despite different server name",
			searchModel: Model{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: "US #5678", SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeSpecificServerName: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeSpecificServer - matches despite different server",
			searchModel: Model{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: "us5678", ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeSpecificServer: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeConnectionType - matches despite different connection type",
			searchModel: Model{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: config.ServerSelectionRule_COUNTRY, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeConnectionType: true},
			wantFound:   true,
		},
		{
			name:        "ExcludeIsVirtual - matches despite different virtual status",
			searchModel: Model{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: !base.IsVirtual},
			opts:        SearchOptions{ExcludeIsVirtual: true},
			wantFound:   true,
		},
		{
			name:        "Multiple exclusions - matches with multiple differences",
			searchModel: Model{Country: "Canada", City: "Toronto", Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{ExcludeCountry: true, ExcludeCity: true},
			wantFound:   true,
		},
		{
			name:        "No exclusions - doesn't match with difference",
			searchModel: Model{Country: "Canada", City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType, IsVirtual: base.IsVirtual},
			opts:        SearchOptions{},
			wantFound:   false,
		},
	}

	list := []Model{base}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idx := store.find(tt.searchModel, list, tt.opts)
			if tt.wantFound {
				assert.Equal(t, 0, idx, "Expected to find model at index 0")
			} else {
				assert.Equal(t, -1, idx, "Expected not to find model")
			}
		})
	}
}

func TestRecentConnectionsStore_Find_ServerTechnologies(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	base := Model{
		Country:            "USA",
		ConnectionType:     config.ServerSelectionRule_COUNTRY,
		ServerTechnologies: []core.ServerTechnology{1, 2, 3},
	}

	tests := []struct {
		name               string
		searchTechnologies []core.ServerTechnology
		opts               SearchOptions
		wantFound          bool
	}{
		{
			name:               "Exact match",
			searchTechnologies: []core.ServerTechnology{1, 2, 3},
			opts:               SearchOptions{},
			wantFound:          true,
		},
		{
			name:               "Different technologies",
			searchTechnologies: []core.ServerTechnology{1, 2, 4},
			opts:               SearchOptions{},
			wantFound:          false,
		},
		{
			name:               "Different order - doesn't match",
			searchTechnologies: []core.ServerTechnology{3, 2, 1},
			opts:               SearchOptions{},
			wantFound:          false,
		},
		{
			name:               "ExcludeServerTechnologies - matches despite difference",
			searchTechnologies: []core.ServerTechnology{4, 5, 6},
			opts:               SearchOptions{ExcludeServerTechnologies: true},
			wantFound:          true,
		},
	}

	list := []Model{base}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchModel := Model{
				Country:            base.Country,
				ConnectionType:     base.ConnectionType,
				ServerTechnologies: tt.searchTechnologies,
			}
			idx := store.find(searchModel, list, tt.opts)
			if tt.wantFound {
				assert.Equal(t, 0, idx, "Expected to find model at index 0")
			} else {
				assert.Equal(t, -1, idx, "Expected not to find model")
			}
		})
	}
}

func TestRecentConnectionsStore_Add_ServerTechnologiesSorting(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn1 := Model{
		Country:            "Germany",
		ConnectionType:     config.ServerSelectionRule_COUNTRY,
		ServerTechnologies: []core.ServerTechnology{3, 1, 2},
	}

	err := store.Add(conn1)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)

	assert.Equal(t, []core.ServerTechnology{1, 2, 3}, connections[0].ServerTechnologies, "ServerTechnologies should be sorted")

	conn2 := Model{
		Country:            "Germany",
		ConnectionType:     config.ServerSelectionRule_COUNTRY,
		ServerTechnologies: []core.ServerTechnology{2, 3, 1},
	}

	err = store.Add(conn2)
	require.NoError(t, err)

	connections, err = store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1, "Should have only one connection as they match after sorting")
	assert.Equal(t, []core.ServerTechnology{1, 2, 3}, connections[0].ServerTechnologies)
}

func TestRecentConnectionsStore_Get_LoadErrorRecreatesFile(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.AddFile("/test/path", []byte("corrupted data"))
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	require.NoError(t, err, "Get should succeed after recreating file")
	assert.Empty(t, connections)

	data, err := fs.ReadFile("/test/path")
	require.NoError(t, err)
	assert.Equal(t, []byte("[]"), data, "File should be recreated with empty array")
}

func TestRecentConnectionsStore_Get_LoadErrorWithSaveError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.AddFile("/test/path", []byte("corrupted data"))
	store := NewRecentConnectionsStore("/test/path", &fs)

	fs.WriteErr = errors.New("write permission denied")

	connections, err := store.Get()

	assert.Error(t, err)
	assert.Nil(t, connections)
	assert.Contains(t, err.Error(), "getting recent vpn connections")
	assert.Contains(t, err.Error(), "recreating recent connections file")
}

func TestRecentConnectionsStore_Add_LoadErrorRecreatesFile(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.AddFile("/test/path", []byte("corrupted data"))
	store := NewRecentConnectionsStore("/test/path", &fs)

	newConn := Model{
		Country:        "Spain",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err := store.Add(newConn)
	require.NoError(t, err, "Add should succeed after recreating file")

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, newConn, connections[0])
}

func TestRecentConnectionsStore_Add_LoadErrorWithSaveError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.AddFile("/test/path", []byte("corrupted data"))
	store := NewRecentConnectionsStore("/test/path", &fs)

	fs.WriteErr = errors.New("write permission denied")

	newConn := Model{
		Country:        "Spain",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err := store.Add(newConn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adding new recent vpn connection")
	assert.Contains(t, err.Error(), "recreating recent connections file")
}

func TestRecentConnectionsStore_Clean_WriteError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn := Model{
		Country:        "Norway",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	err := store.Add(conn)
	require.NoError(t, err)

	fs.WriteErr = errors.New("permission denied")

	err = store.Clean()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cleaning existing recent vpn connections")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRecentConnectionsStore_Add_SpecificServerWithGroup(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn1 := Model{
		SpecificServerName: "uk1234",
		SpecificServer:     "uk1234",
		Group:              config.ServerGroup_P2P,
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	}

	conn2 := Model{
		SpecificServerName: "uk1234",
		SpecificServer:     "uk1234",
		Group:              config.ServerGroup_DOUBLE_VPN,
		ConnectionType:     config.ServerSelectionRule_SPECIFIC_SERVER_WITH_GROUP,
	}

	err := store.Add(conn1)
	require.NoError(t, err)

	err = store.Add(conn2)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, 2, "Different groups should create separate entries")
}

func TestRecentConnectionsStore_CheckExistence_WriteError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.WriteErr = errors.New("permission denied")
	store := NewRecentConnectionsStore("/test/path", &fs)

	err := store.checkExistence()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creating new recent vpn connections store")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRecentConnectionsStore_Get_CheckExistenceError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	// Set write error to make checkExistence fail when trying to create the file
	fs.WriteErr = errors.New("permission denied")
	store := NewRecentConnectionsStore("/test/path", &fs)

	connections, err := store.Get()

	assert.Error(t, err)
	assert.Nil(t, connections)
	assert.Contains(t, err.Error(), "getting recent connections")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRecentConnectionsStore_Add_CheckExistenceError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.WriteErr = errors.New("permission denied")
	store := NewRecentConnectionsStore("/test/path", &fs)

	conn := Model{
		Country:        "Spain",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err := store.Add(conn)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adding new vpn connection")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestRecentConnectionsStore_Add_SaveErrorAfterSuccessfulLoad(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// First add a connection successfully
	conn1 := Model{
		Country:        "Germany",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	err := store.Add(conn1)
	require.NoError(t, err)

	// Now set write error to fail on the next save
	fs.WriteErr = errors.New("disk full")

	conn2 := Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	err = store.Add(conn2)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "adding new recent vpn connection")
	assert.Contains(t, err.Error(), "disk full")
}

func TestRecentConnectionsStore_Add_DifferentConnectionTypes(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	tests := []struct {
		name           string
		conn1          Model
		conn2          Model
		expectSeparate bool
		description    string
	}{
		{
			name: "CITY connections with different specific servers are treated as same",
			conn1: Model{
				Country:        "Germany",
				City:           "Berlin",
				SpecificServer: "de123",
				ConnectionType: config.ServerSelectionRule_CITY,
			},
			conn2: Model{
				Country:        "Germany",
				City:           "Berlin",
				SpecificServer: "de456",
				ConnectionType: config.ServerSelectionRule_CITY,
			},
			expectSeparate: false,
			description:    "CITY type excludes specific server fields from matching",
		},
		{
			name: "COUNTRY connections with different specific servers are treated as same",
			conn1: Model{
				Country:        "France",
				SpecificServer: "fr123",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			},
			conn2: Model{
				Country:        "France",
				SpecificServer: "fr456",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			},
			expectSeparate: false,
			description:    "COUNTRY type excludes specific server fields from matching",
		},
		{
			name: "RECOMMENDED connections are all treated as same",
			conn1: Model{
				SpecificServer: "us123",
				ConnectionType: config.ServerSelectionRule_RECOMMENDED,
			},
			conn2: Model{
				SpecificServer: "uk456",
				ConnectionType: config.ServerSelectionRule_RECOMMENDED,
			},
			expectSeparate: false,
			description:    "RECOMMENDED type excludes specific server fields from matching",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean store before each test
			err := store.Clean()
			require.NoError(t, err)

			err = store.Add(tt.conn1)
			require.NoError(t, err)

			err = store.Add(tt.conn2)
			require.NoError(t, err)

			connections, err := store.Get()
			require.NoError(t, err)

			if tt.expectSeparate {
				assert.Len(t, connections, 2, tt.description)
			} else {
				assert.Len(t, connections, 1, tt.description)
			}
		})
	}
}

func TestRecentConnectionsStore_AddPending_StoresPendingConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	model := Model{
		Country:        "Germany",
		City:           "Berlin",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "DE",
	}

	store.AddPending(model)

	// Verify the pending connection is stored
	exists, retrieved := store.PopPending()
	require.True(t, exists)
	assert.Equal(t, model, retrieved)
}

func TestRecentConnectionsStore_AddPending_OverwritesPreviousPending(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	model1 := Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	model2 := Model{
		Country:        "Spain",
		City:           "Madrid",
		ConnectionType: config.ServerSelectionRule_CITY,
	}

	store.AddPending(model1)
	store.AddPending(model2)

	exists, retrieved := store.PopPending()
	require.True(t, exists)
	assert.Equal(t, model2, retrieved)
	assert.NotEqual(t, model1, retrieved)
}

func TestRecentConnectionsStore_AddPending_ConcurrentAccess(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			model := Model{
				Country:        fmt.Sprintf("Country%d", idx),
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			}
			store.AddPending(model)
		}(i)
	}

	wg.Wait()

	// Should have one pending connection (the last one written)
	exists, retrieved := store.PopPending()
	require.True(t, exists)
	assert.NotEmpty(t, retrieved.Country)
}

func TestRecentConnectionsStore_PopPending_ReturnsAndClearsPending(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	model := Model{
		Country:        "Italy",
		City:           "Rome",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "IT",
	}

	store.AddPending(model)

	// First PopPending should return the model
	exists, retrieved := store.PopPending()
	require.True(t, exists)
	assert.Equal(t, model, retrieved)

	// Second PopPending should return false (no pending connection)
	exists, empty := store.PopPending()
	assert.False(t, exists)
	assert.True(t, empty.IsEmpty())
}

func TestRecentConnectionsStore_PopPending_NoPendingConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	exists, model := store.PopPending()
	assert.False(t, exists)
	assert.True(t, model.IsEmpty())
}

func TestRecentConnectionsStore_PopPending_ReturnsClone(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	original := Model{
		Country:            "Netherlands",
		City:               "Amsterdam",
		ConnectionType:     config.ServerSelectionRule_CITY,
		CountryCode:        "NL",
		ServerTechnologies: []core.ServerTechnology{1, 2, 3},
	}

	store.AddPending(original)

	exists, retrieved := store.PopPending()
	require.True(t, exists)

	// Modify the retrieved model
	retrieved.Country = "Belgium"
	retrieved.ServerTechnologies[0] = 999

	// Add the same pending again and verify it wasn't affected
	store.AddPending(original)
	exists, retrieved2 := store.PopPending()
	require.True(t, exists)
	assert.Equal(t, original, retrieved2)
	assert.NotEqual(t, retrieved, retrieved2)
}

func TestRecentConnectionsStore_PopPending_ConcurrentAccess(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	model := Model{
		Country:        "Sweden",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	store.AddPending(model)

	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exists, _ := store.PopPending()
			mu.Lock()
			defer mu.Unlock()
			if exists {
				successCount++
			} else {
				errorCount++
			}
		}()
	}

	wg.Wait()

	// Only one goroutine should succeed in getting the pending connection
	assert.Equal(t, 1, successCount, "Only one PopPending should succeed")
	assert.Equal(t, goroutines-1, errorCount, "Other PopPending calls should fail")
}

func TestRecentConnectionsStore_AddPending_EmptyModel(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	emptyModel := Model{}
	store.AddPending(emptyModel)

	exists, retrieved := store.PopPending()
	assert.False(t, exists)
	assert.True(t, retrieved.IsEmpty())
}

func TestRecentConnectionsStore_PendingWorkflow_FullCycle(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	// Step 1: Add a pending connection
	model := Model{
		Country:        "Poland",
		City:           "Warsaw",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "PL",
	}
	store.AddPending(model)

	// Step 2: Retrieve and clear the pending connection
	exists, retrieved := store.PopPending()
	require.True(t, exists)
	assert.Equal(t, model, retrieved)

	// Step 3: Add the retrieved connection to the store
	err := store.Add(retrieved)
	require.NoError(t, err)

	// Step 4: Verify the connection is in the store
	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, model, connections[0])

	// Step 5: Verify no pending connection remains
	exists, empty := store.PopPending()
	assert.False(t, exists)
	assert.True(t, empty.IsEmpty())
}
