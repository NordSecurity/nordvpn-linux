//go:build !moose

package remote

import (
	"github.com/NordSecurity/nordvpn-linux/events"
)

type DummyAnalytics struct{}

func NewRemoteConfigAnalytics(events.PublishSubcriber[events.DebuggerEvent], string, int) *DummyAnalytics {
	return &DummyAnalytics{}
}

func (ma *DummyAnalytics) NotifyDownload(string, string)                       {}
func (ma *DummyAnalytics) NotifyDownloadFailure(string, string, DownloadError) {}
func (ma *DummyAnalytics) NotifyLocalUse(string, string, error)                {}
func (ma *DummyAnalytics) NotifyJsonParse(string, string, error)               {}
func (ma *DummyAnalytics) NotifyPartialRollout(string, string, int, bool)      {}
