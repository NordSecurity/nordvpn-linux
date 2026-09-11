package cli

import (
	"context"
	"flag"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/helpers"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestAutoConnectAutoComplete(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mock.MockDaemonClient{}
	c := cmd{client: &mockClient}
	tests := []struct {
		name      string
		countries []*pb.ServerGroup
		groups    []*pb.ServerGroup
		cities    []*pb.ServerGroup
		expected  string
		input     []string
	}{
		{
			name:      "autocomplete without input returns groups first plus countries second",
			countries: []*pb.ServerGroup{{Name: "France"}},
			cities:    []*pb.ServerGroup{{Name: "Paris"}},
			groups:    []*pb.ServerGroup{{Name: "P2P"}},
			input:     []string{"1"},
			expected:  "P2P\nFrance",
		},
		{
			name:      "cities autocomplete works with country name",
			countries: []*pb.ServerGroup{{Name: "France"}},
			cities:    []*pb.ServerGroup{{Name: "Paris"}},
			groups:    []*pb.ServerGroup{{Name: "P2P"}},
			input:     []string{"1", "france"},
			expected:  "Paris",
		},
		{
			name:      "cities autocomplete works with country code",
			cities:    []*pb.ServerGroup{{Name: "Paris"}},
			countries: []*pb.ServerGroup{{Name: "France"}},
			groups:    []*pb.ServerGroup{{Name: "P2P"}},
			input:     []string{"1", "fR"},
			expected:  "Paris",
		},
		{
			name:      "no autocomplete suggestions for disabling auto connect",
			cities:    []*pb.ServerGroup{{Name: "Paris"}},
			countries: []*pb.ServerGroup{{Name: "France"}},
			groups:    []*pb.ServerGroup{{Name: "P2P"}},
			input:     []string{"0", "fR"},
			expected:  "",
		},
		{
			name:      "autocomplete works for groups",
			cities:    []*pb.ServerGroup{{Name: "Paris"}},
			countries: []*pb.ServerGroup{{Name: "France"}},
			groups:    []*pb.ServerGroup{{Name: "P2P"}},
			input:     []string{"1", "--group", "P2"},
			expected:  "P2P",
		},
		{
			name:      "give bool suggestions for no parameters",
			cities:    []*pb.ServerGroup{{Name: "Paris"}},
			countries: []*pb.ServerGroup{{Name: "France"}},
			groups:    []*pb.ServerGroup{{Name: "P2P"}},
			input:     []string{},
			expected:  "0\n1\ndisable\ndisabled\nenable\nenabled\nfalse\noff\non\ntrue",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			mockClient.CitiesResponse = test.cities
			mockClient.CountriesResponse = test.countries
			mockClient.GroupsResponse = test.groups
			set.Parse(test.input)
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := helpers.CaptureOutput(func() {
				c.SetAutoConnectAutoComplete(ctx)
			})

			assert.Nil(t, err)

			assert.Equal(t, test.expected, result)
		})
	}
}
