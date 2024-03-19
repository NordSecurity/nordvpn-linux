package cli

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCountriesList(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mockDaemonClient{}
	c := cmd{&mockClient, nil, nil, "", nil}

	tests := []struct {
		name          string
		countries     []string
		expected      string
		input         string
		expectedError error
	}{
		{
			name:          "error response",
			expectedError: formatError(fmt.Errorf(MsgListIsEmpty, "countries")),
		},
		{
			name:      "countries list",
			expected:  "France, Germany",
			countries: []string{"France", "Germany"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			mockClient.countries = test.countries
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := captureOutput(func() {
				err := c.Countries(ctx)
				assert.Equal(t, test.expectedError, err)
			})
			assert.Nil(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}
