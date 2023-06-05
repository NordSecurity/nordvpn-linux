package cli

import (
	"bytes"
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/client/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

type mockDaemonClient struct {
	pb.DaemonClient
}

func (c mockDaemonClient) Cities(ctx context.Context, in *pb.CitiesRequest, opts ...grpc.CallOption) (*pb.Payload, error) {
	x := &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{"Paris", "Madrid", "Atlanta", "Chicago", "Los_Angeles", "Miami", "New_York"},
	}
	return x, nil
}
func (c mockDaemonClient) Countries(ctx context.Context, in *pb.CountriesRequest, opts ...grpc.CallOption) (*pb.Payload, error) {
	x := &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{"Canada", "France", "Germany", "Hong_Kong", "Italy", "Japan", "Netherlands", "Poland", "Singapore", "Spain", "Sweden", "Switzerland", "Spain", "Turkey", "United_Arab_Emirates", "United_Kingdom", "United_States"},
	}
	return x, nil
}
func (c mockDaemonClient) Groups(ctx context.Context, in *pb.GroupsRequest, opts ...grpc.CallOption) (*pb.Payload, error) {
	x := &pb.Payload{
		Type: internal.CodeSuccess,
		Data: []string{"Africa_The_Middle_East_And_India", "Asia_Pacific", "Europe", "Obfuscated_Servers", "The_Americas"},
	}
	return x, nil
}

func TestConnectAutoComplete(t *testing.T) {
	category.Set(t, category.Unit)
	c := cmd{mockDaemonClient{}, nil, nil, "", nil, config.Config{}, nil}
	tests := []struct {
		name     string
		expected []string
		input    []string
	}{
		{
			name:     "France",
			expected: []string{"Paris"},
			input:    []string{"France"},
		},
		{
			name:     "Spain",
			expected: []string{"Madrid"},
			input:    []string{"Spain"},
		},
		{
			name:     "United_States",
			expected: []string{"Atlanta", "Chicago", "Los_Angeles", "Miami", "New_York"},
			input:    []string{"United_States"},
		},
		{
			name:     "Groups and Countries",
			expected: []string{"Africa_The_Middle_East_And_India", "Asia_Pacific", "Europe", "Obfuscated_Servers", "The_Americas", "Canada", "France", "Germany", "Hong_Kong", "Italy", "Japan", "Netherlands", "Poland", "Singapore", "Spain", "Sweden", "Switzerland", "Spain", "Turkey", "United_Arab_Emirates", "United_Kingdom", "United_States"},
			input:    []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			set.Parse(test.input)
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})
			var output bytes.Buffer
			c.ConnectAutoComplete(ctx)
			log.SetOutput(&output)
			defer log.SetOutput(os.Stdout)
			list, _ := internal.Columns(test.expected)
			assert.Contains(t, output.String(), list)
		})
	}
}
