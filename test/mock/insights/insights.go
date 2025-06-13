package insights

import "github.com/NordSecurity/nordvpn-linux/core"

type InsightsMock struct {
	InsightsResult *core.Insights
	Err            error
}

func (f *InsightsMock) Insights() (*core.Insights, error) {
	return f.InsightsResult, f.Err
}
