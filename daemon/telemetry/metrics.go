//go:generate stringer -type=Metric
package telemetry

type Metric int

const (
	MetricDesktopEnvironment Metric = iota
	MetricDisplayProtocol
)
