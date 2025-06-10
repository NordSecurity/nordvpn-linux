package telemetry

type Metric int

const (
	MetricCustom Metric = iota
	MetricDesktopEnvironment
	MetricDisplayProtocol
)
