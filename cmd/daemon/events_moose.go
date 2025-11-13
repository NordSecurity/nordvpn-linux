//go:build moose

package main

import (
	"net/http"
	"os"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events/moose"
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
	id string) *moose.Subscriber {
	_ = os.Setenv("MOOSE_LOG_FILE", "Stdout")

	return moose.NewSubscriber(eventsDbPath, fs, clientAPI, httpClient, buildTarget, id, EventsDomain, EventsSubdomain)
}
