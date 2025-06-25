package telemetry

import (
	"context"

	pb "github.com/NordSecurity/nordvpn-linux/daemon/pb/telemetry/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MetricsListener func(metric Metric, value any) error

type CustomField struct {
	Label string
	Value string
}

func New(listener MetricsListener) *Telemetry {
	return &Telemetry{listen: listener}
}

type Telemetry struct {
	listen MetricsListener
	pb.UnimplementedTelemetryServiceServer
}

func (t Telemetry) SetDesktopEnvironment(ctx context.Context, in *pb.DesktopEnvironmentRequest) (*emptypb.Empty, error) {
	return nil, t.listen(MetricDesktopEnvironment, in.DesktopEnvName)
}

func (t Telemetry) SetDisplayProtocol(ctx context.Context, in *pb.DisplayProtocolRequest) (*emptypb.Empty, error) {
	return nil, t.listen(MetricDisplayProtocol, in.Protocol)
}
