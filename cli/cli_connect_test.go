package cli

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

func captureOutput(f func()) (string, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	os.Stdout = writer
	os.Stderr = writer

	f()

	writer.Close() // close to unblock io.Copy(&buf, reader)
	var buf bytes.Buffer
	io.Copy(&buf, reader)
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

type mockDaemonClient struct {
	pb.DaemonClient
	cities    []*pb.ServerGroup
	groups    []*pb.ServerGroup
	countries []*pb.ServerGroup
}

func (c mockDaemonClient) Cities(ctx context.Context, in *pb.CitiesRequest, opts ...grpc.CallOption) (*pb.ServerGroupsList, error) {
	if c.cities != nil {
		return &pb.ServerGroupsList{
			Type:    internal.CodeSuccess,
			Servers: c.cities,
		}, nil
	} else {
		return &pb.ServerGroupsList{
			Type:    internal.CodeEmptyPayloadError,
			Servers: nil,
		}, nil
	}
}

func (c mockDaemonClient) Countries(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ServerGroupsList, error) {
	if c.countries != nil {
		return &pb.ServerGroupsList{
			Type:    internal.CodeSuccess,
			Servers: c.countries,
		}, nil
	} else {
		return &pb.ServerGroupsList{
			Type:    internal.CodeEmptyPayloadError,
			Servers: nil,
		}, nil
	}
}
func (c mockDaemonClient) Groups(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ServerGroupsList, error) {
	if c.groups != nil {
		return &pb.ServerGroupsList{
			Type:    internal.CodeSuccess,
			Servers: c.groups,
		}, nil
	} else {
		return &pb.ServerGroupsList{
			Type:    internal.CodeEmptyPayloadError,
			Servers: nil,
		}, nil
	}
}

func TestConnectAutoComplete(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mockDaemonClient{}
	c := cmd{&mockClient, nil, nil, nil, "", nil}
	tests := []struct {
		name      string
		countries []*pb.ServerGroup
		groups    []*pb.ServerGroup
		cities    []*pb.ServerGroup
		expected  string
		input     string
	}{
		{
			name:      "autocomplete without input returns groups first plus countries second",
			countries: []*pb.ServerGroup{{Name: "France", VirtualLocation: false}},
			cities:    []*pb.ServerGroup{{Name: "Paris", VirtualLocation: false}},
			groups:    []*pb.ServerGroup{{Name: "P2P", VirtualLocation: false}},
			expected:  "P2P\nFrance",
		},
		{
			name:      "cities autocomplete works with country name",
			countries: []*pb.ServerGroup{{Name: "France", VirtualLocation: false}},
			cities:    []*pb.ServerGroup{{Name: "Paris", VirtualLocation: false}},
			groups:    []*pb.ServerGroup{{Name: "P2P", VirtualLocation: false}},
			expected:  "Paris",
			input:     "FrAnCe",
		},
		{
			name:      "cities autocomplete works with country code",
			cities:    []*pb.ServerGroup{{Name: "Paris", VirtualLocation: false}},
			countries: []*pb.ServerGroup{{Name: "France", VirtualLocation: false}},
			groups:    []*pb.ServerGroup{{Name: "P2P", VirtualLocation: false}},
			expected:  "Paris",
			input:     "fR",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			mockClient.cities = test.cities
			mockClient.countries = test.countries
			mockClient.groups = test.groups
			set.Parse([]string{test.input})
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := captureOutput(func() {
				c.ConnectAutoComplete(ctx)
			})

			assert.Nil(t, err)

			assert.Equal(t, test.expected, result)
		})
	}
}
