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

func (s *TestSubscriber) NotifyChangeState(e events.DataConnectChangeNotif) error {
	s.notificationCounter++
	s.wg.Done()
	return nil
}

func NewTestSubscriber(wg *sync.WaitGroup) *TestSubscriber {
	return &TestSubscriber{
		notificationCounter: 0,
		wg:                  wg,
	}
}

func TestConnectionInfo_InternalEventShallBePublishedWhenStateEventsAreReceived(t *testing.T) {
	category.Set(t, category.Unit)

	var wg sync.WaitGroup
	wg.Add(2)
	s := NewTestSubscriber(&wg)
	sut := NewConnectionInfo()
	internalVpnEvents := vpn.NewInternalVPNEvents()
	internalVpnEvents.Subscribe(sut)
	sut.Subscribe(s)

	internalVpnEvents.Connected.Publish(events.DataConnect{})
	internalVpnEvents.Disconnected.Publish(events.DataDisconnect{})
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscriber notifications")
	}
	assert.Equal(t, s.notificationCounter, 2)
}

func TestConnectionInfo_VerifyDataConnectConversionToConnectionStatus(t *testing.T) {
	category.Set(t, category.Unit)

	var wg sync.WaitGroup
	wg.Add(1)
	s := NewTestSubscriber(&wg)
	sut := NewConnectionInfo()
	internalVpnEvents := vpn.NewInternalVPNEvents()
	internalVpnEvents.Subscribe(sut)
	sut.Subscribe(s)

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
	internalVpnEvents.Connected.Publish(event)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscriber notifications")
	}

	status := sut.Status()
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

	var wg sync.WaitGroup
	wg.Add(1)
	s := NewTestSubscriber(&wg)
	sut := NewConnectionInfo()
	internalVpnEvents := vpn.NewInternalVPNEvents()
	internalVpnEvents.Subscribe(sut)
	sut.Subscribe(s)

	internalVpnEvents.Disconnected.Publish(events.DataDisconnect{})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscriber notifications")
	}

	status := sut.Status()
	assert.Equal(t, pb.ConnectionState_DISCONNECTED, status.State)
	assert.Nil(t, status.StartTime)
}
