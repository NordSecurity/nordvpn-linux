package remote

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type TimeGetter func() time.Time

type UserInfo struct {
	AppVersion   string `json:"app_version"`
	Country      string `json:"country"`
	ISP          string `json:"isp"`
	RolloutGroup int    `json:"rollout_group"`
}

// EventDetails holds optional details for an event.
type EventDetails struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Event is the main analytics event structure.
type Event struct {
	UserInfo
	EventDetails
	MessageNamespace string    `json:"namespace"`
	Result           string    `json:"result"`
	Subscope         string    `json:"subscope"`
	Client           string    `json:"client,omitempty"`
	FeatureName      string    `json:"feature_name"`
	Type             EventType `json:"type"`
	Timestamp        time.Time `json:"timestamp"`
}

func NewDownloadSuccessEvent(info UserInfo, client string, featureName FeatureName, now TimeGetter) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Type:        DownloadSuccess,
		Timestamp:   now(),
	}
}

func NewDownloadFailureEvent(info UserInfo, client string, featureName FeatureName, errorKind DownloadErrorKind, errorMessage string, now TimeGetter) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Type:        DownloadFailure,
		Timestamp:   now(),
		EventDetails: EventDetails{
			Error:   errorKind.String(),
			Message: errorMessage,
		},
	}
}

func NewLocalUseEvent(info UserInfo, client string, featureName FeatureName, now TimeGetter) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Type:        LocalUse,
		Timestamp:   now(),
	}
}

func NewJSONParseEvent(info UserInfo, client string, featureName FeatureName, errorKind, errorMessage string, now TimeGetter) Event {
	details := EventDetails{}
	var eventType EventType
	if errorKind != "" {
		details.Error = errorKind
		details.Message = errorMessage
		eventType = JSONParseFailure
	} else {
		eventType = JSONParseSuccess
	}

	return Event{
		UserInfo:     info,
		Client:       client,
		FeatureName:  featureName.String(),
		Type:         eventType,
		Timestamp:    now(),
		EventDetails: details,
	}
}

func NewPartialRolloutEvent(info UserInfo, client string, featureName FeatureName, eventError error, featureRollout int, now TimeGetter) Event {
	//message format is slightly different for partial rollout events
	payload := struct {
		FeatureName    string `json:"feature_name"`
		RolloutGroup   int    `json:"rollout_group"`
		FeatureRollout int    `json:"feature_rollout"`
	}{
		FeatureName:    featureName.String(),
		RolloutGroup:   info.RolloutGroup,
		FeatureRollout: featureRollout,
	}

	messageBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("%s%s Failed to marshal partial rollout message: %s. Defaulting to an empty\n", internal.WarningPrefix, logPrefix, err)
		messageBytes = []byte{}
	}

	details := EventDetails{
		Message: string(messageBytes),
	}

	if eventError != nil {
		details.Error = eventError.Error()
	}

	return Event{
		UserInfo:     info,
		Client:       client,
		FeatureName:  featureName.String(),
		Type:         Rollout,
		Timestamp:    now(),
		EventDetails: details,
	}
}

func (e Event) ToMooseDebuggerEvent() *events.MooseDebuggerEvent {
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
		log.Printf("%s%s Failed to marshal Event to JSON %s. Fallback to a limited set of data\n", internal.WarningPrefix, logPrefix, err)
		jsonData = []byte(
			`{"namespace":"` + messageNamespace + `",` +
				`"subscope":"` + subscope + `",` +
				`"result":"` + eventToMarshal.Result + `",` +
				`"type":"` + e.Type.String() + `",` +
				`"feature_name":"` + e.FeatureName + `",` +
				`"client":"` + e.Client + `",` +
				`"error":"` + e.Error + `",` +
				`"message":"` + e.Message + `",` +
				`"app_version":"` + e.AppVersion + `",` +
				`"country":"` + e.Country + `",` +
				`"isp":"` + e.ISP + `",` +
				`"rollout_group":` + fmt.Sprintf("%v", e.RolloutGroup) + `}`,
		)
	}
	return events.NewMooseDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: debuggerEventBaseKey + ".type", Value: e.Type},
			events.ContextValue{Path: debuggerEventBaseKey + ".app_version", Value: e.AppVersion},
			events.ContextValue{Path: debuggerEventBaseKey + ".country", Value: e.Country},
			events.ContextValue{Path: debuggerEventBaseKey + ".isp", Value: e.ISP},
			events.ContextValue{Path: debuggerEventBaseKey + ".error", Value: e.Error},
			events.ContextValue{Path: debuggerEventBaseKey + ".feature_name", Value: e.FeatureName},
			events.ContextValue{Path: debuggerEventBaseKey + ".rollout_group", Value: e.RolloutGroup},
		).
		WithGlobalContextPaths("device.*", "application.nordvpnapp.*", "application.nordvpnapp.version", "application.nordvpnapp.platform")
}
