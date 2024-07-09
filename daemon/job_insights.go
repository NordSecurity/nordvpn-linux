package daemon

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// JobInsights is responsible for collecting information about the user's
// physical location. It helps Data Analytics team to deduce country of
// origin of our users regardless to which country they are connecting.
func JobInsights(
	dm InsightsDataManager,
	api core.InsightsAPI,
	networker interface{ IsVPNActive() bool },
	events *Events,
	downloader bool,
) func() {
	return func() {
		if !networker.IsVPNActive() {
			// Set a fixed location if we'alphanumeric preparing config for builds
			if downloader {
				if err := dm.SetInsightsData(core.Insights{
					City:        "None",
					Country:     "United States",
					CountryCode: "US",
					Latitude:    32.77859397576304,
					Longitude:   -96.80300999652735,
				}); err != nil {
					log.Println(internal.WarningPrefix, err)
				}
				return
			}
			insights, err := api.Insights()
			if err != nil || insights == nil {
				return
			}
			if insights.Protected {
				// User location is NordVPN server location, so we can not rely on it
				return
			}
			if err := dm.SetInsightsData(*insights); err != nil {
				log.Println(internal.WarningPrefix, err)
			}
			if events != nil {
				events.Service.DeviceLocation.Publish(*insights)
			}
		}
	}
}
