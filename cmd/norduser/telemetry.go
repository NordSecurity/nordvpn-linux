package main

import (
	"context"
	"log"
	"sync"
	"time"

	telemetrypb "github.com/NordSecurity/nordvpn-linux/daemon/pb/telemetry/v1"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
	"google.golang.org/grpc"
)

const telemetryTimeout = 5 * time.Second
const tag = "[telemetry]"
const (
	defaultDesktopEnv = ""
	x11DesktopEnv     = "x11"
	waylandDesktopEnv = "wayland"
)

type ReportEvent int

const (
	ReportOnStart ReportEvent = iota
	ReportOnExit
)

// ReportTelemetry reports underlying telemetry asynchronously.
// It reports metrics once.
func ReportTelemetry(conn *grpc.ClientConn, evt ReportEvent, wait bool) {
	client := telemetrypb.NewTelemetryServiceClient(conn)
	submitEmptyMetrics := evt == ReportOnExit // skip gathering actual system info on exit

	var wg *sync.WaitGroup
	if wait {
		wg = new(sync.WaitGroup)
		wg.Add(2)
	}

	go sendDesktopEnvironmentMetric(client, submitEmptyMetrics, wg)
	go sendDisplayProtocolMetric(client, submitEmptyMetrics, wg)

	if wg != nil {
		wg.Wait()
	}
}

func logErrorMessage(metric string, value string, err error) {
	log.Printf("%s %s Failed to send metric: metric=%s, value=%s, error=%v",
		internal.WarningPrefix, tag, metric, value, err)
}

// sendDesktopEnvironmentMetric sends the current desktop environment
// to the telemetry service with a timeout context.
func sendDesktopEnvironmentMetric(
	client telemetrypb.TelemetryServiceClient,
	submitEmpty bool,
	wg *sync.WaitGroup,
) {
	if wg != nil {
		defer wg.Done()
	}

	ctx, cancel := context.WithTimeout(context.Background(), telemetryTimeout)
	defer cancel()

	de := defaultDesktopEnv
	if !submitEmpty {
		de = sysinfo.GetDesktopEnvironment()
	}

	req := &telemetrypb.DesktopEnvironmentRequest{DesktopEnvName: de}
	if _, err := client.SetDesktopEnvironment(ctx, req); err != nil {
		logErrorMessage("DesktopEnvironment", de, err)
	}
}

// sendDisplayProtocolMetric sends the current display protocol (e.g., X11, Wayland)
// to the telemetry service with a timeout context.
func sendDisplayProtocolMetric(
	client telemetrypb.TelemetryServiceClient,
	submitEmpty bool,
	wg *sync.WaitGroup,
) {
	if wg != nil {
		defer wg.Done()
	}

	ctx, cancel := context.WithTimeout(context.Background(), telemetryTimeout)
	defer cancel()

	protocol := telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_UNSPECIFIED
	if !submitEmpty {
		switch sysinfo.GetDisplayProtocol() {
		case x11DesktopEnv:
			protocol = telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_X11
		case waylandDesktopEnv:
			protocol = telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_WAYLAND
		default:
			protocol = telemetrypb.DisplayProtocol_DISPLAY_PROTOCOL_UNKNOWN
		}
	}

	req := &telemetrypb.DisplayProtocolRequest{Protocol: protocol}
	if _, err := client.SetDisplayProtocol(ctx, req); err != nil {
		logErrorMessage("DisplayProtocol", protocol.String(), err)
	}
}
