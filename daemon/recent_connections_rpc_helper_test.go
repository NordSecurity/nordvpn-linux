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

func TestStorePendingRecentConnection_BasicBehavior(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		pendingModel    *recents.Model
		expectedStored  int
		expectedEvent   bool
		expectedCleared bool
		validateStored  func(*testing.T, []recents.Model)
	}{
		{
			name: "with pending connection",
			pendingModel: &recents.Model{
				Country:        "Germany",
				City:           "Berlin",
				ConnectionType: config.ServerSelectionRule_CITY,
				CountryCode:    "DE",
			},
			expectedStored:  1,
			expectedEvent:   true,
			expectedCleared: true,
			validateStored: func(t *testing.T, connections []recents.Model) {
				assert.Equal(t, "Germany", connections[0].Country)
				assert.Equal(t, "Berlin", connections[0].City)
			},
		},
		{
			name:            "no pending connection",
			pendingModel:    nil,
			expectedStored:  0,
			expectedEvent:   false,
			expectedCleared: false,
		},
		{
			name:            "empty model not stored",
			pendingModel:    &recents.Model{},
			expectedStored:  0,
			expectedEvent:   false,
			expectedCleared: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := mockconfig.NewFilesystemMock(t)
			store := recents.NewRecentConnectionsStore("/test/path", &fs)

			if tt.pendingModel != nil {
				store.AddPending(*tt.pendingModel)
			}

			eventPublished := false
			eventPublisher := func(data events.DataRecentsChanged) {
				eventPublished = true
			}

			StorePendingRecentConnection(store, eventPublisher)

			// Verify connections stored
			connections, err := store.Get()
			require.NoError(t, err)
			assert.Len(t, connections, tt.expectedStored)

			if tt.validateStored != nil && len(connections) > 0 {
				tt.validateStored(t, connections)
			}

			// Verify event publication
			assert.Equal(t, tt.expectedEvent, eventPublished)

			// Verify pending was cleared if expected
			if tt.expectedCleared {
				exists, empty := store.PopPending()
				assert.False(t, exists)
				assert.True(t, empty.IsEmpty())
			}
		})
	}
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

func TestStorePendingRecentConnection_ErrorHandling(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name             string
		writeError       error
		existingModel    *recents.Model
		pendingModel     recents.Model
		expectedEvent    bool
		expectedCleared  bool
		validateExisting func(*testing.T, []recents.Model)
	}{
		{
			name:       "add error clears pending but does not publish event",
			writeError: errors.New("disk full"),
			pendingModel: recents.Model{
				Country:        "Austria",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			},
			expectedEvent:   false,
			expectedCleared: true,
		},
		{
			name:       "add error does not affect existing connections",
			writeError: errors.New("write error"),
			existingModel: &recents.Model{
				Country:        "Belgium",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			},
			pendingModel: recents.Model{
				Country:        "Luxembourg",
				ConnectionType: config.ServerSelectionRule_COUNTRY,
			},
			expectedEvent:   false,
			expectedCleared: true,
			validateExisting: func(t *testing.T, connections []recents.Model) {
				require.Len(t, connections, 1)
				assert.Equal(t, "Belgium", connections[0].Country)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := mockconfig.NewFilesystemMock(t)
			store := recents.NewRecentConnectionsStore("/test/path", &fs)

			// Add existing connection if specified
			if tt.existingModel != nil {
				err := store.Add(*tt.existingModel)
				require.NoError(t, err)
			}

			// Set write error
			fs.WriteErr = tt.writeError

			// Add pending connection
			store.AddPending(tt.pendingModel)

			eventPublished := false
			eventPublisher := func(data events.DataRecentsChanged) {
				eventPublished = true
			}

			StorePendingRecentConnection(store, eventPublisher)

			// Verify event was not published on error
			assert.Equal(t, tt.expectedEvent, eventPublished)

			// Verify pending was cleared
			if tt.expectedCleared {
				exists, empty := store.PopPending()
				assert.False(t, exists)
				assert.True(t, empty.IsEmpty())
			}

			// Validate existing connections if needed
			if tt.validateExisting != nil {
				fs.WriteErr = nil // Clear error to read
				connections, err := store.Get()
				require.NoError(t, err)
				tt.validateExisting(t, connections)
			}
		})
	}
}
