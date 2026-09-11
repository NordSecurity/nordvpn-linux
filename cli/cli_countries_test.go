package cli

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCountriesList(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mock.MockDaemonClient{}
	c := cmd{client: &mockClient}

	tests := []struct {
		name          string
		countries     []*pb.ServerGroup
		expected      string
		input         string
		expectedError error
	}{
		{
			name:          "error message when countries list is empty",
			expectedError: formatError(fmt.Errorf(MsgListIsEmpty, "countries")),
		},
		{
			name:      "return all servers",
			expected:  "France\nGermany",
			countries: []*pb.ServerGroup{{Name: "France"}, {Name: "Germany"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			mockClient.CountriesResponse = test.countries
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := helpers.CaptureOutput(func() {
				err := c.Countries(ctx)
				assert.Equal(t, test.expectedError, err)
			})
			assert.Nil(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}
