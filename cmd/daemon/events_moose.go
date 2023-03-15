//go:build moose

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events/moose"
)

func newAnalytics(eventsDbPath string, fs *config.Filesystem,
	ver, salt, env string) *moose.Subscriber {
	return &moose.Subscriber{
		EventsDbPath: eventsDbPath,
		Config:       fs,
		Version:      ver,
		Salt:         salt,
		Environment:  env,
		Domain:       EventsDomain,
		Subdomain:    EventsSubdomain,
	}
}
