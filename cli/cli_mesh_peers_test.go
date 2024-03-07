package cli

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
)

type mockMeshClient struct {
	pb.MeshnetClient
}

func (m mockMeshClient) GetPeers(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.GetPeersResponse, error) {
	x := &pb.GetPeersResponse{
		Response: &pb.GetPeersResponse_Peers{
			Peers: &pb.PeerList{
				Self: &pb.Peer{Hostname: "test"},
				Local: []*pb.Peer{{
					Hostname:          "IAllowInbound_InboundAllowed_IsRoutable",
					IsInboundAllowed:  true,
					IsRoutable:        true,
					DoIAllowInbound:   true,
					DoIAllowRouting:   false,
					DoIAllowFileshare: false,
					Status:            1,
				}, {
					Hostname:          "IAllowInbound_IAllowRouting_IAllowFileshare",
					IsInboundAllowed:  false,
					IsRoutable:        false,
					DoIAllowInbound:   true,
					DoIAllowRouting:   true,
					DoIAllowFileshare: true,
					Status:            0,
				}},
				External: []*pb.Peer{{
					Hostname:          "AllowsEverything",
					IsInboundAllowed:  true,
					IsRoutable:        true,
					DoIAllowInbound:   true,
					DoIAllowRouting:   true,
					DoIAllowFileshare: true,
					Status:            1,
				}, {
					Hostname:          "IsRoutable",
					IsInboundAllowed:  false,
					IsRoutable:        true,
					DoIAllowInbound:   false,
					DoIAllowRouting:   false,
					DoIAllowFileshare: false,
					Status:            0,
				}},
			},
		},
	}
	return x, nil
}

func TestMeshPeerList(t *testing.T) {
	category.Set(t, category.Unit)
	c := cmd{nil, mockMeshClient{}, nil, nil, "", nil}
	tests := []struct {
		name     string
		expected []string
	}{
		{
			name:     "online",
			expected: []string{"IAllowInbound_InboundAllowed_IsRoutable", "AllowsEverything"},
		},
		{
			name:     "offline",
			expected: []string{"IAllowInbound_IAllowRouting_IAllowFileshare", "IsRoutable"},
		},
		{
			name:     "internal",
			expected: []string{"IAllowInbound_InboundAllowed_IsRoutable", "IAllowInbound_IAllowRouting_IAllowFileshare"},
		},
		{
			name:     "external",
			expected: []string{"AllowsEverything", "IsRoutable"},
		},
		{
			name:     "allows-incoming-traffic",
			expected: []string{"IAllowInbound_InboundAllowed_IsRoutable", "AllowsEverything"},
		},
		{
			name:     "allows-routing",
			expected: []string{"AllowsEverything", "IsRoutable", "IAllowInbound_InboundAllowed_IsRoutable"},
		},
		{
			name:     "incoming-traffic-allowed",
			expected: []string{"IAllowInbound_InboundAllowed_IsRoutable", "IAllowInbound_IAllowRouting_IAllowFileshare", "AllowsEverything"},
		},
		{
			name:     "routing-allowed",
			expected: []string{"AllowsEverything", "IAllowInbound_IAllowRouting_IAllowFileshare"},
		},
		{
			name:     "online,external",
			expected: []string{"AllowsEverything"},
		},
		{
			name:     "offline,routing-allowed",
			expected: []string{"no peers"},
		},
		{
			name:     "internal,incoming-traffic-allowed",
			expected: []string{"IAllowInbound_InboundAllowed_IsRoutable", "IAllowInbound_IAllowRouting_IAllowFileshare"},
		},
		{
			name:     "allows-sending-files",
			expected: []string{"AllowsEverything", "IAllowInbound_IAllowRouting_IAllowFileshare"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := cli.NewApp()
			ctx := &cli.Context{
				Context: context.Background(),
				App: &cli.App{
					Flags: []cli.Flag{&cli.StringFlag{
						Name:  flagFilter,
						Value: test.name,
					}},
				},
				Command: &cli.Command{},
			}
			set := flag.NewFlagSet(flagFilter, flag.ContinueOnError)
			ctx = cli.NewContext(app, set, ctx)
			set.String(flagFilter, test.name, "filter flag")
			set.Parse([]string{"--filter", test.name})
			set.Set(flagFilter, test.name)
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			c.MeshPeerList(ctx)
			w.Close()
			os.Stdout = old
			var buf bytes.Buffer
			io.Copy(&buf, r)
			for _, i := range test.expected {
				assert.Contains(t, buf.String(), i)
			}
		})
	}
}
