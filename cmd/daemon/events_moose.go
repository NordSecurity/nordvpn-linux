//go:build moose

package main

import (
	"net/http"
	"os"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events/moose"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
)

var (
	EventsDomain    = ""
	EventsSubdomain = ""
)

func newAnalytics(
	eventsDbPath string,
	fs *config.FilesystemConfigManager,
	clientAPI core.ClientAPI,
	httpClient *http.Client,
	buildTarget config.BuildTarget,
	hostInfo sysinfo.HostInfo,
	id string,
) *moose.Subscriber {
	_ = os.Setenv("MOOSE_LOG_FILE", "Stdout")

	return moose.NewSubscriber(eventsDbPath, fs, clientAPI, httpClient, buildTarget, hostInfo, id, EventsDomain, EventsSubdomain)
}
