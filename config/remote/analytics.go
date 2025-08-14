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
	EmitJsonParseFailureEvent(string, string, LoadError)
	EmitPartialRolloutEvent(string, string, int, bool)
}
type RemoteConfigAnalytics struct {
	publisher        events.PublishSubcriber[events.DebuggerEvent]
	userRolloutGroup int
}

func NewRemoteConfigAnalytics(publisher events.PublishSubcriber[events.DebuggerEvent], rg int) *RemoteConfigAnalytics {
	return &RemoteConfigAnalytics{publisher: publisher, userRolloutGroup: rg}
}
func (rca *RemoteConfigAnalytics) EmitDownloadEvent(client, featureName string) {
	rca.publisher.Publish(*NewDownloadSuccessEvent(rca.userRolloutGroup, client, featureName).ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitDownloadFailureEvent(client, featureName string, err DownloadError) {
	rca.publisher.Publish(
		*NewDownloadFailureEvent(rca.userRolloutGroup, client, featureName, err.Kind, err.Error()).ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitLocalUseEvent(client, featureName string, err error) {
	var errMsg string
	var errorKind string
	if err != nil {
		errMsg = err.Error()
		errorKind = err.Error()
	}
	rca.publisher.Publish(
		*NewLocalUseEvent(rca.userRolloutGroup, client, featureName, errorKind, errMsg).ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitJsonParseFailureEvent(client, featureName string, err LoadError) {
	rca.publisher.Publish(
		*NewJSONParseEventFailure(rca.userRolloutGroup, client, featureName, err.Kind, err.Error()).
			ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) EmitPartialRolloutEvent(client, featureName string, frg int, rolloutPerformed bool) {
	rca.publisher.Publish(
		*NewRolloutEvent(rca.userRolloutGroup, client, featureName, frg, rolloutPerformed).
			ToDebuggerEvent())
}
