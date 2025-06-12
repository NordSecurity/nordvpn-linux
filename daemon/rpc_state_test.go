package daemon

import (
	"context"
	"net/netip"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/state/types"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testIP            = "192.168.13.37"
	testUID           = 42
	testServerName    = "test server"
	testCountry       = "test country"
	testCity          = "test city"
	testCountryCode   = "XX"
	testServerDefault = "server1337"
)

var (
	testServerGroup      = config.ServerGroup_ONION_OVER_VPN
	testConnectionParams = ServerParameters{
		ServerName:  testServerName,
		Country:     testCountry,
		City:        testCity,
		Group:       testServerGroup,
		CountryCode: testCountryCode,
	}
)

type testSetup struct {
	srv        *testStateServer
	stateChan  chan any
	stopChan   chan struct{}
	cancel     context.CancelFunc
	connParams *RequestedConnParamsStorage
}

func setupTest(t *testing.T) *testSetup {
	t.Helper()

	stateChan := make(chan any)
	stopChan := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	srv := &testStateServer{
		ctx:    ctx,
		wg:     &wg,
		states: make([]*pb.AppState, 0),
	}

	connParams := &RequestedConnParamsStorage{}
	connParams.Set(
		pb.ConnectionSource_MANUAL,
		testConnectionParams,
	)

	return &testSetup{
		srv:        srv,
		stateChan:  stateChan,
		stopChan:   stopChan,
		cancel:     cancel,
		connParams: connParams,
	}
}

func verifyConnectionStatus(t *testing.T, expected types.ConnectionStatus, actual *pb.StatusResponse) {
	t.Helper()
	assert.Equal(t, expected.State, actual.State)
	assert.Equal(t, expected.Technology, actual.Technology)
	assert.Equal(t, expected.Protocol, actual.Protocol)
	assert.Equal(t, expected.IP.String(), actual.Ip)
	assert.Equal(t, expected.Name, actual.Name)
}

func verifyConnectionParameters(t *testing.T, params *pb.ConnectionParameters) {
	t.Helper()
	require.NotNil(t, params)
	assert.Equal(t, testServerName, params.ServerName)
	assert.Equal(t, pb.ConnectionSource_MANUAL, params.Source)
	assert.Equal(t, testCountry, params.Country)
	assert.Equal(t, testCity, params.City)
	assert.Equal(t, testServerGroup, params.Group)
	assert.Equal(t, testCountryCode, params.CountryCode)
}

type testStateServer struct {
	pb.Daemon_SubscribeToStateChangesServer
	ctx    context.Context
	wg     *sync.WaitGroup
	states []*pb.AppState
}

func (s *testStateServer) Send(state *pb.AppState) error {
	s.states = append(s.states, state)
	s.wg.Done()
	return nil
}

func (s *testStateServer) Context() context.Context {
	return s.ctx
}

// TestRpcState_HandleDataConnectChangedEvents tests the functionality of statusStream function
// in handling DataConnectChangeNotif events. It verifies that:
//   - When connection state is DISCONNECTED, no connection parameters shall be provided in the app state
//   - When connection state is CONNECTED, all connection parameters shall be provided in the app state
func TestRpcState_HandleDataConnectChangedEvents(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                     string
		connectionStatus         types.ConnectionStatus
		connectionParamsProvided bool
	}{
		{
			name: "when disconnected connection parameters shall not be provided",
			connectionStatus: types.ConnectionStatus{
				State:      pb.ConnectionState_DISCONNECTED,
				Technology: config.Technology_NORDLYNX,
				Protocol:   config.Protocol_UDP,
				IP:         netip.MustParseAddr(testIP),
				Name:       testServerDefault,
			},
			connectionParamsProvided: false,
		},
		{
			name: "when connected connection parameters shall be provided",
			connectionStatus: types.ConnectionStatus{
				State:      pb.ConnectionState_CONNECTED,
				Technology: config.Technology_NORDLYNX,
				Protocol:   config.Protocol_UDP,
				IP:         netip.MustParseAddr(testIP),
				Name:       testServerDefault,
			},
			connectionParamsProvided: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := setupTest(t)

			go statusStream(ts.stateChan, ts.stopChan, testUID, ts.srv, ts.connParams)

			ts.stateChan <- events.DataConnectChangeNotif{
				Status: tt.connectionStatus,
			}

			ts.srv.wg.Wait()

			require.Len(t, ts.srv.states, 1)
			status := ts.srv.states[0].GetConnectionStatus()
			require.NotNil(t, status)

			verifyConnectionStatus(t, tt.connectionStatus, status)

			if tt.connectionParamsProvided {
				verifyConnectionParameters(t, status.Parameters)
			} else {
				require.Nil(t, status.Parameters)
			}
		})
	}
}
