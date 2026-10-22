package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type countingSettingsClient struct {
	pb.DaemonClient
	calls       int
	vpnProtocol config.VPNProtocol
	err         error
}

func (c *countingSettingsClient) Settings(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.SettingsResponse, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	return &pb.SettingsResponse{
		Data: &pb.Settings{VpnProtocol: c.vpnProtocol},
	}, nil
}

func TestExceptMemoizesSettings(t *testing.T) {
	category.Set(t, category.Unit)

	client := &countingSettingsClient{vpnProtocol: config.VPNProtocol_VPN_PROTOCOL_NORDLYNX}
	c := &cmd{client: client}

	assert.False(t, c.Except(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX), "matching protocol is not excepted")
	assert.True(t, c.Except(config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP), "differing protocol is excepted")
	assert.True(t, c.Except(config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER), "differing protocol is excepted")

	assert.Equal(t, 1, client.calls, "Settings must be fetched once and memoized")
}

func TestExceptReturnsFalseOnSettingsError(t *testing.T) {
	category.Set(t, category.Unit)

	client := &countingSettingsClient{err: errors.New("daemon unavailable")}
	c := &cmd{client: client}

	assert.False(t, c.Except(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX))
	assert.False(t, c.Except(config.VPNProtocol_VPN_PROTOCOL_NORDLYNX))
	assert.Equal(t, 2, client.calls, "failed fetches must not be memoized")
}
