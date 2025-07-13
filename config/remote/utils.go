package remote

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	baseKey = "remote-config"

	// defaultMaxGroup represents the maximum value for a rollout group,
	// effectively making the value to be in range of 1-100 (inclusive) to reflect percentage-based groups.
	defaultMaxGroup  uint32 = 100
	logPrefix               = "[Remote Confg]"
	messageNamespace        = "nordvpn-linux"
	rcFailure               = "failure"
	rcSuccess               = "success"
	subscope                = "linux-rc"
)

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
	RCDownload         RCEventType = "rc_download"
	RCDownloadSuccess  RCEventType = "rc_download_success"
	RCDownloadFailure  RCEventType = "rc_download_failure"
	RCLocalUse         RCEventType = "rc_local_use"
	RCJSONParseSuccess RCEventType = "rc_json_parse_success"
	RCJSONParseFailure RCEventType = "rc_json_parse_failure"
	RCPartialRollout   RCEventType = "rc_partial_rollout"
)

// RCDownloadErrorKind defines types of download errors for remote config.
type RCDownloadErrorKind string

const (
	RCDownloadErrorRemoteHashNotFound RCDownloadErrorKind = "remote_hash_not_found"
	RCDownloadErrorRemoteFileNotFound RCDownloadErrorKind = "remote_file_not_found"
	RCDownloadErrorIntegrity          RCDownloadErrorKind = "integrity_error"
	RCDownloadErrorFileDownload       RCDownloadErrorKind = "file_download_error"
	RCDownloadErrorNetwork            RCDownloadErrorKind = "network_error"
	RCDownloadErrorOther              RCDownloadErrorKind = "other_error"
)

type Context struct {
	AppVersion   string
	Country      string
	ISP          string
	RolloutGroup int
}

// RCEventDetails holds optional details for an event.
type RCEventDetails struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// RCEvent is the main analytics event structure.
type RCEvent struct {
	Context
	RCEventDetails
	MessageNamespace string `json:"namespace"`
	Result           string `json:"result"`
	Subscope         string `json:"subscope"`
	Client           string `json:"client,omitempty"`
	FeatureName      string
	Type             RCEventType `json:"type"`
	Timestamp        time.Time   `json:"timestamp"`
}

func NewRCDownloadSuccess(ctx Context, client, featureName string) RCEvent {
	return RCEvent{
		Context:     ctx,
		Client:      client,
		FeatureName: featureName,
		Type:        RCDownloadSuccess,
		Timestamp:   time.Now(),
	}
}

func NewRCDownloadFailure(ctx Context, client, featureName string, errorKind RCDownloadErrorKind, errorMessage string) RCEvent {
	return RCEvent{
		Context:     ctx,
		Client:      client,
		FeatureName: featureName,
		Type:        RCDownloadFailure,
		Timestamp:   time.Now(),
		RCEventDetails: RCEventDetails{
			Error:   string(errorKind),
			Message: errorMessage,
		},
	}
}

func NewRCLocalUse(ctx Context, client, featureName string) RCEvent {
	return RCEvent{
		Context:     ctx,
		Client:      client,
		FeatureName: featureName,
		Type:        RCLocalUse,
		Timestamp:   time.Now(),
	}
}

func NewRCJSONParse(ctx Context, client, featureName, errorKind, errorMessage string) RCEvent {
	detail := RCEventDetails{}
	var eventType RCEventType
	if errorKind != "" {
		detail.Error = errorKind
		detail.Message = errorMessage
		eventType = RCJSONParseFailure
	} else {
		eventType = RCJSONParseSuccess
	}

	return RCEvent{
		Context:        ctx,
		Client:         client,
		FeatureName:    featureName,
		Type:           eventType,
		Timestamp:      time.Now(),
		RCEventDetails: detail,
	}
}

func NewRCPartialRollout(ctx Context, client, featureName, errorKind string, featureRollout int) RCEvent {
	//messagae format is slightly different for partial rollout events
	payload := struct {
		FeatureName    string `json:"feature.name"`
		RolloutGroup   int    `json:"rollout_group"`
		FeatureRollout int    `json:"feature.rollout"`
	}{
		FeatureName:    featureName,
		RolloutGroup:   ctx.RolloutGroup,
		FeatureRollout: featureRollout,
	}

	messageBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("%s%s Failed to marshal partial rollout message: %s. Defaulting to an empty\n", internal.WarningPrefix, logPrefix, err)
		messageBytes = []byte{}
	}

	detail := RCEventDetails{
		Message: string(messageBytes),
	}

	if errorKind != "" {
		detail.Error = errorKind
	}

	return RCEvent{
		Context:        ctx,
		Client:         client,
		FeatureName:    featureName,
		Type:           RCPartialRollout,
		Timestamp:      time.Now(),
		RCEventDetails: detail,
	}
}

func (e RCEvent) ToMooseDebuggerEvent() *events.MooseDebuggerEvent {
	eventToMarshal := e
	eventToMarshal.MessageNamespace = messageNamespace
	eventToMarshal.Subscope = subscope
	if e.Error != "" {
		eventToMarshal.Result = rcFailure
	} else {
		eventToMarshal.Result = rcSuccess
	}

	jsonData, err := json.Marshal(eventToMarshal)
	if err != nil {
		log.Printf("%s%s Failed to marshal RCEvent to JSON %s. Defaulting to empty\n", internal.WarningPrefix, logPrefix, err)
		jsonData = []byte("{}")
	}
	return events.NewMooseDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: baseKey + ".type", Value: string(e.Type)},
			events.ContextValue{Path: baseKey + ".app_version", Value: e.AppVersion},
			events.ContextValue{Path: baseKey + ".country", Value: e.Country},
			events.ContextValue{Path: baseKey + ".isp", Value: e.ISP},
			events.ContextValue{Path: baseKey + ".error", Value: e.Error},
			events.ContextValue{Path: baseKey + ".feature_name", Value: e.FeatureName},
			events.ContextValue{Path: baseKey + ".rollout_group", Value: e.RolloutGroup},
		).
		WithGlobalContextPaths("device.*", "application.nordvpnapp.*", "application.nordvpnapp.version", "application.nordvpnapp.platform")
}
