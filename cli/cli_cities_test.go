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

func TestCitiesList(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mockDaemonClient{}
	c := cmd{&mockClient, nil, nil, nil, "", nil}

	tests := []struct {
		name          string
		cities        []string
		country       string
		expected      string
		expectedError error
	}{
		{
			name:          "missing country name",
			expectedError: formatError(fmt.Errorf("The command you entered is not valid. Enter 'cli.test --help' to see the options.")),
		},
		{
			name:          "no cities data",
			country:       "France",
			expectedError: formatError(fmt.Errorf("We couldn’t load the list of cities. Please try again later.")),
		},
		{
			name:     "cities list",
			country:  "France",
			expected: "Marseille, Paris",
			cities:   []string{"Marseille", "Paris"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			if test.country != "" {
				set.Parse([]string{test.country})
			}
			mockClient.cities = test.cities
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := captureOutput(func() {
				err := c.Cities(ctx)
				assert.Equal(t, test.expectedError, err)
			})
			assert.Nil(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}
