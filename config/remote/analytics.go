package remote

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// Analytics defines an interface for reporting various analytics related to remote config
// including: download, RC local usage, JSON parsing, and partial rollouts event.
type Analytics interface {
	EmitDownloadEvent(client string, featureName string)
	EmitDownloadFailureEvent(client string, featureName string, err DownloadError)
	EmitLocalUseEvent(client string, featureName string, err error)
	EmitJsonParseFailureEvent(client string, featureName string, err LoadError)
	EmitPartialRolloutEvent(client string, featureName string, frg int, rolloutPerformed bool)
	ClearEventFlags()
}

type eventKey struct {
	feature   string
	eventType EventType
}

// eventFlagsSet is used to control event frequency per feature & event
type eventFlagsSet map[eventKey]struct{}

func (m eventFlagsSet) set(feature string, e EventType) {
	m[eventKey{feature: feature, eventType: e}] = struct{}{}
}
func (m eventFlagsSet) has(feature string, e EventType) bool {
	_, found := m[eventKey{feature: feature, eventType: e}]
	return found
}
func (m eventFlagsSet) clear() {
	clear(m)
}

type RemoteConfigAnalytics struct {
	publisher        events.PublishSubcriber[events.DebuggerEvent]
	userRolloutGroup int
	eventFlags       eventFlagsSet
	mu               sync.Mutex
}

func NewRemoteConfigAnalytics(publisher events.PublishSubcriber[events.DebuggerEvent], rolloutGroup int) *RemoteConfigAnalytics {
	return &RemoteConfigAnalytics{
		publisher:        publisher,
		userRolloutGroup: rolloutGroup,
		eventFlags:       make(eventFlagsSet),
	}
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
	rca.mu.Lock()
	if rca.eventAlreadyEmitted(featureName, Rollout) {
		rca.mu.Unlock()
		return
	}
	rca.eventFlags.set(featureName, Rollout)
	rca.mu.Unlock()

	rca.publisher.Publish(
		*NewRolloutEvent(rca.userRolloutGroup, client, featureName, frg, rolloutPerformed).
			ToDebuggerEvent())
}

func (rca *RemoteConfigAnalytics) eventAlreadyEmitted(feature string, event EventType) bool {
	return rca.eventFlags.has(feature, event)
}

func (rca *RemoteConfigAnalytics) ClearEventFlags() {
	rca.mu.Lock()
	defer rca.mu.Unlock()
	rca.eventFlags.clear()
}
