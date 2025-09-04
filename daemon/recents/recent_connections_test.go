package recents

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
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

	assert.Equal(t, 0, store.find(models[0], models))
	assert.Equal(t, 1, store.find(models[1], models))
	assert.Equal(t, 2, store.find(models[2], models))
	assert.Equal(t, 3, store.find(models[3], models))

	nonExisting := Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	assert.Equal(t, -1, store.find(nonExisting, models))
}

func TestRecentConnectionsStore_Find_DifferentServersAreDifferent(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := NewRecentConnectionsStore("/test/path", &fs)

	cityConn1 := Model{
		Country:        "Germany",
		City:           "Berlin",
		SpecificServer: "de123",
		ConnectionType: config.ServerSelectionRule_CITY,
	}
	cityConn2 := Model{
		Country:        "Germany",
		City:           "Berlin",
		SpecificServer: "de456",
		ConnectionType: config.ServerSelectionRule_CITY,
	}

	err := store.Add(cityConn1)
	require.NoError(t, err)
	err = store.Add(cityConn2)
	require.NoError(t, err)

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, 2)
}

func TestRecentConnectionsStore_Find_AllFieldsMustMatch(t *testing.T) {
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
		ConnectionType:     config.ServerSelectionRule_CITY,
	}

	variations := []Model{
		{Country: "Canada", City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType},
		{Country: base.Country, City: "Los Angeles", Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType},
		{Country: base.Country, City: base.City, Group: config.ServerGroup_DOUBLE_VPN, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: "CA", SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: "us #5678", SpecificServer: base.SpecificServer, ConnectionType: base.ConnectionType},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: "us5678", ConnectionType: base.ConnectionType},
		{Country: base.Country, City: base.City, Group: base.Group, CountryCode: base.CountryCode, SpecificServerName: base.SpecificServerName, SpecificServer: base.SpecificServer, ConnectionType: config.ServerSelectionRule_COUNTRY},
	}

	err := store.Add(base)
	require.NoError(t, err)

	for _, variation := range variations {
		err = store.Add(variation)
		require.NoError(t, err)
	}

	connections, err := store.Get()
	require.NoError(t, err)
	assert.Len(t, connections, len(variations)+1)
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
