package cli

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestGroupsList(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mockDaemonClient{}
	c := cmd{&mockClient, nil, nil, nil, "", nil}

	tests := []struct {
		name          string
		groups        []*pb.ServerGroup
		expected      string
		input         string
		expectedError error
	}{
		{
			name:          "error response",
			expectedError: formatError(fmt.Errorf(MsgListIsEmpty, "server groups")),
		},
		{
			name:     "groups list",
			expected: "P2P\nDouble_VPN",
			groups:   []*pb.ServerGroup{{Name: "P2P", VirtualLocation: false}, {Name: "Double_VPN", VirtualLocation: false}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			mockClient.groups = test.groups
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := captureOutput(func() {
				err := c.Groups(ctx)
				assert.Equal(t, test.expectedError, err)
			})
			assert.Nil(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}
