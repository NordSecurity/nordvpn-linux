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
	cities    []string
	groups    []string
	countries []string
}

func (c mockDaemonClient) Cities(ctx context.Context, in *pb.CitiesRequest, opts ...grpc.CallOption) (*pb.Payload, error) {
	if c.cities != nil {
		return &pb.Payload{
			Type: internal.CodeSuccess,
			Data: c.cities,
		}, nil
	} else {
		return &pb.Payload{
			Type: internal.CodeEmptyPayloadError,
			Data: nil,
		}, nil
	}
}

func (c mockDaemonClient) Countries(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.Payload, error) {
	if c.countries != nil {
		return &pb.Payload{
			Type: internal.CodeSuccess,
			Data: c.countries,
		}, nil
	} else {
		return &pb.Payload{
			Type: internal.CodeEmptyPayloadError,
			Data: nil,
		}, nil
	}
}
func (c mockDaemonClient) Groups(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.Payload, error) {
	if c.groups != nil {
		return &pb.Payload{
			Type: internal.CodeSuccess,
			Data: c.groups,
		}, nil
	} else {
		return &pb.Payload{
			Type: internal.CodeEmptyPayloadError,
			Data: nil,
		}, nil
	}
}

func TestConnectAutoComplete(t *testing.T) {
	category.Set(t, category.Unit)
	mockClient := mockDaemonClient{}
	c := cmd{&mockClient, nil, nil, nil, "", nil}
	tests := []struct {
		name      string
		countries []string
		groups    []string
		expected  []string
		input     []string
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
		{ // in this case because input is empty, countries and groups will be displayed
			name:      "Groups and Countries",
			groups:    []string{"Europe", "Obfuscated_Servers", "The_Americas"},
			countries: []string{"Canada", "France", "Germany"},
			expected:  []string{"Canada", "France", "Germany", "Europe", "Obfuscated_Servers", "The_Americas"},
			input:     []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			set := flag.NewFlagSet("test", 0)
			mockClient.cities = test.expected
			mockClient.countries = test.countries
			mockClient.groups = test.groups
			set.Parse(test.input)
			ctx := cli.NewContext(app, set, &cli.Context{Context: context.Background()})

			result, err := captureOutput(func() {
				c.ConnectAutoComplete(ctx)
			})

			assert.Nil(t, err)

			list, _ := internal.Columns(test.expected)
			assert.NotEmpty(t, list)
			assert.Equal(t, list, result)
		})
	}
}
