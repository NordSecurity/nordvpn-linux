package daemon

import (
	"context"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/NordSecurity/nordvpn-linux/test/mock/networker"
	"github.com/stretchr/testify/assert"
)

type setVPNProtocolEnv struct {
	rpc       *RPC
	cm        *mock.ConfigManager
	netw      *networker.Mock
	published *daemonevents.MockPublisherSubscriber[config.VPNProtocol]
}

func newSetVPNProtocolEnv(current config.VPNProtocol) setVPNProtocolEnv {
	cm := mock.NewMockConfigManager()
	cm.Cfg.VPNProtocol = current

	netw := &networker.Mock{}
	published := &daemonevents.MockPublisherSubscriber[config.VPNProtocol]{}
	evts := daemonevents.NewEventsEmpty()
	evts.Settings.VPNProtocol = published

	return setVPNProtocolEnv{
		rpc: &RPC{
			remoteConfigGetter: mock.NewRemoteConfigMock(),
			cm:                 cm,
			netw:               netw,
			factory:            func(config.Technology) (vpn.VPN, error) { return nil, nil },
			events:             evts,
		},
		cm:        cm,
		netw:      netw,
		published: published,
	}
}

func (e setVPNProtocolEnv) call(t *testing.T, target config.VPNProtocol) *pb.Payload {
	t.Helper()
	resp, err := e.rpc.SetVPNProtocol(context.Background(), &pb.SetVPNProtocolRequest{VpnProtocol: target})
	assert.NoError(t, err, "SetVPNProtocol must report outcomes in-band, not as a gRPC error")
	return resp
}

func (e setVPNProtocolEnv) assertUnchanged(t *testing.T, current config.VPNProtocol, msg string) {
	t.Helper()
	assert.Equal(t, current, e.cm.Cfg.VPNProtocol, "%s: VPN protocol must not change", msg)
	assert.Equal(t, 0, e.cm.SaveCallCount, "%s: config must not be saved", msg)
	assert.False(t, e.published.EventPublished, "%s: no event must be published", msg)
}

func TestSetVPNProtocol_UnavailableProtocol_ReturnsFeatureHidden(t *testing.T) {
	category.Set(t, category.Unit)

	targets := []config.VPNProtocol{
		config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED,
		config.VPNProtocol(99),
	}
	if !features.NordWhisperEnabled {
		targets = append(targets, config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER)
	}

	for _, target := range targets {
		env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
		resp := env.call(t, target)

		msg := fmt.Sprintf("target %v", target)
		assert.Equal(t, internal.CodeFeatureHidden, resp.Type, msg)
		env.assertUnchanged(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, msg)
	}
}

func TestSetVPNProtocol_ConfigLoadError_ReturnsConfigError(t *testing.T) {
	category.Set(t, category.Unit)

	env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
	env.cm.LoadErr = fmt.Errorf("load failed")

	resp := env.call(t, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP)

	assert.Equal(t, internal.CodeConfigError, resp.Type)
	env.assertUnchanged(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, "load error")
}

func TestSetVPNProtocol_AlreadySet_ReturnsNothingToDo(t *testing.T) {
	category.Set(t, category.Unit)

	for _, current := range config.AvailableVPNProtocols() {
		env := newSetVPNProtocolEnv(current)
		resp := env.call(t, current)

		msg := fmt.Sprintf("current %v", current)
		assert.Equal(t, internal.CodeNothingToDo, resp.Type, msg)
		assert.Equal(t, []string{current.DisplayName()}, resp.Data, msg)
		env.assertUnchanged(t, current, msg)
	}
}

func TestSetVPNProtocol_Success_SavesPublishesAndReportsReconnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		current      config.VPNProtocol
		target       config.VPNProtocol
		vpnActive    bool
		expectedCode int64
	}{
		{
			name:         "NordLynx to OpenVPN UDP, VPN off",
			current:      config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			target:       config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
			expectedCode: internal.CodeSuccess,
		},
		{
			name:         "OpenVPN UDP to OpenVPN TCP, VPN off",
			current:      config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
			target:       config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
			expectedCode: internal.CodeSuccess,
		},
		{
			name:         "OpenVPN TCP to OpenVPN UDP, VPN on",
			current:      config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
			target:       config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
			vpnActive:    true,
			expectedCode: internal.CodeSuccessReconnectRequired,
		},
		{
			name:         "OpenVPN TCP to NordLynx, VPN on",
			current:      config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
			target:       config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			vpnActive:    true,
			expectedCode: internal.CodeSuccessReconnectRequired,
		},
	}
	if features.NordWhisperEnabled {
		tests = append(tests, struct {
			name         string
			current      config.VPNProtocol
			target       config.VPNProtocol
			vpnActive    bool
			expectedCode int64
		}{
			name:         "NordLynx to NordWhisper, VPN off",
			current:      config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			target:       config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
			expectedCode: internal.CodeSuccess,
		})
	}

	for _, test := range tests {
		env := newSetVPNProtocolEnv(test.current)
		env.netw.VpnActive = test.vpnActive

		resp := env.call(t, test.target)

		assert.Equal(t, test.expectedCode, resp.Type, test.name)
		assert.Equal(t, []string{test.target.DisplayName()}, resp.Data, test.name)
		assert.Equal(t, test.target, env.cm.Cfg.VPNProtocol, "%s: VPN protocol must be saved", test.name)
		assert.True(t, env.published.EventPublished, "%s: event must be published", test.name)
		assert.Equal(t, test.target, env.published.Event, "%s: published event must carry the target", test.name)
	}
}

func TestSetVPNProtocol_DedicatedServerAutoconnect_RejectsNonNordLynx(t *testing.T) {
	category.Set(t, category.Unit)

	targets := []config.VPNProtocol{
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
	}
	if features.NordWhisperEnabled {
		targets = append(targets, config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER)
	}

	for _, target := range targets {
		env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
		env.cm.Cfg.AutoConnect = true
		env.cm.Cfg.AutoConnectData.Group = config.ServerGroup_DEDICATED_SERVER

		resp := env.call(t, target)

		msg := fmt.Sprintf("target %v", target)
		assert.Equal(t, internal.CodeDedicatedServersNoNordlynx, resp.Type, msg)
		env.assertUnchanged(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, msg)
	}
}

func TestSetVPNProtocol_DedicatedServerAutoconnect_AllowsNordLynx(t *testing.T) {
	category.Set(t, category.Unit)

	env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP)
	env.cm.Cfg.AutoConnect = true
	env.cm.Cfg.AutoConnectData.Group = config.ServerGroup_DEDICATED_SERVER

	resp := env.call(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)

	assert.Equal(t, internal.CodeSuccess, resp.Type)
	assert.Equal(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, env.cm.Cfg.VPNProtocol)
}

func TestSetVPNProtocol_ObfuscatedAutoconnect_RejectsNonNordWhisper(t *testing.T) {
	category.Set(t, category.Unit)

	targets := []config.VPNProtocol{
		config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
	}

	for _, target := range targets {
		current := config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER
		env := newSetVPNProtocolEnv(current)
		env.cm.Cfg.AutoConnect = true
		env.cm.Cfg.AutoConnectData.Group = config.ServerGroup_NW_OBFUSCATED

		resp := env.call(t, target)

		msg := fmt.Sprintf("target %v", target)
		assert.Equal(t, internal.CodeObfuscatedNeedsNordwhisper, resp.Type, msg)
		env.assertUnchanged(t, current, msg)
	}
}

func TestSetVPNProtocol_PostQuantumEnabled_RejectsNonNordLynx(t *testing.T) {
	category.Set(t, category.Unit)

	targets := []config.VPNProtocol{
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
	}
	if features.NordWhisperEnabled {
		targets = append(targets, config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER)
	}

	for _, target := range targets {
		env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
		env.cm.Cfg.AutoConnectData.PostquantumVpn = true

		resp := env.call(t, target)

		msg := fmt.Sprintf("target %v", target)
		assert.Equal(t, internal.CodePqWithoutNordlynx, resp.Type, msg)
		assert.Equal(t, []string{target.DisplayName()}, resp.Data, msg)
		env.assertUnchanged(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, msg)
	}
}

func TestSetVPNProtocol_VPNFactoryError_ReturnsConfigError(t *testing.T) {
	category.Set(t, category.Unit)

	env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
	env.rpc.factory = func(config.Technology) (vpn.VPN, error) { return nil, fmt.Errorf("no such technology") }

	resp := env.call(t, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP)

	assert.Equal(t, internal.CodeConfigError, resp.Type)
	env.assertUnchanged(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, "factory error")
}

func TestSetVPNProtocol_ConfigSaveError_ReturnsConfigError(t *testing.T) {
	category.Set(t, category.Unit)

	env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
	env.cm.SaveErr = fmt.Errorf("save failed")

	resp := env.call(t, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP)

	assert.Equal(t, internal.CodeConfigError, resp.Type)
	assert.Equal(t, config.VPNProtocol_VPN_PROTOCOL_NORDLYNX, env.cm.Cfg.VPNProtocol)
	assert.False(t, env.published.EventPublished, "no event must be published when saving fails")
}

type remoteECHDefaultMock struct {
	*mock.RemoteConfigMock
	echDefault string
}

func (r remoteECHDefaultMock) GetFeatureParam(_, _ string) (string, error) { return r.echDefault, nil }

func TestSetVPNProtocol_ECHEnabled_ResetsToRemoteDefaultWhenTargetIsNotNordWhisper(t *testing.T) {
	category.Set(t, category.Unit)

	targets := []config.VPNProtocol{
		config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP,
	}

	for _, target := range targets {
		env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER)
		env.rpc.remoteConfigGetter = remoteECHDefaultMock{RemoteConfigMock: mock.NewRemoteConfigMock(), echDefault: "false"}
		env.cm.Cfg.AutoConnectData.ECH.Set(true)

		resp := env.call(t, target)

		msg := fmt.Sprintf("target %v", target)
		assert.Equal(t, internal.CodeSuccess, resp.Type, msg)
		assert.False(t, env.cm.Cfg.AutoConnectData.ECH.Get(), "%s: enabled ECH must be reset to the remote default", msg)
	}
}

func TestSetVPNProtocol_ECHDisabled_StaysDisabled(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		current config.VPNProtocol
		target  config.VPNProtocol
	}{
		{
			name:    "leaving NordWhisper for NordLynx",
			current: config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
			target:  config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
		},
		{
			name:    "leaving NordWhisper for OpenVPN",
			current: config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER,
			target:  config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		},
		{
			name:    "switching between non-NordWhisper protocols",
			current: config.VPNProtocol_VPN_PROTOCOL_NORDLYNX,
			target:  config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP,
		},
	}

	for _, test := range tests {
		env := newSetVPNProtocolEnv(test.current)
		env.cm.Cfg.AutoConnectData.ECH.Set(false)

		resp := env.call(t, test.target)

		assert.Equal(t, internal.CodeSuccess, resp.Type, test.name)
		assert.False(t, env.cm.Cfg.AutoConnectData.ECH.Get(), "%s: disabled ECH must not be reset", test.name)
	}
}

func TestSetVPNProtocol_ToNordWhisper_KeepsECH(t *testing.T) {
	category.Set(t, category.Unit)
	if !features.NordWhisperEnabled {
		t.Skip("NordWhisper is not available in this build")
	}

	for _, stored := range []bool{true, false} {
		env := newSetVPNProtocolEnv(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX)
		env.rpc.remoteConfigGetter = remoteECHDefaultMock{RemoteConfigMock: mock.NewRemoteConfigMock(), echDefault: fmt.Sprint(!stored)}
		env.cm.Cfg.AutoConnectData.ECH.Set(stored)

		resp := env.call(t, config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER)

		msg := fmt.Sprintf("stored ECH %v", stored)
		assert.Equal(t, internal.CodeSuccess, resp.Type, msg)
		assert.Equal(t, stored, env.cm.Cfg.AutoConnectData.ECH.Get(), "%s: ECH must be kept when switching to NordWhisper", msg)
	}
}
