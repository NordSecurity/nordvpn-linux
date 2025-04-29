//go:build moose

package main

import (
	"log"
	"os"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/moose"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	EventsDomain    = ""
	EventsSubdomain = ""
)

func newAnalytics(
	eventsDbPath string,
	fs *config.FilesystemConfigManager,
	subAPI moose.SubscribtionServicesAPI,
	ver, env, id string) *moose.Subscriber {
	_ = os.Setenv("MOOSE_LOG_FILE", "Stdout")

	sub := &moose.Subscriber{
		EventsDbPath:    eventsDbPath,
		Config:          fs,
		Version:         ver,
		Environment:     env,
		Domain:          EventsDomain,
		Subdomain:       EventsSubdomain,
		DeviceID:        id,
		SubscriptionAPI: subAPI,
	}
	if err := sub.Init(); err != nil {
		log.Println(internal.ErrorPrefix, "MOOSE: Initialization error:", err)
	}
	return sub
}
