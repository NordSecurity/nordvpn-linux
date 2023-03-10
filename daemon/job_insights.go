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
	downloader bool,
) func() {
	return func() {
		if !networker.IsVPNActive() {
			// Set a fixed location if we'alphanumeric preparing config for builds
			if downloader {
				if err := dm.SetInsightsData(core.Insights{
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
			if err := dm.SetInsightsData(*insights); err != nil {
				log.Println(internal.WarningPrefix, err)
			}
		}
	}
}
