package remote

import (
	"encoding/json"
	"testing"

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
			name:            "Rollout Success",
			event:           NewRolloutEvent(ctx, "test-client", FeatureMeshnet, 50, true),
			expectedResult:  rolloutYes,
			expectedError:   "meshnet 42 / 50",
			expectedMessage: FeatureMeshnet,
		},
		{
			name:            "Rollout Failure",
			event:           NewRolloutEvent(ctx, "test-client", FeatureMeshnet, 50, false),
			expectedResult:  rolloutNo,
			expectedError:   "meshnet 42 / 50",
			expectedMessage: FeatureMeshnet,
		},
		{
			name:            "Download Success",
			event:           NewDownloadSuccessEvent(ctx, "client", FeatureMeshnet),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
		{
			name:            "Download Failure",
			event:           NewDownloadFailureEvent(ctx, "client", FeatureMain, DownloadErrorNetwork, "timeout"),
			expectedResult:  rcFailure,
			expectedError:   DownloadErrorNetwork.String(),
			expectedMessage: "timeout",
		},
		{
			name:            "JSON Parse Success",
			event:           NewJSONParseEvent(ctx, "client", FeatureLibtelio, "", ""),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
		{
			name:            "JSON Parse Failure",
			event:           NewJSONParseEvent(ctx, "client", FeatureLibtelio, "syntax-error", "bad token"),
			expectedResult:  rcFailure,
			expectedError:   "syntax-error",
			expectedMessage: "bad token",
		},
		{
			name:            "Local Use",
			event:           NewLocalUseEvent(ctx, "client", FeatureLibtelio),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			debugerEvent := tc.event.ToDebuggerEvent()
			var decodedEvent Event
			err := json.Unmarshal([]byte(debugerEvent.JsonData), &decodedEvent)
			require.NoError(t, err, "JSON should be valid and parsable into remote-config event")

			assert.Equal(t, messageNamespace, decodedEvent.MessageNamespace, "namespace should be correctly set")
			assert.Equal(t, subscope, decodedEvent.Subscope, "subscope should be correctly set")
			assert.Equal(t, tc.event.Event, decodedEvent.Event, "event type should match")
			assert.Equal(t, tc.expectedResult, decodedEvent.Result, "result should match expected")
			assert.Equal(t, tc.expectedError, decodedEvent.Error, "error message should match expected")
			assert.Equal(t, tc.expectedMessage, decodedEvent.Message, "message content should match expected")
		})
	}
}

// TestDebuggerEventContextPaths verifies that the ToDebuggerEvent method
// correctly generates DebuggerEvent objects with the expected context paths.
// It tests that:
//  1. The GeneralContextPaths includes device and application information paths
//  2. The KeyBasedContextPaths contains all the relevant remote config download failure
//     information including type, app version, country, ISP, error, feature name and rollout group
func TestDebuggerEventContextPaths(t *testing.T) {
	category.Set(t, category.Unit)

	ctx := UserInfo{
		AppVersion:   "3.1.0",
		Country:      "Testland",
		ISP:          "TestISP",
		RolloutGroup: 42,
	}
	debugerEvent := NewDownloadFailureEvent(ctx, "test-client", FeatureLibtelio, DownloadErrorNetwork, "timeout").ToDebuggerEvent()

	expectedGeneralPaths := []string{
		"device.*",
		"application.nordvpnapp.*",
		"application.nordvpnapp.version",
		"application.nordvpnapp.platform",
	}
	assert.ElementsMatch(t, expectedGeneralPaths, debugerEvent.GeneralContextPaths)

	// Assert: Verify the KeyBased context paths.
	expectedKeyBasedPaths := []events.ContextValue{
		{Path: debuggerEventBaseKey + ".type", Value: DownloadFailure.String()},
		{Path: debuggerEventBaseKey + ".app_version", Value: "3.1.0"},
		{Path: debuggerEventBaseKey + ".country", Value: "Testland"},
		{Path: debuggerEventBaseKey + ".isp", Value: "TestISP"},
		{Path: debuggerEventBaseKey + ".error", Value: DownloadErrorNetwork.String()},
		{Path: debuggerEventBaseKey + ".feature_name", Value: FeatureLibtelio},
		{Path: debuggerEventBaseKey + ".rollout_group", Value: 42},
	}
	assert.ElementsMatch(t, expectedKeyBasedPaths, debugerEvent.KeyBasedContextPaths)
}

func TestDebuggerEventContainsOnlyDesignedFields(t *testing.T) {
	category.Set(t, category.Unit)
	ctx := UserInfo{
		AppVersion:   "9.8.7",
		Country:      "AA",
		ISP:          "Testland ISP",
		RolloutGroup: 99,
	}

	event := NewDownloadFailureEvent(ctx, "test-env", FeatureLibtelio, DownloadErrorIntegrity, "Integrity corrupted")
	debugerEvent := event.ToDebuggerEvent()

	var payload map[string]interface{}
	err := json.Unmarshal([]byte(debugerEvent.JsonData), &payload)
	require.NoError(t, err, "JSON should be valid")

	// Define the expected set of keys
	expectedKeys := []string{
		"namespace",
		"subscope",
		"client",
		"event",
		"result",
		"error",
		"message",
	}

	// Check that the keys match exactly
	var actualKeys []string
	for k := range payload {
		actualKeys = append(actualKeys, k)
	}
	assert.ElementsMatch(t, expectedKeys, actualKeys, "JSON fields should match expected set")

	// Optionally, check that the names are correct and values are as expected
	assert.Equal(t, messageNamespace, payload["namespace"])
	assert.Equal(t, subscope, payload["subscope"])
	assert.Equal(t, "test-env", payload["client"])
	assert.Equal(t, DownloadFailure.String(), payload["event"])
	assert.Equal(t, rcFailure, payload["result"])
	assert.Equal(t, DownloadErrorIntegrity.String(), payload["error"])
	assert.Equal(t, "Integrity corrupted", payload["message"])
}
