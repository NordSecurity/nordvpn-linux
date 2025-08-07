package remote

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type UserInfo struct {
	AppVersion   string
	Country      string
	ISP          string
	RolloutGroup int
}

// EventDetails holds optional details for an event.
type EventDetails struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Event is the main analytics event structure.
type Event struct {
	UserInfo `json:"-"`
	EventDetails
	MessageNamespace string `json:"namespace"`
	Subscope         string `json:"subscope"`
	Client           string `json:"client"`
	Event            string `json:"event"`
	Result           string `json:"result"`
	FeatureName      string `json:"-"`
	RolloutPerformed bool   `json:"-"`
}

func NewDownloadSuccessEvent(info UserInfo, client string, featureName FeatureName) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Event:       DownloadSuccess.String(),
	}
}

func NewDownloadFailureEvent(info UserInfo, client string, featureName FeatureName, errorKind DownloadErrorKind, errorMessage string) Event {
	return Event{
		UserInfo:    info,
		Client:      client,
		FeatureName: featureName.String(),
		Event:       DownloadFailure.String(),
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
		Event:       LocalUse.String(),
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
		Event:        eventType.String(),
		EventDetails: details,
	}
}

func NewRolloutEvent(info UserInfo, client string, featureName FeatureName, featureRollout int, rolloutPerformed bool) Event {
	details := EventDetails{
		Error:   fmt.Sprintf("%s %d / %d", featureName.String(), info.RolloutGroup, featureRollout),
		Message: featureName.String(),
	}

	return Event{
		UserInfo:         info,
		Client:           client,
		FeatureName:      featureName.String(),
		Event:            Rollout.String(),
		EventDetails:     details,
		RolloutPerformed: rolloutPerformed,
	}
}

func (e Event) ToMooseDebuggerEvent() *events.MooseDebuggerEvent {
	eventToMarshal := e
	eventToMarshal.MessageNamespace = messageNamespace
	eventToMarshal.Subscope = subscope
	if e.Event == Rollout.String() {
		// rollout events have a different result values: yes|no
		// while other events have: success|failure
		if e.RolloutPerformed {
			eventToMarshal.Result = rolloutYes
		} else {
			eventToMarshal.Result = rolloutNo
		}
	} else {
		if e.Error != "" {
			eventToMarshal.Result = rcFailure
		} else {
			eventToMarshal.Result = rcSuccess
		}
	}

	jsonData, err := json.Marshal(eventToMarshal)
	if err != nil {
		//in case of any marshalling error, let's at least provide basic information, we know anyway
		log.Printf("%s%s Failed to marshal Event to JSON %s. Fallback to a limited set of data\n", internal.WarningPrefix, logPrefix, err)

		fallbackData := struct {
			MessageNamespace string `json:"namespace"`
			Subscope         string `json:"subscope"`
			Client           string `json:"client"`
			Event            string `json:"event"`
			Result           string `json:"result"`
			Error            string `json:"error"`
			FeatureName      string `json:"feature_name"`
			RolloutGroup     int    `json:"rollout_group"`
		}{
			MessageNamespace: messageNamespace,
			Subscope:         subscope,
			Client:           e.Client,
			Event:            e.Event,
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
			events.ContextValue{Path: debuggerEventBaseKey + ".type", Value: e.Event},
			events.ContextValue{Path: debuggerEventBaseKey + ".app_version", Value: e.AppVersion},
			events.ContextValue{Path: debuggerEventBaseKey + ".country", Value: e.Country},
			events.ContextValue{Path: debuggerEventBaseKey + ".isp", Value: e.ISP},
			events.ContextValue{Path: debuggerEventBaseKey + ".error", Value: e.Error},
			events.ContextValue{Path: debuggerEventBaseKey + ".feature_name", Value: e.FeatureName},
			events.ContextValue{Path: debuggerEventBaseKey + ".rollout_group", Value: e.RolloutGroup},
		).
		WithGlobalContextPaths("device.*", "application.nordvpnapp.*", "application.nordvpnapp.version", "application.nordvpnapp.platform")
}
