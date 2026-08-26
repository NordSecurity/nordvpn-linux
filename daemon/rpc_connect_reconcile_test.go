package daemon

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	testnetworker "github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hookedNetworker lets a test observe (and suspend) the moment
// reconcileStatusAfterFailedConnect inspects the tunnel state.
type hookedNetworker struct {
	*testnetworker.Mock
	onIsVPNActive func()
}

func (n *hookedNetworker) IsVPNActive() bool {
	if n.onIsVPNActive != nil {
		n.onIsVPNActive()
	}
	return n.VpnActive
}

// suspendOnReconcile returns a networker whose IsVPNActive call (the first thing
// reconcileStatusAfterFailedConnect does) signals `entered` and then blocks until `release` is
// closed, plus the resulting *RPC.
func suspendOnReconcile(t *testing.T, vpnActive bool) (rpc *RPC, entered <-chan struct{}, release func()) {
	t.Helper()

	rpc = testRPC()
	enteredCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var once sync.Once

	rpc.netw = &hookedNetworker{
		Mock: &testnetworker.Mock{VpnActive: vpnActive},
		onIsVPNActive: func() {
			once.Do(func() { close(enteredCh) })
			<-releaseCh
		},
	}

	var releaseOnce sync.Once
	return rpc, enteredCh, func() { releaseOnce.Do(func() { close(releaseCh) }) }
}

// failingConnect emulates a connection attempt that gave up without touching the existing tunnel.
func failingConnect(context.Context) (bool, error) { return true, nil }

// TestExecuteConnect_ReconcileRunsOutsideCriticalSection reproduces the defect: because
// sharedctx.Context.TryExecuteWith releases its mutex before returning,
// reconcileStatusAfterFailedConnect runs unprotected and another connect attempt (a second
// `nordvpn connect`, an ENS reconnect, autoconnect or a meshnet exit node connect - they all share
// the same sharedctx.Context) can already be running while the failed attempt is still fixing up
// the reported status.
func TestExecuteConnect_ReconcileRunsOutsideCriticalSection(t *testing.T) {
	category.Set(t, category.Unit)

	rpc, reconcileEntered, releaseReconcile := suspendOnReconcile(t, true)
	defer releaseReconcile()

	connectDone := make(chan error, 1)
	go func() {
		connectDone <- rpc.executeConnect(&mockRPCServer{}, failingConnect)
	}()

	select {
	case <-reconcileEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the failed attempt to start reconciling the status")
	}

	// The failed attempt is in the middle of reconcileStatusAfterFailedConnect. No other connect
	// may be admitted until that cleanup is done, otherwise its status is overwritten by ours.
	admitted := rpc.connectContext.TryExecuteWith(func(context.Context) {})

	releaseReconcile()
	require.NoError(t, <-connectDone)

	assert.False(t, admitted,
		"another connect attempt was admitted while the failed attempt was still reconciling "+
			"the connection status")
}

// TestExecuteConnect_FailedAttemptClobbersNextAttemptStatus shows the user visible consequence of
// the race above: the failed attempt publishes a Disconnect event on top of the status of an
// attempt that is already running, so the daemon reports DISCONNECTED in the middle of a connect.
func TestExecuteConnect_FailedAttemptClobbersNextAttemptStatus(t *testing.T) {
	category.Set(t, category.Unit)

	// no tunnel left behind, so the reconcile takes the "publish Disconnect" branch
	rpc, reconcileEntered, releaseReconcile := suspendOnReconcile(t, false)
	defer releaseReconcile()
	// wire the status up to the event bus the same way cmd/daemon/main.go does
	rpc.events.Service.Disconnect.Subscribe(rpc.connectionInfo.ConnectionStatusNotifyDisconnect)

	connectDone := make(chan error, 1)
	go func() {
		connectDone <- rpc.executeConnect(&mockRPCServer{}, failingConnect)
	}()

	select {
	case <-reconcileEntered:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the failed attempt to start reconciling the status")
	}

	// A second connect attempt starts. It is admitted as soon as the guard lets it in - today that
	// happens while the first attempt is still reconciling, once the reconcile is inside the
	// critical section it happens after the first attempt is fully done. Either way the status it
	// sets must survive.
	secondAttempted := make(chan struct{})
	secondDone := make(chan struct{})
	go func() {
		defer close(secondDone)
		var once sync.Once
		deadline := time.Now().Add(5 * time.Second)
		for {
			admitted := rpc.connectContext.TryExecuteWith(func(context.Context) {
				rpc.connectionInfo.SetInitialConnecting()
			})
			once.Do(func() { close(secondAttempted) })
			if admitted || time.Now().After(deadline) {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case <-secondAttempted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the second connect attempt")
	}

	releaseReconcile()
	require.NoError(t, <-connectDone)
	<-secondDone

	assert.Equal(t, pb.ConnectionState_CONNECTING, rpc.connectionInfo.Status().State,
		"the status of the running connect attempt was overwritten by the cleanup of the "+
			"previous failed attempt")
}

// stateChangeRecorder records the internal connection status notifications.
type stateChangeRecorder struct {
	mu     sync.Mutex
	events []events.DataConnectChangeNotif
}

func (r *stateChangeRecorder) OnStateChange(e events.DataConnectChangeNotif) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return nil
}

func (r *stateChangeRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events)
}

// TestExecuteConnect_FailureBeforeSetInitialConnecting_LeavesStatusUntouched covers the secondary
// defect: reconcileStatusAfterFailedConnect also runs for attempts that bail out before
// SetInitialConnecting is reached (not logged in, config load error, maintenance bail-out). Such an
// attempt never changed the reported status, so the cleanup must not restore a snapshot captured by
// an earlier attempt and drop the live connection to a stale status.
func TestExecuteConnect_FailureBeforeSetInitialConnecting_LeavesStatusUntouched(t *testing.T) {
	category.Set(t, category.Unit)

	rpc := testRPC()
	// the tunnel is up, so the reconcile takes the "restore previous status" branch
	rpc.netw = &testnetworker.Mock{VpnActive: true}

	// an earlier attempt captured a snapshot while disconnected and then connected successfully
	rpc.connectionInfo.SetInitialConnecting()
	require.NoError(t, rpc.connectionInfo.ConnectionStatusNotifyConnect(
		events.DataConnect{EventStatus: events.StatusSuccess, TargetServerName: "server1"}))
	require.Equal(t, pb.ConnectionState_CONNECTED, rpc.connectionInfo.Status().State)

	recorder := &stateChangeRecorder{}
	rpc.connectionInfo.SubscribeToInternalStateChanges(recorder)

	// a new attempt fails before it ever reports CONNECTING
	err := rpc.executeConnect(&mockRPCServer{}, func(context.Context) (bool, error) {
		return true, internal.ErrNotLoggedIn
	})
	assert.ErrorIs(t, err, internal.ErrNotLoggedIn)

	status := rpc.connectionInfo.Status()
	assert.Equal(t, pb.ConnectionState_CONNECTED, status.State,
		"an attempt that failed before reporting CONNECTING must not change the reported status")
	assert.Equal(t, "server1", status.Name)
	assert.Zero(t, recorder.count(),
		"no status change notification should be emitted for an attempt that never changed the status")
}
