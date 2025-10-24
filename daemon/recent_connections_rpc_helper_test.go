package daemon

import (
	"errors"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	mockconfig "github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorePendingRecentConnection_WithPendingConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	model := recents.Model{
		Country:        "Germany",
		City:           "Berlin",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "DE",
	}

	store.AddPending(model)

	eventPublished := false
	eventPublisher := func(data events.DataRecentsChanged) {
		eventPublished = true
	}

	StorePendingRecentConnection(store, eventPublisher)

	// Verify the connection was added to the store
	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, model, connections[0])

	// Verify the event was published
	assert.True(t, eventPublished, "RecentsChanged event should be published")

	// Verify pending connection was cleared
	err, empty := store.GetPending()
	assert.Error(t, err)
	assert.True(t, empty.IsEmpty())
}

func TestStorePendingRecentConnection_NoPendingConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	eventPublished := false
	eventPublisher := func(data events.DataRecentsChanged) {
		eventPublished = true
	}

	StorePendingRecentConnection(store, eventPublisher)

	// Verify no connection was added
	connections, err := store.Get()
	require.NoError(t, err)
	assert.Empty(t, connections)

	// Verify no event was published
	assert.False(t, eventPublished, "RecentsChanged event should not be published when no pending connection")
}

func TestStorePendingRecentConnection_MultiplePendingConnections(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	model1 := recents.Model{
		Country:        "France",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	model2 := recents.Model{
		Country:        "Spain",
		City:           "Madrid",
		ConnectionType: config.ServerSelectionRule_CITY,
	}

	// Add first pending
	store.AddPending(model1)

	eventCount := 0
	eventPublisher := func(data events.DataRecentsChanged) {
		eventCount++
	}

	// Store first pending
	StorePendingRecentConnection(store, eventPublisher)

	// Add second pending
	store.AddPending(model2)

	// Store second pending
	StorePendingRecentConnection(store, eventPublisher)

	// Verify both connections were added
	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 2)
	assert.Equal(t, model2, connections[0]) // Most recent first
	assert.Equal(t, model1, connections[1])

	// Verify event was published twice
	assert.Equal(t, 2, eventCount, "RecentsChanged event should be published twice")
}

func TestStorePendingRecentConnection_DuplicateConnection(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	model := recents.Model{
		Country:        "Italy",
		City:           "Rome",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "IT",
	}

	// Add and store the connection first time
	store.AddPending(model)

	eventCount := 0
	eventPublisher := func(data events.DataRecentsChanged) {
		eventCount++
	}

	StorePendingRecentConnection(store, eventPublisher)

	// Add and store the same connection again
	store.AddPending(model)
	StorePendingRecentConnection(store, eventPublisher)

	// Verify only one connection exists (moved to front)
	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, model, connections[0])

	// Verify event was published twice
	assert.Equal(t, 2, eventCount)
}

func TestStorePendingRecentConnection_ConcurrentCalls(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	var eventMu sync.Mutex
	eventCount := 0
	eventPublisher := func(data events.DataRecentsChanged) {
		eventMu.Lock()
		defer eventMu.Unlock()
		eventCount++
	}

	var wg sync.WaitGroup
	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			model := recents.Model{
				Country:        "Country" + string(rune('A'+idx)),
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			}
			store.AddPending(model)
			StorePendingRecentConnection(store, eventPublisher)
		}(i)
	}

	wg.Wait()

	// Verify connections were added
	connections, err := store.Get()
	require.NoError(t, err)
	assert.LessOrEqual(t, len(connections), goroutines)

	// Verify events were published
	eventMu.Lock()
	defer eventMu.Unlock()
	assert.Greater(t, eventCount, 0, "At least one event should be published")
}

func TestStorePendingRecentConnection_WithServerTechnologies(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	model := recents.Model{
		Country:            "Netherlands",
		City:               "Amsterdam",
		ConnectionType:     config.ServerSelectionRule_CITY,
		CountryCode:        "NL",
		ServerTechnologies: []core.ServerTechnology{1, 3, 5},
	}

	store.AddPending(model)

	eventPublished := false
	eventPublisher := func(data events.DataRecentsChanged) {
		eventPublished = true
	}

	StorePendingRecentConnection(store, eventPublisher)

	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, model.Country, connections[0].Country)
	assert.Equal(t, model.City, connections[0].City)
	assert.ElementsMatch(t, model.ServerTechnologies, connections[0].ServerTechnologies)
	assert.True(t, eventPublished)
}

func TestStorePendingRecentConnection_FullWorkflow(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	eventCount := 0
	eventPublisher := func(data events.DataRecentsChanged) {
		eventCount++
	}

	// Simulate connect workflow
	connectModel := recents.Model{
		Country:        "Poland",
		City:           "Warsaw",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "PL",
	}

	// Step 1: Connection is established, pending is added
	store.AddPending(connectModel)

	// Step 2: Disconnect happens, StorePendingRecentConnection is called
	StorePendingRecentConnection(store, eventPublisher)

	// Verify the connection was stored
	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, connectModel, connections[0])
	assert.Equal(t, 1, eventCount)

	// Step 3: Another connection
	reconnectModel := recents.Model{
		Country:        "Poland",
		City:           "Krakow",
		ConnectionType: config.ServerSelectionRule_CITY,
		CountryCode:    "PL",
	}

	store.AddPending(reconnectModel)
	StorePendingRecentConnection(store, eventPublisher)

	// Verify both connections are stored
	connections, err = store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 2)
	assert.Equal(t, reconnectModel, connections[0]) // Most recent first
	assert.Equal(t, connectModel, connections[1])
	assert.Equal(t, 2, eventCount)
}

func TestStorePendingRecentConnection_EmptyModelNotStored(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	emptyModel := recents.Model{}
	store.AddPending(emptyModel)

	eventPublished := false
	eventPublisher := func(data events.DataRecentsChanged) {
		eventPublished = true
	}

	StorePendingRecentConnection(store, eventPublisher)

	// Verify no connection was added
	connections, err := store.Get()
	require.NoError(t, err)
	assert.Empty(t, connections)

	// Verify no event was published
	assert.False(t, eventPublished)
}

func TestStorePendingRecentConnection_AddError(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	fs.WriteErr = errors.New("disk full")
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	model := recents.Model{
		Country:        "Austria",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}

	store.AddPending(model)

	eventPublished := false
	eventPublisher := func(data events.DataRecentsChanged) {
		eventPublished = true
	}

	StorePendingRecentConnection(store, eventPublisher)

	// Verify event was NOT published due to Add error
	assert.False(t, eventPublished, "RecentsChanged event should not be published when Add fails")

	// Verify pending connection was still cleared (GetPending was called)
	err, empty := store.GetPending()
	assert.Error(t, err)
	assert.True(t, empty.IsEmpty())
}

func TestStorePendingRecentConnection_AddErrorDoesNotAffectExisting(t *testing.T) {
	category.Set(t, category.Unit)

	fs := mockconfig.NewFilesystemMock(t)
	store := recents.NewRecentConnectionsStore("/test/path", &fs)

	// Add an existing connection successfully
	existingModel := recents.Model{
		Country:        "Belgium",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	err := store.Add(existingModel)
	require.NoError(t, err)

	// Now simulate an error when adding new pending
	fs.WriteErr = errors.New("write error")

	newModel := recents.Model{
		Country:        "Luxembourg",
		ConnectionType: config.ServerSelectionRule_COUNTRY,
	}
	store.AddPending(newModel)

	eventPublished := false
	eventPublisher := func(data events.DataRecentsChanged) {
		eventPublished = true
	}

	StorePendingRecentConnection(store, eventPublisher)

	// Verify event was NOT published
	assert.False(t, eventPublished)

	// Clear the error to read existing connections
	fs.WriteErr = nil

	// Verify the existing connection is still there
	connections, err := store.Get()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, existingModel, connections[0])
}
