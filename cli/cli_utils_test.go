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
	calls      int
	technology config.Technology
	err        error
}

func (c *countingSettingsClient) Settings(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.SettingsResponse, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	return &pb.SettingsResponse{
		Data: &pb.Settings{Technology: c.technology},
	}, nil
}

func TestExceptMemoizesSettings(t *testing.T) {
	category.Set(t, category.Unit)

	client := &countingSettingsClient{technology: config.Technology_NORDLYNX}
	c := &cmd{client: client}

	assert.False(t, c.Except(config.Technology_NORDLYNX), "matching technology is not excepted")
	assert.True(t, c.Except(config.Technology_OPENVPN), "differing technology is excepted")
	assert.True(t, c.Except(config.Technology_NORDWHISPER), "differing technology is excepted")

	assert.Equal(t, 1, client.calls, "Settings must be fetched once and memoized")
}

func TestExceptReturnsFalseOnSettingsError(t *testing.T) {
	category.Set(t, category.Unit)

	client := &countingSettingsClient{err: errors.New("daemon unavailable")}
	c := &cmd{client: client}

	assert.False(t, c.Except(config.Technology_NORDLYNX))
	assert.False(t, c.Except(config.Technology_NORDLYNX))
	assert.Equal(t, 2, client.calls, "failed fetches must not be memoized")
}
