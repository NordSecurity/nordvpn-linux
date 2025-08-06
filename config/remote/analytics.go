package remote

import (
	"github.com/NordSecurity/nordvpn-linux/events"
)

// Analytics defines an interface for reporting various analytics events related to downloads,
// local usage, JSON parsing, and partial rollouts. Implementations of this interface are
// responsible for handling the notification logic for each event type.
type Analytics interface {
	NotifyDownload(string, string)
	NotifyDownloadFailure(string, string, DownloadError)
	NotifyLocalUse(string, string, error)
	NotifyJsonParse(string, string, error)
	NotifyPartialRollout(string, string, int, bool)
}
type MooseAnalytics struct {
	publisher events.PublishSubcriber[events.DebuggerEvent]
	ctx       Event
}

func NewRemoteConfigAnalytics(publisher events.PublishSubcriber[events.DebuggerEvent], ver string, rg int) *MooseAnalytics {
	ctx := Event{
		UserInfo: UserInfo{
			AppVersion:   ver,
			RolloutGroup: rg,
		},
	}
	return &MooseAnalytics{publisher: publisher, ctx: ctx}
}
func (ma *MooseAnalytics) NotifyDownload(client, featureName string) {
	ma.publisher.Publish(*NewDownloadSuccessEvent(ma.ctx.UserInfo, client, featureName).ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyDownloadFailure(client, featureName string, err DownloadError) {
	ma.publisher.Publish(
		*NewDownloadFailureEvent(ma.ctx.UserInfo, client, featureName, err.Kind, err.Error()).ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyLocalUse(client, featureName string, err error) {
	ma.publisher.Publish(
		*NewLocalUseEvent(ma.ctx.UserInfo, client, featureName).ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyJsonParse(client, featureName string, err error) {
	var errMsg string
	errorKind := ""
	if err != nil {
		errMsg = err.Error()
		errorKind = err.Error()
	}
	ma.publisher.Publish(
		*NewJSONParseEvent(ma.ctx.UserInfo, client, featureName, errorKind, errMsg).
			ToDebuggerEvent())
}

func (ma *MooseAnalytics) NotifyPartialRollout(client, featureName string, frg int, rolloutPerformed bool) {
	ma.publisher.Publish(
		*NewRolloutEvent(ma.ctx.UserInfo, client, featureName, frg, rolloutPerformed).
			ToDebuggerEvent())
}
