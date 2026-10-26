package cli

import (
	"context"
	"flag"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestSetDefaults(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		resp        *pb.Payload
		expectedErr string
	}{
		{
			name: "all settings applied",
			resp: &pb.Payload{Type: internal.CodeSuccess},
		},
		{
			name: "failed settings return an error",
			resp: &pb.Payload{Type: internal.CodeSetDefaultsNotApplied, Data: []string{"routing", "firewall"}},
			expectedErr: "Some default settings could not be applied. The reset finished, but these settings may not be active: routing, firewall. " +
				"Run 'nordvpn set defaults' again to retry. If the problem continues, contact support.",
		},
		{
			name: "unknown failed settings return a generic error",
			resp: &pb.Payload{Type: internal.CodeSetDefaultsNotApplied},
			expectedErr: "Some default network settings could not be applied. Run 'nordvpn set defaults' again to retry. " +
				"If the problem continues, contact support.",
		},
		{
			name:        "config error returns an error",
			resp:        &pb.Payload{Type: internal.CodeConfigError},
			expectedErr: formatError(ErrConfig).Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := cmd{client: mock.MockDaemonClient{
				SetDefaultsFn: func(context.Context, *pb.SetDefaultsRequest) (*pb.Payload, error) {
					return test.resp, nil
				},
			}}
			ctx := cli.NewContext(cli.NewApp(), flag.NewFlagSet("defaults", flag.ContinueOnError), nil)

			err := c.SetDefaults(ctx)
			if test.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.expectedErr)
			}
		})
	}
}
