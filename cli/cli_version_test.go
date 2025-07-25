package cli

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

// TestVersion checks that the cli.Version command adds when needed the outdated information to the version
func TestVersion(t *testing.T) {
	category.Set(t, category.Unit)

	mockClient := mock.MockDaemonClient{}
	c := cmd{&mockClient, nil, nil, "", nil}
	// this is constructed using composeAppVersion,
	// covered in other tests and is expected to be set in app.Version field
	const appVersion = "1.2.3"

	tests := []struct {
		name      string
		outdated  bool
		pingError error
		expected  string
	}{
		{
			name:     "App version",
			outdated: false,
		},
		{
			name:     "Application is outdated",
			outdated: true,
		},
		{
			name:      "Ping returns error",
			outdated:  true,
			pingError: fmt.Errorf("some error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			app.Version = appVersion
			set := flag.NewFlagSet("test", 0)
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})
			mockClient.PingFn = func() (*pb.PingResponse, error) {
				typeResponse := internal.CodeSuccess
				if test.outdated {
					typeResponse = internal.CodeOutdated
				}
				// only the Type member is used from the response,
				// because of this set the version to something different than the appVersion
				// string to be sure is not used in the output
				return &pb.PingResponse{
					Major:    5,
					Minor:    6,
					Patch:    7,
					Metadata: "456",
					Type:     typeResponse,
				}, test.pingError
			}

			result, err := helpers.CaptureOutput(func() {
				assert.Nil(t, c.Version(ctx))
			})
			assert.Nil(t, err)

			// compose the app version message taking outdated value into account
			expectedVersion := "NordVPN Version " + appVersion
			if test.outdated && test.pingError == nil {
				expectedVersion = expectedVersion + " (outdated)"
			}
			assert.Equal(t, expectedVersion, result)
		})
	}
}
