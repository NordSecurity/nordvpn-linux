package daemon

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type workingInsightsAPI struct{}

func (workingInsightsAPI) Insights() (*core.Insights, error) {
	return &core.Insights{
		CountryCode: "US",
		Latitude:    34.023300,
		Longitude:   -117.851200,
	}, nil
}

func (a workingInsightsAPI) InsightsViaTunnel() (*core.Insights, error) {
	return a.Insights()
}

type workingInsightsDataManager struct {
	data InsightsData
}

func (w *workingInsightsDataManager) GetInsightsData() InsightsData {
	return w.data
}

func (w *workingInsightsDataManager) SetInsightsData(data core.Insights) error {
	w.data = InsightsData{Insights: data}
	return nil
}

type InsightsTestData struct {
	CountryCode          string
	Latitude, Longtitude float64
}

type vpnActiveNetworker struct{}

func (vpnActiveNetworker) IsVPNActive() bool { return true }

type vpnInactiveNetworker struct{}

func (vpnInactiveNetworker) IsVPNActive() bool { return false }

func TestJobInsights(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		dm           InsightsDataManager
		networker    interface{ IsVPNActive() bool }
		country      string
		isDownloader bool
		expected     InsightsData
	}{
		{
			name:      "connected to vpn",
			dm:        &workingInsightsDataManager{},
			networker: vpnActiveNetworker{},
			country:   "US",
		},
		{
			name:      "not connected to vpn",
			dm:        &workingInsightsDataManager{},
			networker: vpnInactiveNetworker{},
			country:   "US",
			expected: InsightsData{
				Insights: core.Insights{
					CountryCode: "US",
					Latitude:    34.023300,
					Longitude:   -117.851200,
				},
			},
		},
		{
			name:         "downloader",
			dm:           &workingInsightsDataManager{},
			networker:    vpnInactiveNetworker{},
			country:      "US",
			isDownloader: true,
			expected: InsightsData{
				Insights: core.Insights{
					City:        "None",
					Country:     "United States",
					CountryCode: "US",
					Latitude:    32.77859397576304,
					Longitude:   -96.80300999652735,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			JobInsights(
				test.dm,
				workingInsightsAPI{},
				test.networker,
				nil,
				test.isDownloader,
			)()
			assert.Equal(t, test.expected, test.dm.GetInsightsData())
		})
	}
}
