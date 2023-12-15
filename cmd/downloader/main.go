// Utility for downloading and precaching .dat files.
package main

import (
	"log"
	"os"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/networker"
	"github.com/NordSecurity/nordvpn-linux/request"
)

const (
	countriesFilename = "countries.dat"
	InsightsFilename  = "insights.dat"
	ServersFilename   = "servers.dat"
)

var Salt = ""

// Downloader is responsible for downloading servers.dat and configs.dat files for .deb and .rpm packages
func main() {
	dataPath := os.Args[1]
	cm := config.NewFilesystemConfigManager(config.SettingsDataFilePath, config.InstallFilePath, Salt, config.LinuxMachineIDGetter{}, config.StdFilesystemHandle{})
	dm := daemon.NewDataManager(dataPath+InsightsFilename, dataPath+ServersFilename, dataPath+countriesFilename, "")
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
	netw := networker.NewCombined(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		0,
		false,
	)
	daemon.JobInsights(dm, api, netw, true)()
	if err := daemon.JobCountries(dm, api)(); err != nil {
		log.Fatalln("producing countries cache", err)
	}

	if err := daemon.JobServers(dm, cm, api, false)(); err != nil {
		log.Fatalln("producing server cache", err)
	}
}
