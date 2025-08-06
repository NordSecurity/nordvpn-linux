//go:build !moose

package remote

import (
	"github.com/NordSecurity/nordvpn-linux/events"
)

type DummyAnalytics struct{}

func NewRemoteConfigAnalytics(events.PublishSubcriber[events.DebuggerEvent], string, int) *DummyAnalytics {
	return &DummyAnalytics{}
}

func (*DummyAnalytics) NotifyDownload(string, string)                       {}
func (*DummyAnalytics) NotifyDownloadFailure(string, string, DownloadError) {}
func (*DummyAnalytics) NotifyLocalUse(string, string, error)                {}
func (*DummyAnalytics) NotifyJsonParse(string, string, error)               {}
func (*DummyAnalytics) NotifyPartialRollout(string, string, int, bool)      {}
