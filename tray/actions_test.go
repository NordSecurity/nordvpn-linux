package tray

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/alert"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

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
	uiEvents          []*pb.UIEvent
}

func (c *trayDaemonClient) ReportUIEvent(
	ctx context.Context,
	in *pb.UIEvent,
	opts ...grpc.CallOption,
) (*pb.Payload, error) {
	c.uiEvents = append(c.uiEvents, in)
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
	instance  *Instance
	n         *alert.NotifierMock
	client    *trayDaemonClient
	openedURI string
}

func newTrayFixture(t *testing.T, payloads ...*pb.Payload) *trayFixture {
	t.Helper()

	notifier := &alert.NotifierMock{
		Alerts:   []alert.Alert{},
		NextID:   0,
		UpdateCh: make(chan struct{}, 1),
	}
	client := &trayDaemonClient{
		connectStream: &fakeConnectStream{payloads: payloads},
	}

	ti := &Instance{
		client: client,
		n:      notifier,
	}
	ti.state.initialSyncCompleted = true
	ti.state.notificationsStatus = Enabled

	trayFixture := trayFixture{instance: ti, n: notifier, client: client}

	ti.openURI = func(uri string) error {
		trayFixture.openedURI = uri
		return nil
	}

	return &trayFixture
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
			require.Len(t, f.n.Alerts, 1, "exactly one notification expected")
			assert.Equal(t, "NordVPN", f.n.Alerts[0].Summary)
			assert.Equal(t, tc.wantBody, f.n.Alerts[0].Body)
		})
	}
}

func TestGUIDownloadURL_UTMParameters(t *testing.T) {
	category.Set(t, category.Unit)

	parsed, err := url.Parse(guiDownloadPageURL)
	require.NoError(t, err, "guiDownloadPageURL must be a valid URL")

	query := parsed.Query()
	tests := []struct {
		name  string
		param string
		want  string
	}{
		{name: "medium", param: "utm_medium", want: "app"},
		{name: "source", param: "utm_source", want: "nordvpn-linux-tray"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, query.Get(tc.param),
				"UTM parameter %q missing or wrong", tc.param)
		})
	}
}

func TestOpenGUIActions_ReportEventAndOpenURI(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		action   func(*Instance)
		wantItem pb.UIEvent_ItemName
		wantURI  string
	}{
		{
			name:     "open app",
			action:   (*Instance).openGUI,
			wantItem: pb.UIEvent_OPEN_APP,
			wantURI:  guiLaunchURI,
		},
		{
			name:     "download app",
			action:   (*Instance).openGUIDownloadPage,
			wantItem: pb.UIEvent_DOWNLOAD_APP,
			wantURI:  guiDownloadPageURL,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newTrayFixture(t)

			tc.action(f.instance)

			require.Len(t, f.client.uiEvents, 1, "expected exactly one UI event")
			ev := f.client.uiEvents[0]
			assert.Equal(t, pb.UIEvent_TRAY, ev.FormReference)
			assert.Equal(t, tc.wantItem, ev.ItemName)
			assert.Equal(t, pb.UIEvent_CLICK, ev.ItemType)
			assert.Equal(t, tc.wantURI, f.openedURI, "opened the wrong URI")
		})
	}
}
