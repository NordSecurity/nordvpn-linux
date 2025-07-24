package remote

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventJSONOutput(t *testing.T) {
	category.Set(t, category.Unit)
	ctx := UserInfo{
		AppVersion:   "1.2.3",
		Country:      "XX",
		ISP:          "Super Duper ISP",
		RolloutGroup: 42,
	}

	testCases := []struct {
		name            string
		event           Event
		expectedResult  string
		expectedError   string
		expectedMessage string
	}{
		{
			name:            "PartialRollout Success",
			event:           NewPartialRolloutEvent(ctx, "test-client", FeatureMeshnet, nil, 50, time.Now),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: `{"feature_name":"meshnet","rollout_group":42,"feature_rollout":50}`,
		},
		{
			name:            "PartialRollout Failure",
			event:           NewPartialRolloutEvent(ctx, "test-client", FeatureMeshnet, fmt.Errorf("config-error"), 50, time.Now),
			expectedResult:  rcFailure,
			expectedError:   "config-error",
			expectedMessage: `{"feature_name":"meshnet","rollout_group":42,"feature_rollout":50}`,
		},
		{
			name:            "Download Success",
			event:           NewDownloadSuccessEvent(ctx, "client", FeatureMeshnet, time.Now),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
		{
			name:            "Download Failure",
			event:           NewDownloadFailureEvent(ctx, "client", FeatureMain, DownloadErrorNetwork, "timeout", time.Now),
			expectedResult:  rcFailure,
			expectedError:   DownloadErrorNetwork.String(),
			expectedMessage: "timeout",
		},
		{
			name:            "JSON Parse Success",
			event:           NewJSONParseEvent(ctx, "client", FeatureLibtelio, "", "", time.Now),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
		{
			name:            "JSON Parse Failure",
			event:           NewJSONParseEvent(ctx, "client", FeatureLibtelio, "syntax-error", "bad token", time.Now),
			expectedResult:  rcFailure,
			expectedError:   "syntax-error",
			expectedMessage: "bad token",
		},
		{
			name:            "Local Use",
			event:           NewLocalUseEvent(ctx, "client", FeatureLibtelio, time.Now),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mooseEvent := tc.event.ToMooseDebuggerEvent()
			var decodedEvent Event
			err := json.Unmarshal([]byte(mooseEvent.JsonData), &decodedEvent)
			require.NoError(t, err, "JSON should be valid and parsable into remote-config event")

			assert.Equal(t, messageNamespace, decodedEvent.MessageNamespace, "namespace should be correctly set")
			assert.Equal(t, subscope, decodedEvent.Subscope, "subscope should be correctly set")
			assert.Equal(t, tc.event.Type, decodedEvent.Type, "event type should match")
			assert.Equal(t, tc.expectedResult, decodedEvent.Result, "result should match expected")
			assert.Equal(t, tc.expectedError, decodedEvent.Error, "error message should match expected")
			assert.Equal(t, tc.expectedMessage, decodedEvent.Message, "message content should match expected")
		})
	}
}

// TestMooseDebuggerEventContextPaths verifies that the ToMooseDebuggerEvent method
// correctly generates MooseDebuggerEvent objects with the expected context paths.
// It tests that:
//  1. The GeneralContextPaths includes device and application information paths
//  2. The KeyBasedContextPaths contains all the relevant remote config download failure
//     information including type, app version, country, ISP, error, feature name and rollout group
func TestMooseDebuggerEventContextPaths(t *testing.T) {
	category.Set(t, category.Unit)

	ctx := UserInfo{
		AppVersion:   "3.1.0",
		Country:      "Testland",
		ISP:          "TestISP",
		RolloutGroup: 42,
	}
	event := NewDownloadFailureEvent(ctx, "test-client", FeatureLibtelio, DownloadErrorNetwork, "timeout", time.Now).ToMooseDebuggerEvent()

	expectedGeneralPaths := []string{
		"device.*",
		"application.nordvpnapp.*",
		"application.nordvpnapp.version",
		"application.nordvpnapp.platform",
	}
	assert.ElementsMatch(t, expectedGeneralPaths, event.GeneralContextPaths)

	// Assert: Verify the KeyBased context paths.
	expectedKeyBasedPaths := []events.ContextValue{
		{Path: debuggerEventBaseKey + ".type", Value: DownloadFailure},
		{Path: debuggerEventBaseKey + ".app_version", Value: "3.1.0"},
		{Path: debuggerEventBaseKey + ".country", Value: "Testland"},
		{Path: debuggerEventBaseKey + ".isp", Value: "TestISP"},
		{Path: debuggerEventBaseKey + ".error", Value: DownloadErrorNetwork.String()},
		{Path: debuggerEventBaseKey + ".feature_name", Value: "libtelio"},
		{Path: debuggerEventBaseKey + ".rollout_group", Value: 42},
	}
	assert.ElementsMatch(t, expectedKeyBasedPaths, event.KeyBasedContextPaths)
}
