package remote

import (
	"github.com/NordSecurity/nordvpn-linux/events"
)

// Analytics defines an interface for reporting various analytics related to remote config
// including: download, RC local usage, JSON parsing, and partial rollouts event.
type Analytics interface {
	EmitDownloadEvent(string, string)
	EmitDownloadFailureEvent(string, string, DownloadError)
	EmitLocalUseEvent(string, string, error)
	EmitJsonParseEvent(string, string, error)
	EmitPartialRolloutEvent(string, string, int, bool)
}
type RemoteConfigAnalytics struct {
	publisher events.PublishSubcriber[events.DebuggerEvent]
	userInfo  UserInfo
}

func NewRemoteConfigAnalytics(publisher events.PublishSubcriber[events.DebuggerEvent], ver string, rg int) *RemoteConfigAnalytics {
	return &RemoteConfigAnalytics{publisher: publisher, userInfo: UserInfo{AppVersion: ver, RolloutGroup: rg}}
}
func (rca *RemoteConfigAnalytics) EmitDownloadEvent(client, featureName string) {
	rca.publisher.Publish(*NewDownloadSuccessEvent(rca.userInfo, client, featureName).ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitDownloadFailureEvent(client, featureName string, err DownloadError) {
	rca.publisher.Publish(
		*NewDownloadFailureEvent(rca.userInfo, client, featureName, err.Kind, err.Error()).ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitLocalUseEvent(client, featureName string, err error) {
	rca.publisher.Publish(
		*NewLocalUseEvent(rca.userInfo, client, featureName).ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitJsonParseEvent(client, featureName string, err error) {
	var errMsg string
	errorKind := ""
	if err != nil {
		errMsg = err.Error()
		errorKind = err.Error()
	}
	rca.publisher.Publish(
		*NewJSONParseEvent(rca.userInfo, client, featureName, errorKind, errMsg).
			ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitPartialRolloutEvent(client, featureName string, frg int, rolloutPerformed bool) {
	rca.publisher.Publish(
		*NewRolloutEvent(rca.userInfo, client, featureName, frg, rolloutPerformed).
			ToDebuggerEvent())
}
