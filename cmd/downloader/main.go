// Utility for downloading and precaching .dat files.
package main

import (
	"log"
	"os"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/request"
)

const (
	countriesFilename = "countries.dat"
	InsightsFilename  = "insights.dat"
	ServersFilename   = "servers.dat"
)

type vpnChecker struct{}

func (vpnChecker) IsVPNActive() bool {
	return false
}

// Downloader is responsible for downloading servers.dat and configs.dat files for .deb and .rpm packages
func main() {
	dataPath := os.Args[1]
	dm := daemon.NewDataManager(dataPath+InsightsFilename, dataPath+ServersFilename, dataPath+countriesFilename, "", events.NewDataUpdateEvents())
	client := request.NewStdHTTP()
	validator, err := response.NewNordValidator()
	if err != nil {
		log.Fatalln("creating nord validator:", err)
	}

	api := core.NewDefaultAPI(
		"",
		daemon.BaseURL,
		client,
		validator,
	)
	daemon.JobInsights(dm, api, vpnChecker{}, nil, true)()
	if err := daemon.JobCountries(dm, api)(); err != nil {
		log.Fatalln("producing countries cache", err)
	}

	if err := daemon.JobServers(dm, api, false)(); err != nil {
		log.Fatalln("producing server cache", err)
	}
}
