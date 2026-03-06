package cli

import (
	"context"
	"flag"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestLoginWithToken_ValidToken(t *testing.T) {
	category.Set(t, category.Unit)

	loginCalled := false
	var capturedToken string
	mockClient := mock.MockDaemonClient{
		LoginWithTokenFn: func(ctx context.Context, in *pb.LoginWithTokenRequest) (*pb.LoginResponse, error) {
			loginCalled = true
			capturedToken = in.Token
			return &pb.LoginResponse{Type: internal.CodeSuccess}, nil
		},
	}

	c := &cmd{client: mockClient}
	app := cli.NewApp()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.Parse([]string{"abc123def456"})
	ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

	err := c.loginWithToken(ctx)

	assert.NoError(t, err)
	assert.True(t, loginCalled, "LoginWithToken should be called")
	assert.Equal(t, "abc123def456", capturedToken)
}

func TestLoginWithToken_NoArgument_NonTTY_ReturnsError(t *testing.T) {
	category.Set(t, category.Unit)

	loginCalled := false
	mockClient := mock.MockDaemonClient{
		LoginWithTokenFn: func(ctx context.Context, in *pb.LoginWithTokenRequest) (*pb.LoginResponse, error) {
			loginCalled = true
			return &pb.LoginResponse{Type: internal.CodeSuccess}, nil
		},
	}
	c := &cmd{client: mockClient}
	app := cli.NewApp()
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

	err := c.loginWithToken(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), client.TokenInputNotTerminal)
	assert.False(t, loginCalled, "LoginWithToken should not be called when interactive fails")
}
