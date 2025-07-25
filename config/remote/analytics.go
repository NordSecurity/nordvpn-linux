package remote

import (
	"encoding/json"
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

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
}

func NewDownloadSuccessEvent(info UserInfo, client string, featureName FeatureName) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Type:        DownloadSuccess,
	}
}

func NewDownloadFailureEvent(info UserInfo, client string, featureName FeatureName, errorKind DownloadErrorKind, errorMessage string) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Type:        DownloadFailure,
		EventDetails: EventDetails{
			Error:   errorKind.String(),
			Message: errorMessage,
		},
	}
}

func NewLocalUseEvent(info UserInfo, client string, featureName FeatureName) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Type:        LocalUse,
	}
}

func NewJSONParseEvent(info UserInfo, client string, featureName FeatureName, errorKind, errorMessage string) Event {
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
		EventDetails: details,
	}
}

func NewPartialRolloutEvent(info UserInfo, client string, featureName FeatureName, eventError error, featureRollout int) Event {
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
		//in case of any marshalling error, let's at least provide basic information, we know anyway
		log.Printf("%s%s Failed to marshal Event to JSON %s. Fallback to a limited set of data\n", internal.WarningPrefix, logPrefix, err)

		fallbackData := struct {
			MessageNamespace string    `json:"namespace"`
			Subscope         string    `json:"subscope"`
			Client           string    `json:"client"`
			Type             EventType `json:"type"`
			Result           string    `json:"result"`
			Error            string    `json:"error"`
			FeatureName      string    `json:"feature_name"`
			RolloutGroup     int       `json:"rollout_group"`
		}{
			MessageNamespace: messageNamespace,
			Subscope:         subscope,
			Client:           e.Client,
			Type:             e.Type,
			Result:           eventToMarshal.Result,
			Error:            err.Error(),
			FeatureName:      e.FeatureName,
			RolloutGroup:     e.RolloutGroup,
		}

		jsonData, err = json.Marshal(fallbackData)
		//this should never happen...
		if err != nil {
			log.Printf("%s%s Failed to marshal fallback data to JSON %s. Defaulting to empty JSON\n", internal.ErrorPrefix, logPrefix, err)
			jsonData = []byte(`{}`)
		}
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
