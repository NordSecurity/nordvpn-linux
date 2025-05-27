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
	notificationCounter int
	wg                  *sync.WaitGroup
}

func NewTestSubscriber() *TestSubscriber {
	return &TestSubscriber{
		notificationCounter: 0,
		wg:                  &sync.WaitGroup{},
	}
}

func (s *TestSubscriber) NotifyChangeState(events.DataConnectChangeNotif) error {
	s.notificationCounter++
	s.wg.Done()
	return nil
}

func (s *TestSubscriber) ExpectEvents(count int) {
	s.wg.Add(count)
}

type testFixture struct {
	sut            *ConnectionInfo
	subscriber     *TestSubscriber
	internalEvents *vpn.Events
	done           chan struct{}
}

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()

	s := NewTestSubscriber()
	sut := NewConnectionInfo()
	events := vpn.NewInternalVPNEvents()

	events.Subscribe(sut)
	sut.Subscribe(s)

	return &testFixture{
		sut:            sut,
		subscriber:     s,
		internalEvents: events,
		done:           make(chan struct{}),
	}
}

func (f *testFixture) waitForCompletion(t *testing.T) {
	t.Helper()
	select {
	case <-f.done:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscriber notifications")
	}
}

func TestConnectionInfo_InternalEventShallBePublishedWhenStateEventsAreReceived(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(2)

	go func() {
		tf.internalEvents.Connected.Publish(events.DataConnect{})
		tf.internalEvents.Disconnected.Publish(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)
	assert.Equal(t, tf.subscriber.notificationCounter, 2)
}

func TestConnectionInfo_VerifyDataConnectConversionToConnectionStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	event := events.DataConnect{
		Technology:              config.Technology_OPENVPN,
		Protocol:                config.Protocol_UDP,
		IP:                      netip.MustParseAddr("192.168.1.1"),
		Name:                    "server1",
		Hostname:                "42.example.pl",
		EventStatus:             events.StatusSuccess,
		TargetServerCountry:     "Test Country",
		TargetServerCountryCode: "TC",
		TargetServerCity:        "Test City",
		StartTime:               nil,
		IsVirtualLocation:       true,
		IsPostQuantum:           false,
		IsObfuscated:            true,
		TunnelName:              "tun0",
		IsMeshnetPeer:           false,
	}

	go func() {
		tf.internalEvents.Connected.Publish(event)
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)

	status := tf.sut.Status()
	assert.Equal(t, pb.ConnectionState_CONNECTED, status.State)
	assert.Equal(t, event.Technology, status.Technology)
	assert.Equal(t, event.Protocol, status.Protocol)
	assert.Equal(t, event.IP, status.IP)
	assert.Equal(t, event.Name, status.Name)
	assert.Equal(t, event.Hostname, status.Hostname)
	assert.Equal(t, event.TargetServerCountry, status.Country)
	assert.Equal(t, event.TargetServerCountryCode, status.CountryCode)
	assert.Equal(t, event.TargetServerCity, status.City)
	assert.Equal(t, event.IsVirtualLocation, status.VirtualLocation)
	assert.Equal(t, event.IsPostQuantum, status.PostQuantum)
	assert.Equal(t, event.IsObfuscated, status.Obfuscated)
	assert.Equal(t, event.TunnelName, status.TunnelName)
	assert.Equal(t, event.IsMeshnetPeer, status.MeshnetPeer)
}

func TestConnectionInfo_VerifyDataDisconnectConversionToConnectionStatus(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	go func() {
		tf.internalEvents.Disconnected.Publish(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)

	status := tf.sut.Status()
	assert.Equal(t, pb.ConnectionState_DISCONNECTED, status.State)
	assert.Nil(t, status.StartTime)
}

func TestConnectionInfo_TracksStateProperlyOnSuccess(t *testing.T) {
	category.Set(t, category.Unit)

	tf := newTestFixture(t)
	tf.subscriber.ExpectEvents(1)

	go func() {
		assert.Equal(t, pb.ConnectionState_UNKNOWN_STATE, tf.sut.Status().State)

		event := events.DataConnect{EventStatus: events.StatusSuccess}
		tf.internalEvents.Connected.Publish(event)
		tf.subscriber.wg.Wait()

		assert.Equal(t, pb.ConnectionState_CONNECTED, tf.sut.Status().State)

		tf.subscriber.ExpectEvents(1)
		tf.internalEvents.Disconnected.Publish(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)
	assert.Equal(t, pb.ConnectionState_DISCONNECTED, tf.sut.Status().State)
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
			StartTime:   &time.Time{},
		}
		tf.internalEvents.Connected.Publish(event)
		tf.subscriber.wg.Wait()

		status = tf.sut.Status()
		assert.NotNil(t, status.StartTime)

		tf.subscriber.ExpectEvents(1)
		tf.internalEvents.Disconnected.Publish(events.DataDisconnect{})
		tf.subscriber.wg.Wait()
		close(tf.done)
	}()

	tf.waitForCompletion(t)

	status := tf.sut.Status()
	assert.Nil(t, status.StartTime)
}
