//go:build !moose

package main

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
)

type dummyAnalytics struct{}

func (*dummyAnalytics) Enable() error                                  { return nil }
func (*dummyAnalytics) Disable() error                                 { return nil }
func (*dummyAnalytics) NotifyKillswitch(bool) error                    { return nil }
func (*dummyAnalytics) NotifyAutoconnect(bool) error                   { return nil }
func (*dummyAnalytics) NotifyDNS(events.DataDNS) error                 { return nil }
func (*dummyAnalytics) NotifyThreatProtectionLite(bool) error          { return nil }
func (*dummyAnalytics) NotifyProtocol(config.Protocol) error           { return nil }
func (*dummyAnalytics) NotifyAllowlist(events.DataAllowlist) error     { return nil }
func (*dummyAnalytics) NotifyTechnology(config.Technology) error       { return nil }
func (*dummyAnalytics) NotifyObfuscate(bool) error                     { return nil }
func (*dummyAnalytics) NotifyFirewall(bool) error                      { return nil }
func (*dummyAnalytics) NotifyRouting(bool) error                       { return nil }
func (*dummyAnalytics) NotifyNotify(bool) error                        { return nil }
func (*dummyAnalytics) NotifyMeshnet(bool) error                       { return nil }
func (*dummyAnalytics) NotifyIpv6(bool) error                          { return nil }
func (*dummyAnalytics) NotifyDefaults(any) error                       { return nil }
func (*dummyAnalytics) NotifyConnect(events.DataConnect) error         { return nil }
func (*dummyAnalytics) NotifyDisconnect(events.DataDisconnect) error   { return nil }
func (*dummyAnalytics) NotifyLogin(any) error                          { return nil }
func (*dummyAnalytics) NotifyAccountCheck(core.ServicesResponse) error { return nil }
func (*dummyAnalytics) NotifyRequestAPI(events.DataRequestAPI) error   { return nil }
func (*dummyAnalytics) NotifyRate(events.ServerRating) error           { return nil }
func (*dummyAnalytics) NotifyHeartBeat(int) error                      { return nil }
func (*dummyAnalytics) NotifyLANDiscovery(bool) error                  { return nil }

func newAnalytics(eventsDbPath string, fs *config.FilesystemConfigManager,
	version, env, id string) *dummyAnalytics {
	return &dummyAnalytics{}
}
