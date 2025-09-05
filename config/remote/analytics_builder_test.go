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
	rolloutGroup := 42

	testCases := []struct {
		name                string
		event               Event
		expectedResult      string
		expectedError       string
		expectedMessage     string
		expectedRolloutInfo string
		expectedFeatureName string
	}{
		{
			name:                "Rollout Success",
			event:               NewRolloutEvent(rolloutGroup, "test-client", FeatureMeshnet, 50, true),
			expectedResult:      rolloutYes,
			expectedRolloutInfo: "meshnet 42 / app 50",
			expectedFeatureName: FeatureMeshnet,
		},
		{
			name:                "Rollout Failure",
			event:               NewRolloutEvent(rolloutGroup, "test-client", FeatureMeshnet, 20, false),
			expectedResult:      rolloutNo,
			expectedRolloutInfo: "meshnet 42 / app 20",
			expectedFeatureName: FeatureMeshnet,
		},
		{
			name:                "Download Success",
			event:               NewDownloadSuccessEvent(rolloutGroup, "client", FeatureMeshnet),
			expectedResult:      rcSuccess,
			expectedError:       "",
			expectedMessage:     "",
			expectedFeatureName: FeatureMeshnet,
		},
		{
			name:                "JSON Parse Failure",
			event:               NewJSONParseEventFailure(rolloutGroup, "client", FeatureLibtelio, LoadErrorMainHashJsonParsing, "bad hash"),
			expectedResult:      rcFailure,
			expectedError:       LoadErrorMainHashJsonParsing.String(),
			expectedMessage:     "bad hash",
			expectedFeatureName: FeatureLibtelio,
		},
		{
			name:                "Local Use",
			event:               NewLocalUseEvent(rolloutGroup, "client", FeatureLibtelio, "", ""),
			expectedResult:      rcSuccess,
			expectedError:       "",
			expectedMessage:     "",
			expectedFeatureName: FeatureLibtelio,
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
			assert.Equal(t, tc.expectedRolloutInfo, decodedEvent.RolloutInfo, "rollout info should match expected")
			assert.Equal(t, tc.expectedFeatureName, decodedEvent.FeatureName, "feature name should match expected")
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

	rolloutGroup := 42
	debugerEvent := NewDownloadFailureEvent(rolloutGroup, "test-client", FeatureLibtelio, DownloadErrorIntegrity, "timeout").ToDebuggerEvent()

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
		{Path: debuggerEventBaseKey + ".error", Value: DownloadErrorIntegrity.String()},
		{Path: debuggerEventBaseKey + ".feature_name", Value: FeatureLibtelio},
		{Path: debuggerEventBaseKey + ".rollout_group", Value: 42},
	}
	assert.ElementsMatch(t, expectedKeyBasedPaths, debugerEvent.KeyBasedContextPaths)
}

func TestDebuggerEventContainsOnlyDesignedFields(t *testing.T) {
	category.Set(t, category.Unit)
	rolloutGroup := 99

	event := NewDownloadFailureEvent(rolloutGroup, "test-env", FeatureLibtelio, DownloadErrorIntegrity, "Integrity corrupted")
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
		"feature_name",
		"rollout_info",
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
