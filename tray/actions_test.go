package tray

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type capturedNotification struct {
	summary string
	body    string
}

type fakeNotifier struct {
	notifications []capturedNotification
}

func (f *fakeNotifier) start() {}

func (f *fakeNotifier) sendNotification(summary, body string) error {
	f.notifications = append(f.notifications, capturedNotification{summary: summary, body: body})
	return nil
}

type fakeConnectStream struct {
	grpc.ServerStreamingClient[pb.Payload]
	payloads []*pb.Payload
	idx      int
}

func (f *fakeConnectStream) Recv() (*pb.Payload, error) {
	if f.idx >= len(f.payloads) {
		return nil, io.EOF
	}
	p := f.payloads[f.idx]
	f.idx++
	return p, nil
}

type trayDaemonClient struct {
	pb.DaemonClient
	connectStream     *fakeConnectStream
	tokenInfoResponse *pb.TokenInfoResponse
}

func (c *trayDaemonClient) ReportUIEvent(
	ctx context.Context,
	in *pb.UIEvent,
	opts ...grpc.CallOption,
) (*pb.Payload, error) {
	return &pb.Payload{}, nil
}

func (c *trayDaemonClient) Connect(
	ctx context.Context,
	in *pb.ConnectRequest,
	opts ...grpc.CallOption,
) (grpc.ServerStreamingClient[pb.Payload], error) {
	return c.connectStream, nil
}

func (c *trayDaemonClient) TokenInfo(
	ctx context.Context,
	in *pb.Empty,
	opts ...grpc.CallOption,
) (*pb.TokenInfoResponse, error) {
	if c.tokenInfoResponse != nil {
		return c.tokenInfoResponse, nil
	}
	return &pb.TokenInfoResponse{}, nil
}

type trayFixture struct {
	instance *Instance
	notifier *fakeNotifier
	client   *trayDaemonClient
}

func newTrayFixture(t *testing.T, payloads ...*pb.Payload) *trayFixture {
	t.Helper()

	notifier := &fakeNotifier{}
	client := &trayDaemonClient{
		connectStream: &fakeConnectStream{payloads: payloads},
	}
	ti := &Instance{
		client:   client,
		notifier: notifier,
	}
	ti.state.initialSyncCompleted = true
	ti.state.notificationsStatus = Enabled

	return &trayFixture{instance: ti, notifier: notifier, client: client}
}

func TestConnect_DedicatedServersErrorPaths(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		code     int64
		wantBody string
	}{
		{
			name: "renew error opens upsell with subscription URL",
			code: internal.CodeDedicatedServersRenewError,
			wantBody: fmt.Sprintf(
				cli.DedicatedServersNoServiceMessage,
				client.DedicatedServersUpselURL,
			),
		},
		{
			name: "service-but-no-servers opens setup link",
			code: internal.CodeDedicatedServersServiceButNoServers,
			wantBody: fmt.Sprintf(
				cli.DedicatedServersNoServersAvailable,
				client.DedicatedServersSetupURL,
			),
		},
		{
			name:     "not ready notifies user",
			code:     internal.CodeDedicatedServersNotReady,
			wantBody: cli.DedicatedServersServerNotReadyMessage,
		},
		{
			name:     "no nordlynx notifies user",
			code:     internal.CodeDedicatedServersNoNordlynx,
			wantBody: cli.DedicatedServersNoNordlynxMessage,
		},
		{
			name:     "server is stopping or stopped",
			code:     internal.CodeDedicatedServersCanNotConnect,
			wantBody: cli.DedicatedServersCanNotConnectMessage,
		},
		{
			name:     "user hit device limit",
			code:     internal.CodeDedicatedServersSessionMaxLimitReached,
			wantBody: cli.DedicatedServersConnectionLimitReached,
		},
		{
			name: "server is in new state",
			code: internal.CodeDedicatedServersServerNotSetUp,
			wantBody: fmt.Sprintf(
				cli.DedicatedServersNoServersAvailable,
				client.DedicatedServersSetupURL,
			),
		},
		{
			name:     "with post-quantun cryptography enabled, connection can't be established",
			code:     internal.CodeDedicatedServersPq,
			wantBody: internal.ServerUnavailableErrorMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newTrayFixture(t, &pb.Payload{Type: tc.code})

			ok := f.instance.connectWithUIEvent(
				"",
				"dedicated_servers",
				pb.UIEvent_CONNECT,
				pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
			)

			assert.False(t, ok, "connect should fail on dedicated-servers error code")
			require.Len(t, f.notifier.notifications, 1, "exactly one notification expected")
			assert.Equal(t, "NordVPN", f.notifier.notifications[0].summary)
			assert.Equal(t, tc.wantBody, f.notifier.notifications[0].body)
		})
	}
}
