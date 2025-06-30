package telemetry

import (
	"context"
	"log"

	pb "github.com/NordSecurity/nordvpn-linux/daemon/pb/telemetry/v1"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const logTag = "[telemery]"

// MetricsListener defines a callback function used to report collected metrics.
type MetricsListener func(metric Metric, value any) error

// New creates a new Telemetry instance with the provided MetricsListener.
func New(listener MetricsListener) *Telemetry {
	return &Telemetry{metricReporter: listener}
}

// Telemetry implements the TelemetryServiceServer interface and is used to
// collect and report telemetry metrics from clients to the daemon.
type Telemetry struct {
	metricReporter MetricsListener
	pb.UnimplementedTelemetryServiceServer
}

func (t Telemetry) submitMetric(metric Metric, value any) error {
	if t.metricReporter == nil {
		return status.Errorf(codes.FailedPrecondition,
			"metric '%s' not reported: reporter not configured", metric)
	}

	if err := t.metricReporter(metric, value); err != nil {
		log.Printf("%s %s Failed to report metric %s: %s\n",
			logTag, internal.WarningPrefix, metric, err)
		return status.Errorf(codes.Internal, "failed to report metric: %v", err)
	}

	return nil
}

// SetDesktopEnvironment handles the gRPC request to set the desktop environment
// and reports the corresponding telemetry metric.
func (t Telemetry) SetDesktopEnvironment(
	ctx context.Context,
	in *pb.DesktopEnvironmentRequest,
) (*emptypb.Empty, error) {
	return nil, t.submitMetric(MetricDesktopEnvironment, in.DesktopEnvName)
}

// SetDisplayProtocol handles the gRPC request to set the display protocol
// and reports the corresponding telemetry metric.
func (t Telemetry) SetDisplayProtocol(
	ctx context.Context,
	in *pb.DisplayProtocolRequest,
) (*emptypb.Empty, error) {
	return nil, t.submitMetric(MetricDisplayProtocol, in.Protocol)
}
