package remote

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// EventDetails holds optional details for an event.
type EventDetails struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Event is the main analytics event structure.
type Event struct {
	EventDetails
	RolloutGroup     int    `json:"-"`
	MessageNamespace string `json:"namespace"`
	Subscope         string `json:"subscope"`
	Client           string `json:"client"`
	Event            string `json:"event"`
	Result           string `json:"result"`
	FeatureName      string `json:"-"`
	RolloutPerformed bool   `json:"-"`
}

func NewDownloadSuccessEvent(userRolloutGroup int, client string, featureName string) Event {
	return Event{
		RolloutGroup: userRolloutGroup,
		Client:       client,
		FeatureName:  featureName,
		Event:        DownloadSuccess.String(),
	}
}

func NewDownloadFailureEvent(userRolloutGroup int, client string, featureName string, errorKind DownloadErrorKind, errorMessage string) Event {
	return Event{
		RolloutGroup: userRolloutGroup,
		Client:       client,
		FeatureName:  featureName,
		Event:        DownloadFailure.String(),
		EventDetails: EventDetails{
			Error:   errorKind.String(),
			Message: errorMessage,
		},
	}
}

func NewLocalUseEvent(userRolloutGroup int, client, featureName, errorKind, errorMessage string) Event {
	details := EventDetails{}
	if errorKind != "" {
		details.Error = errorKind
		details.Message = errorMessage
	}
	return Event{
		RolloutGroup: userRolloutGroup,
		Client:       client,
		FeatureName:  featureName,
		Event:        LocalUse.String(),
		EventDetails: details,
	}
}

func NewJSONParseEventFailure(userRolloutGroup int, client string, featureName string, errorKind LoadErrorKind, errorMessage string) Event {
	details := EventDetails{
		Error:   errorKind.String(),
		Message: errorMessage,
	}

	return Event{
		RolloutGroup: userRolloutGroup,
		Client:       client,
		FeatureName:  featureName,
		Event:        JSONParseFailure.String(),
		EventDetails: details,
	}
}

func NewRolloutEvent(userRolloutGroup int, client string, featureName string, featureRollout int, rolloutPerformed bool) Event {
	details := EventDetails{
		Error:   fmt.Sprintf("%s %d / %d", featureName, userRolloutGroup, featureRollout),
		Message: featureName,
	}

	return Event{
		RolloutGroup:     userRolloutGroup,
		Client:           client,
		FeatureName:      featureName,
		Event:            Rollout.String(),
		EventDetails:     details,
		RolloutPerformed: rolloutPerformed,
	}
}

func (e Event) ToDebuggerEvent() *events.DebuggerEvent {
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
	return events.NewDebuggerEvent(string(jsonData)).
		WithKeyBasedContextPaths(
			events.ContextValue{Path: debuggerEventBaseKey + ".type", Value: e.Event},
			events.ContextValue{Path: debuggerEventBaseKey + ".error", Value: e.Error},
			events.ContextValue{Path: debuggerEventBaseKey + ".feature_name", Value: e.FeatureName},
			events.ContextValue{Path: debuggerEventBaseKey + ".rollout_group", Value: e.RolloutGroup},
		).
		WithGlobalContextPaths("device.*", "application.nordvpnapp.*", "application.nordvpnapp.version", "application.nordvpnapp.platform")
}
