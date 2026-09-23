//go:build moose

package moose

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	moose "moose/events"

	"gotest.tools/v3/assert"
)

type techContextRecorder struct {
	userPrefTech  []moose.NordvpnappVpnConnectionTechnology
	currentTech   []moose.NordvpnappVpnConnectionTechnology
	userPrefProto []moose.NordvpnappVpnConnectionProtocol
	currentProto  []moose.NordvpnappVpnConnectionProtocol
	obfuscation   []bool
}

func newTechTestSubscriber() (*Subscriber, *techContextRecorder) {
	sub := NewSubscriber("", nil, nil, nil, config.BuildTarget{}, "", "", "")
	noopDisconnectAmbientMooseFuncs(sub)
	sub.mooseFuncs.setTPLiteCurrentState = func(_ bool) uint32 { return 0 }
	sub.mooseFuncs.sendConnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.TargetConnectionAdditionalParams,
		_ moose.ConnectionParams,
		_ moose.NordvpnappOptBool,
		_ int32,
		_ string,
		_ *string,
	) uint32 {
		return 0
	}
	sub.mooseFuncs.sendDisconnect = func(
		_ moose.EventParams,
		_ moose.TargetConnectionParams,
		_ moose.ConnectionParams,
		_ int32,
		_ int32,
		_ *string,
	) uint32 {
		return 0
	}

	r := &techContextRecorder{}
	sub.mooseFuncs.setTechnologyUserPreference = func(v moose.NordvpnappVpnConnectionTechnology) uint32 {
		r.userPrefTech = append(r.userPrefTech, v)
		return 0
	}
	sub.mooseFuncs.setTechnologyCurrentState = func(v moose.NordvpnappVpnConnectionTechnology) uint32 {
		r.currentTech = append(r.currentTech, v)
		return 0
	}
	sub.mooseFuncs.setProtocolUserPreference = func(v moose.NordvpnappVpnConnectionProtocol) uint32 {
		r.userPrefProto = append(r.userPrefProto, v)
		return 0
	}
	sub.mooseFuncs.setProtocolCurrentState = func(v moose.NordvpnappVpnConnectionProtocol) uint32 {
		r.currentProto = append(r.currentProto, v)
		return 0
	}
	sub.mooseFuncs.setObfuscationEnabledUserPreference = func(v bool) uint32 {
		r.obfuscation = append(r.obfuscation, v)
		return 0
	}
	return sub, r
}

func connectVPN(t *testing.T, sub *Subscriber, tech config.Technology, proto config.Protocol) {
	t.Helper()
	assert.NilError(t, sub.NotifyConnect(events.DataConnect{
		EventStatus: events.StatusSuccess,
		Technology:  tech,
		Protocol:    proto,
	}))
}

func TestNotifyTechnology_Disconnected_ReportsEffectiveState(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		technology config.Technology
		obfuscated bool
	}{
		{name: "nordlynx is not obfuscated", technology: config.Technology_NORDLYNX, obfuscated: false},
		{name: "openvpn is not obfuscated", technology: config.Technology_OPENVPN, obfuscated: false},
		{name: "nordwhisper is obfuscated", technology: config.Technology_NORDWHISPER, obfuscated: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, rec := newTechTestSubscriber()

			err := sub.NotifyTechnology(tt.technology)

			assert.NilError(t, err)
			want := connectionTechnologyToInternalType(tt.technology)
			assert.DeepEqual(t, []moose.NordvpnappVpnConnectionTechnology{want}, rec.userPrefTech)
			assert.DeepEqual(t, []moose.NordvpnappVpnConnectionTechnology{want}, rec.currentTech)
			assert.DeepEqual(t, []bool{tt.obfuscated}, rec.obfuscation)
			assert.Equal(t, 0, len(rec.currentProto))
		})
	}
}

func TestNotifyTechnology_Unknown_ReturnsError(t *testing.T) {
	category.Set(t, category.Unit)
	sub, rec := newTechTestSubscriber()

	err := sub.NotifyTechnology(config.Technology_UNKNOWN_TECHNOLOGY)

	assert.ErrorIs(t, err, errUnknownTechnology)
	assert.Equal(t, 0, len(rec.userPrefTech))
	assert.Equal(t, 0, len(rec.currentTech))
}

func TestNotifyProtocol_Disconnected_ReportsCurrentState(t *testing.T) {
	category.Set(t, category.Unit)
	sub, rec := newTechTestSubscriber()

	err := sub.NotifyProtocol(config.Protocol_TCP)

	assert.NilError(t, err)
	want := connectionProtocolToInternalType(config.Protocol_TCP)
	assert.DeepEqual(t, []moose.NordvpnappVpnConnectionProtocol{want}, rec.userPrefProto)
	assert.DeepEqual(t, []moose.NordvpnappVpnConnectionProtocol{want}, rec.currentProto)
	assert.Equal(t, 0, len(rec.currentTech))
	assert.Equal(t, 0, len(rec.obfuscation))
}

func TestNotifyConnect_NotSuccessful_DoesNotTouchConnectionTechnology(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		data events.DataConnect
	}{
		{
			name: "attempt",
			data: events.DataConnect{EventStatus: events.StatusAttempt, Technology: config.Technology_NORDWHISPER},
		},
		{
			name: "failure",
			data: events.DataConnect{EventStatus: events.StatusFailure, Technology: config.Technology_NORDWHISPER},
		},
		{
			name: "meshnet peer success",
			data: events.DataConnect{EventStatus: events.StatusSuccess, IsMeshnetPeer: true, Technology: config.Technology_NORDWHISPER},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, rec := newTechTestSubscriber()

			err := sub.NotifyConnect(tt.data)

			assert.NilError(t, err)
			assert.Equal(t, 0, len(rec.currentTech))
			assert.Equal(t, 0, len(rec.currentProto))
			assert.Equal(t, 0, len(rec.obfuscation))
		})
	}
}

func TestNotifyTechnology_MeshnetPeerOnly_IsNotDeferred(t *testing.T) {
	category.Set(t, category.Unit)
	sub, rec := newTechTestSubscriber()
	assert.NilError(t, sub.NotifyConnect(events.DataConnect{EventStatus: events.StatusSuccess, IsMeshnetPeer: true}))

	assert.NilError(t, sub.NotifyTechnology(config.Technology_NORDWHISPER))

	assert.DeepEqual(t, []bool{true}, rec.obfuscation)
	assert.DeepEqual(t,
		[]moose.NordvpnappVpnConnectionTechnology{connectionTechnologyToInternalType(config.Technology_NORDWHISPER)},
		rec.currentTech)

	assert.NilError(t, sub.NotifyDisconnect(events.DataDisconnect{EventStatus: events.StatusSuccess}))
	assert.DeepEqual(t, []bool{true}, rec.obfuscation)
}

func TestTechnologyChangeWhileConnected_ReportedOnNextConnect(t *testing.T) {
	category.Set(t, category.Unit)
	sub, rec := newTechTestSubscriber()
	udp := connectionProtocolToInternalType(config.Protocol_UDP)
	webtunnel := connectionProtocolToInternalType(config.Protocol_Webtunnel)

	assert.NilError(t, sub.NotifyProtocol(config.Protocol_UDP))
	assert.NilError(t, sub.NotifyTechnology(config.Technology_OPENVPN))
	assert.DeepEqual(t, []bool{false}, rec.obfuscation)

	connectVPN(t, sub, config.Technology_OPENVPN, config.Protocol_UDP)
	assert.DeepEqual(t, []bool{false, false}, rec.obfuscation)

	assert.NilError(t, sub.NotifyTechnology(config.Technology_NORDWHISPER))
	assert.NilError(t, sub.NotifyProtocol(config.Protocol_Webtunnel))
	assert.DeepEqual(t, []bool{false, false}, rec.obfuscation)
	assert.DeepEqual(t, []moose.NordvpnappVpnConnectionProtocol{udp, udp, udp}, rec.currentProto)
	assert.Equal(t, 2, len(rec.userPrefTech))
	assert.Equal(t, 2, len(rec.userPrefProto))

	assert.NilError(t, sub.NotifyDisconnect(events.DataDisconnect{EventStatus: events.StatusSuccess}))
	assert.DeepEqual(t, []bool{false, false, true}, rec.obfuscation)

	connectVPN(t, sub, config.Technology_NORDWHISPER, config.Protocol_Webtunnel)
	assert.DeepEqual(t, []bool{false, false, true, true}, rec.obfuscation)
	assert.DeepEqual(t, []moose.NordvpnappVpnConnectionProtocol{udp, udp, udp, webtunnel, webtunnel}, rec.currentProto)
}
