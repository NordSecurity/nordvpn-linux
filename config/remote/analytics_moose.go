//go:build moose

package remote

import (
	"github.com/NordSecurity/nordvpn-linux/events"
)

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

func (ma *MooseAnalytics) NotifyDownload(client, featureName string, err error) {
	var evt Event
	if err != nil { //TODO/FIXME: downloadErrorKind
		evt = NewDownloadFailureEvent(ma.ctx.UserInfo, client, featureName, DownloadErrorOther, err.Error())
	} else {
		evt = NewDownloadSuccessEvent(ma.ctx.UserInfo, client, featureName)
	}
	ma.publisher.Publish(
		*evt.ToDebuggerEvent())
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
