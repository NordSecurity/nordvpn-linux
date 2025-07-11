package remote

import (
	"crypto/sha256"
	"encoding/binary"

	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/NordSecurity/nordvpn-linux/events"
)

// defaultMaxGroup represents the maximum value for a rollout group,
// effectively making the value to be in range of 1-100 (inclusive) to reflect percentage-based groups.
const defaultMaxGroup uint32 = 100

// GenerateRolloutGroup creates a new RolloutGroup instance based on a provided UUID.
// It computes a group value by hashing the UUID and deriving a number between 1 and defaultMaxGroup (inclusive).
//
// Parameters:
//   - uuid: The UUID used as the basis for group assignment
//
// Returns:
//   - RolloutGroup: A new RolloutGroup instance with the computed value
func GenerateRolloutGroup(uuid uuid.UUID) int {
	hash := sha256.Sum256(uuid[:])
	num := binary.BigEndian.Uint32(hash[:])
	value := int(num%defaultMaxGroup) + 1
	return value
}

// RCEventType defines the type of remote config analytics event.
type RCEventType string

const (
	RCDownload        RCEventType = "rc_download"
	RCDownloadSuccess RCEventType = "rc_download_success"
	RCDownloadFailure RCEventType = "rc_download_failure"
	RCLocalUse        RCEventType = "rc_local_use"
	RCJSONParse       RCEventType = "rc_json_parse"
	RCPartialRollout  RCEventType = "rc_partial_rollout"
)

const keyBasedFeatureKey = "application.nordvpnapp"

// RCEventDetails holds optional details for an event.
type RCEventDetails struct {
	Error        string `json:"error,omitempty"`
	FeatureName  string `json:"feature_name,omitempty"`
	RolloutGroup int    `json:"rollout_group,omitempty"`
	RolloutValue int    `json:"rollout_value,omitempty"`
}

// RCEvent is the main analytics event structure.
type RCEvent struct {
	Type       RCEventType    `json:"type"`
	Timestamp  time.Time      `json:"timestamp"`
	AppVersion string         `json:"app_version,omitempty"`
	Country    string         `json:"country,omitempty"`
	ISP        string         `json:"isp,omitempty"`
	Details    RCEventDetails `json:"details,omitempty"`
}

// --- Helper constructors ---

func NewRCDownloadSuccess(appVersion, country, isp string) RCEvent {
	return RCEvent{
		Type:       RCDownloadSuccess,
		Timestamp:  time.Now(),
		AppVersion: appVersion,
		Country:    country,
		ISP:        isp,
	}
}

func NewRCDownloadFailure(appVersion, country, isp, errDetail string) RCEvent {
	return RCEvent{
		Type:       RCDownloadFailure,
		Timestamp:  time.Now(),
		AppVersion: appVersion,
		Country:    country,
		ISP:        isp,
		Details:    RCEventDetails{Error: errDetail},
	}
}

func NewRCJSONParse(appVersion, country, isp, errDetail string, success bool) RCEvent {
	detail := RCEventDetails{}
	if !success {
		detail.Error = errDetail
	}
	return RCEvent{
		Type:       RCJSONParse,
		Timestamp:  time.Now(),
		AppVersion: appVersion,
		Country:    country,
		ISP:        isp,
		Details:    detail,
	}
}

func NewRCPartialRollout(appVersion, country, isp, featureName string, rolloutGroup, rolloutValue int, failure string) RCEvent {
	detail := RCEventDetails{
		FeatureName:  featureName,
		RolloutGroup: rolloutGroup,
		RolloutValue: rolloutValue,
	}
	if failure != "" {
		detail.Error = failure
	}
	return RCEvent{
		Type:       RCPartialRollout,
		Timestamp:  time.Now(),
		AppVersion: appVersion,
		Country:    country,
		ISP:        isp,
		Details:    detail,
	}
}

func (e RCEvent) ToMooseDebuggerEvent() (*events.MooseDebuggerEvent, error) {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	ev := events.NewMooseDebuggerEvent(string(jsonData))
	ev.WithKeyBasedContextPaths()
	return ev.WithKeyBasedContextPaths(
		events.ContextValue{Path: keyBasedFeatureKey + ".type.", Value: e.Type},
		events.ContextValue{Path: keyBasedFeatureKey + ".app_version.", Value: e.AppVersion},
		events.ContextValue{Path: keyBasedFeatureKey + ".country.", Value: e.Country},
		events.ContextValue{Path: keyBasedFeatureKey + ".isp.", Value: e.ISP},
		events.ContextValue{Path: keyBasedFeatureKey + ".error.", Value: e.Details.Error},
		events.ContextValue{Path: keyBasedFeatureKey + ".feature_name.", Value: e.Details.FeatureName},
		events.ContextValue{Path: keyBasedFeatureKey + ".rollout_group.", Value: e.Details.RolloutGroup},
		events.ContextValue{Path: keyBasedFeatureKey + ".rollout_value.", Value: e.Details.RolloutValue},
	).WithGlobalContextPaths("device.*",
		"application.nordvpnapp.name",
		"application.nordvpnapp.version",
		"application.nordvpnapp.platform"), nil
}
