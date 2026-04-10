package state

import (
	"net/netip"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

type TestSubscriber struct {
	notifications []events.DataConnectChangeNotif
	wg            *sync.WaitGroup
}

func NewTestSubscriber() *TestSubscriber {
	return &TestSubscriber{
		wg: &sync.WaitGroup{},
	}
}

// PopNotification returns first received notification and removes it from the notifications list. Returns false if
// notifications list is empty.
func (s *TestSubscriber) PopNotification() (events.DataConnectChangeNotif, bool) {
	if len(s.notifications) == 0 {
		return events.DataConnectChangeNotif{}, false
	}
	notification := s.notifications[0]
	s.notifications = s.notifications[1:]
	return notification, true
}

func (s *TestSubscriber) GetNotificationsCount() int {
	return len(s.notifications)
}

func (s *TestSubscriber) NotifyChangeState(e events.DataConnectChangeNotif) error {
	s.notifications = append(s.notifications, e)
	s.wg.Done()
	return nil
}

func (s *TestSubscriber) ExpectEvents(count int) {
	s.wg.Add(count)
}

type testFixture struct {
	sut        *ConnectionInfo
	subscriber *TestSubscriber
	done       chan struct{}
}

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()

	s := NewTestSubscriber()
	sut := NewConnectionInfo()
	events := vpn.NewInternalVPNEvents()
	events.Subscribe(sut)
	sut.SubscribeToInternalStateChanges(s)

	return &testFixture{
		sut:        sut,
		subscriber: s,
		done:       make(chan struct{}),
	}
}

func (f *testFixture) waitForCompletion(t *testing.T) {
	t.Helper()
	select {
	case <-f.done:
	case <-time.After(5000 * time.Second):
		t.Fatal("Timeout waiting for subscriber notifications")
	}
}

func TestConnectionInfo_VerifyDataConnectConversionToConnectionStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	event := events.DataConnect{
		Technology:              config.Technology_OPENVPN,
		Protocol:                config.Protocol_UDP,
		EventStatus:             events.StatusSuccess,
		TargetServerName:        "server1",
		TargetServerDomain:      "42.example.pl",
		TargetServerCountry:     "Test Country",
		TargetServerCountryCode: "TC",
		TargetServerCity:        "Test City",
		TargetServerIP:          netip.MustParseAddr("192.168.1.1"),
		IsVirtualLocation:       true,
		IsPostQuantum:           false,
		IsObfuscated:            true,
		IsMeshnetPeer:           false,
		RecommendationUUID:      "c0b4c990-3000-457f-8b81-6850b8cdb54e",
	}

	go func() {
		tf.sut.ConnectionStatusNotifyConnect(event)
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)

	status := tf.sut.Status()
	assert.Equal(t, pb.ConnectionState_CONNECTED, status.State)
	assert.True(t, tf.sut.fullyConnected)
	assert.Equal(t, event.Technology, status.Technology)
	assert.Equal(t, event.Protocol, status.Protocol)
	assert.Equal(t, event.TargetServerName, status.Name)
	assert.Equal(t, event.TargetServerDomain, status.Hostname)
	assert.Equal(t, event.TargetServerCountry, status.Country)
	assert.Equal(t, event.TargetServerCountryCode, status.CountryCode)
	assert.Equal(t, event.TargetServerCity, status.City)
	assert.Equal(t, event.TargetServerIP, status.IP)
	assert.Equal(t, event.IsVirtualLocation, status.IsVirtualLocation)
	assert.Equal(t, event.IsPostQuantum, status.IsPostQuantum)
	assert.Equal(t, event.IsObfuscated, status.IsObfuscated)
	assert.Equal(t, "", status.TunnelName)
	assert.Equal(t, event.IsMeshnetPeer, status.IsMeshnetPeer)
	assert.Equal(t, event.RecommendationUUID, status.RecommendationUUID)
}

func TestConnectionInfo_VerifyDataDisconnectConversionToConnectionStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	go func() {
		tf.sut.ConnectionStatusNotifyDisconnect(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)

	status := tf.sut.Status()
	assert.Equal(t, pb.ConnectionState_DISCONNECTED, status.State)
	assert.False(t, tf.sut.fullyConnected)
	assert.Nil(t, status.StartTime)
	assert.Equal(t, status.RecommendationUUID, "")
}

func TestConnectionInfo_TracksStateProperly(t *testing.T) {
	category.Set(t, category.Unit)

	for _, tt := range []struct {
		name           string
		eventStatus    events.TypeEventStatus
		state          pb.ConnectionState
		fullyConnected bool
	}{
		{
			name:           "success",
			eventStatus:    events.StatusSuccess,
			state:          pb.ConnectionState_CONNECTED,
			fullyConnected: true,
		},
		{
			name:           "canceled",
			eventStatus:    events.StatusCanceled,
			state:          pb.ConnectionState_DISCONNECTED,
			fullyConnected: false,
		},
		{
			name:           "failure",
			eventStatus:    events.StatusFailure,
			state:          pb.ConnectionState_DISCONNECTED,
			fullyConnected: false,
		},
		{
			name:           "attempt",
			eventStatus:    events.StatusAttempt,
			state:          pb.ConnectionState_CONNECTING,
			fullyConnected: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tf := newTestFixture(t)
			tf.subscriber.ExpectEvents(1)

			go func() {
				assert.Equal(t, pb.ConnectionState_UNKNOWN_STATE, tf.sut.Status().State)
				tf.sut.ConnectionStatusNotifyConnect(
					events.DataConnect{EventStatus: tt.eventStatus},
				)
				tf.subscriber.wg.Wait()

				assert.Equal(t, tt.state, tf.sut.Status().State)
				assert.Equal(t, tt.fullyConnected, tf.sut.fullyConnected)

				tf.subscriber.ExpectEvents(1)
				tf.sut.ConnectionStatusNotifyDisconnect(events.DataDisconnect{})
				tf.subscriber.wg.Wait()
				close(tf.done)
			}()

			tf.waitForCompletion(t)
			assert.Equal(t, pb.ConnectionState_DISCONNECTED, tf.sut.Status().State)
			assert.False(t, tf.sut.fullyConnected)
		})
	}
}

func TestConnectionInfo_TracksStartTime(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	go func() {
		status := tf.sut.Status()
		assert.Nil(t, status.StartTime)

		event := events.DataConnect{
			EventStatus: events.StatusSuccess,
		}
		tf.sut.ConnectionStatusNotifyConnect(event)
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.NotNil(t, status.StartTime)

		tf.subscriber.ExpectEvents(1)
		tf.sut.ConnectionStatusNotifyDisconnect(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)

	status := tf.sut.Status()
	assert.Nil(t, status.StartTime)
}

func TestConnectionInfo_TransferRatesShallBeProvidedOnlyForConnectedState(t *testing.T) {
	category.Set(t, category.Link)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	go func() {
		status := tf.sut.Status()
		assert.Zero(t, status.Tx)
		assert.Zero(t, status.Rx)

		// Update the tunnel name in connection state
		event := events.DataConnect{
			EventStatus: events.StatusAttempt,
		}
		tf.sut.ConnectionStatusNotifyConnect(event)

		// Transfer rates should be zero after connection attempt.
		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusSuccess, TunnelName: "lo"},
		)
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.Zero(t, status.Tx)
		assert.Zero(t, status.Rx)

		// Transfer rates should be non zero after connection is established.
		event = events.DataConnect{
			EventStatus: events.StatusSuccess,
		}
		tf.subscriber.ExpectEvents(1)
		tf.sut.ConnectionStatusNotifyConnect(event)
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.NotZero(t, status.Tx)
		assert.NotZero(t, status.Rx)

		// Transfer rates should be zero after connection failure.
		event = events.DataConnect{
			EventStatus: events.StatusFailure,
		}
		tf.subscriber.ExpectEvents(1)
		tf.sut.ConnectionStatusNotifyConnect(event)
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.Zero(t, status.Tx)
		assert.Zero(t, status.Rx)

		// Transfer rates should be zero after a disconnect.
		tf.subscriber.ExpectEvents(1)
		tf.sut.ConnectionStatusNotifyDisconnect(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.Zero(t, status.Tx)
		assert.Zero(t, status.Rx)

		// Transfer rates should be zero after successful connection when tunnel name was
		// not updated by the internal event.
		tf.subscriber.ExpectEvents(1)
		tf.sut.ConnectionStatusNotifyConnect(events.DataConnect{
			EventStatus: events.StatusSuccess,
		})
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.Zero(t, status.Tx)
		assert.Zero(t, status.Rx)

		close(tf.done)
	}()

	tf.waitForCompletion(t)
}

func TestConnectionInfo_InternalEventsIgnoredUntilFullyConnected(t *testing.T) {
	category.Set(t, category.Unit)
	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(4)

	go func() {
		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusAttempt})
		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusSuccess})
		assert.Equal(t, 0, tf.subscriber.GetNotificationsCount())

		tf.sut.ConnectionStatusNotifyConnect(
			events.DataConnect{EventStatus: events.StatusAttempt},
		)
		assert.Equal(t, 1, tf.subscriber.GetNotificationsCount())

		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusAttempt})
		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusSuccess})
		assert.Equal(t, 1, tf.subscriber.GetNotificationsCount())

		tf.sut.ConnectionStatusNotifyConnect(
			events.DataConnect{EventStatus: events.StatusSuccess},
		)
		assert.Equal(t, 2, tf.subscriber.GetNotificationsCount())

		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusAttempt})
		tf.sut.ConnectionStatusNotifyInternalConnect(
			vpn.ConnectEvent{Status: events.StatusSuccess})
		assert.Equal(t, 4, tf.subscriber.GetNotificationsCount())
		close(tf.done)
	}()
	tf.waitForCompletion(t)
}

func TestConnectionInfo_TunnelNameUpdatedByInternalEventsOnly(t *testing.T) {
	category.Set(t, category.Unit)
	tun0 := "tun0"
	tun1 := "tun1"
	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(5)
	tf.sut.ConnectionStatusNotifyConnect(events.DataConnect{EventStatus: events.StatusSuccess})
	assert.Empty(t, tf.sut.Status().TunnelName)
	tf.sut.ConnectionStatusNotifyInternalConnect(vpn.ConnectEvent{
		Status: events.StatusAttempt, TunnelName: tun0,
	})
	assert.Equal(t, tun0, tf.sut.Status().TunnelName)

	tf.sut.ConnectionStatusNotifyInternalConnect(vpn.ConnectEvent{
		Status: events.StatusAttempt, TunnelName: tun1,
	})
	assert.Equal(t, tun1, tf.sut.Status().TunnelName)

	tf.sut.ConnectionStatusNotifyConnect(events.DataConnect{EventStatus: events.StatusSuccess})
	assert.Equal(t, tun1, tf.sut.Status().TunnelName)

	tf.sut.ConnectionStatusNotifyDisconnect(
		events.DataDisconnect{EventStatus: events.StatusSuccess})
	assert.Empty(t, tf.sut.Status().TunnelName)

	tf.sut.ConnectionStatusNotifyInternalConnect(vpn.ConnectEvent{
		Status: events.StatusAttempt, TunnelName: tun0,
	})
	assert.Equal(t, tun0, tf.sut.Status().TunnelName)

	tf.sut.ConnectionStatusNotifyInternalDisconnect(events.StatusSuccess)
	assert.Empty(t, tf.sut.Status().TunnelName)
}

func TestConnectionInfo_RefreshDisconnectEventsAreIgnored(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                  string
		disconnectEventStatus events.TypeEventStatus
	}{
		{
			name:                  "disconnect success",
			disconnectEventStatus: events.StatusSuccess,
		},
		{
			name:                  "disconnect failure",
			disconnectEventStatus: events.StatusFailure,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tf := newTestFixture(t)
			tf.subscriber.ExpectEvents(2)
			tf.sut.ConnectionStatusNotifyConnect(events.DataConnect{EventStatus: events.StatusSuccess})
			assert.Equal(t, pb.ConnectionState_CONNECTED, tf.sut.status.State,
				"State was not changed to connected after receiving a connect event.")

			tf.sut.ConnectionStatusNotifyDisconnect(events.DataDisconnect{EventStatus: events.StatusSuccess, IsRefresh: true})
			assert.Equal(t, pb.ConnectionState_CONNECTED, tf.sut.status.State,
				"State was changed after receiving a refresh disconnect events. Refresh events should be ignored.")
		})
	}
}

func TestConnectionInfo_PauseHandling(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	duration := 10 * time.Second
	pauseTime := time.Unix(1774276303, 0)
	tf.sut.Pause(pauseTime, duration)

	tf.sut.remainingDurationFunc = func(time.Time, time.Duration) uint32 {
		return 5
	}

	status := tf.sut.Status()
	assert.Equal(t, pauseTime, status.PausedAt)
	assert.Equal(t, status.State, pb.ConnectionState_PAUSED)
	assert.Equal(t, status.PauseRemainingTimeSec, uint32(5))

	go func() {
		tf.subscriber.ExpectEvents(1)
		tf.sut.ConnectionStatusNotifyDisconnect(events.DataDisconnect{})
		tf.subscriber.wg.Wait()

		assert.Equal(t, 1, tf.subscriber.GetNotificationsCount(),
			"Unexpected number of notifications after a disconnect event.")

		notification, notificationReceived := tf.subscriber.PopNotification()
		assert.True(t, notificationReceived,
			"Expected notification not received when DataDisconnect event was emitted.")
		assert.Equal(t, notification.Status.State, pb.ConnectionState_PAUSED,
			"Unexpected state received in the status notification.")
		assert.Equal(t, pauseTime, notification.Status.PausedAt,
			"Unexpected pause time in a disconnect notification.")
		assert.Equal(t, notification.Status.PauseRemainingTimeSec, uint32(5),
			"Unexpected remaining pause time in a disconnect notification.")

		tf.subscriber.ExpectEvents(1)

		tf.sut.remainingDurationFunc = func(time.Time, time.Duration) uint32 {
			return 2
		}

		tf.sut.fullyConnected = true
		tf.sut.ConnectionStatusNotifyInternalDisconnect(events.StatusSuccess)

		assert.Equal(t, 1, tf.subscriber.GetNotificationsCount(),
			"Unexpected number of notifications after an internal disconnect event.")

		notification, notificationReceived = tf.subscriber.PopNotification()
		assert.True(t, notificationReceived,
			"Expected notification not received when internal disconnect event was emitted.")
		assert.Equal(t, notification.Status.State, pb.ConnectionState_PAUSED,
			"Unexpected state received in the status notification.")
		assert.Equal(t, notification.Status.PausedAt, pauseTime,
			"Unexpected pause time in an internal disconnect notification.")
		assert.Equal(t, notification.Status.PauseRemainingTimeSec, uint32(2),
			"Unexpected remaining pause time in an internal disconnect notification.")
		close(tf.done)
	}()
	tf.waitForCompletion(t)
}
