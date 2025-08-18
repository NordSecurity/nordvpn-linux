package remote

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Event is the main analytics event structure.
type Event struct {
	Client           string `json:"client"`
	Error            string `json:"error"`
	Event            string `json:"event"`
	FeatureName      string `json:"feature_name"`
	Message          string `json:"message"`
	MessageNamespace string `json:"namespace"`
	Result           string `json:"result"`
	RolloutGroup     int    `json:"-"`
	RolloutInfo      string `json:"rollout_info"`
	RolloutPerformed bool   `json:"-"`
	Subscope         string `json:"subscope"`
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
		Error:        errorKind.String(),
		Message:      errorMessage,
	}
}

func NewLocalUseEvent(userRolloutGroup int, client, featureName, errorKind, errorMessage string) Event {
	return Event{
		RolloutGroup: userRolloutGroup,
		Client:       client,
		FeatureName:  featureName,
		Event:        LocalUse.String(),
		Error:        errorKind,
		Message:      errorMessage,
	}
}

func NewJSONParseEventFailure(userRolloutGroup int, client string, featureName string, errorKind LoadErrorKind, errorMessage string) Event {
	return Event{
		RolloutGroup: userRolloutGroup,
		Client:       client,
		FeatureName:  featureName,
		Event:        JSONParseFailure.String(),
		Error:        errorKind.String(),
		Message:      errorMessage,
	}
}

func NewRolloutEvent(userRolloutGroup int, client string, featureName string, featureRollout int, rolloutPerformed bool) Event {
	return Event{
		RolloutGroup:     userRolloutGroup,
		Client:           client,
		FeatureName:      featureName,
		Event:            Rollout.String(),
		RolloutInfo:      fmt.Sprintf("%s %d / app %d", featureName, userRolloutGroup, featureRollout),
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
