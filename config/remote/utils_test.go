package remote

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGroupDistribution verifies that the distribution of groups created by GenerateRolloutGroup is uniform.
// The test generates an arbitrary number of 10kk random UUIDs and assigns them to groups using GenerateRolloutGroup,
// then checks that the distribution of groups across five equal-sized brackets is statistically
// balanced. It validates:
//  1. Each bracket (1-20, 21-40, 41-60, 61-80, 81-100) contains approximately 20% of the groups
//  2. The deviation from the expected distribution is within a 2% tolerance
//  3. The standard deviation is within acceptable bounds (±2σ)
//  4. The total count of groups matches the number of iterations
func TestGroupDistribution(t *testing.T) {
	category.Set(t, category.Unit)
	const (
		bracketSize        = 20
		iterations         = 10_000_000
		expectedCount      = iterations / 5
		expectedPercentage = 20.0 // 100% / 5 brackets
		tolerance          = 0.02 // 2% tolerance
	)

	uuid.EnableRandPool()
	defer uuid.DisableRandPool()

	groups := make([]int, iterations)
	for i := 0; i < iterations; i++ {
		id := uuid.New()
		g := GenerateRolloutGroup(id)
		groups[i] = g
	}

	// calculate specific brackets counts
	brackets := make([]int, 5)
	for _, g := range groups {
		bracketIndex := (g - 1) / bracketSize
		brackets[bracketIndex]++
	}

	diffs := make([]float64, len(brackets))
	sum := 0
	for i, count := range brackets {
		sum += count
		diffs[i] = float64(count - expectedCount)
	}

	sumSquares := 0.0
	for _, diff := range diffs {
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(brackets)))

	bracketNames := []string{"1-20", "21-40", "41-60", "61-80", "81-100"}

	for i, count := range brackets {
		percentage := float64(count) / float64(iterations) * 100
		deviations := diffs[i] / stdDev

		t.Logf("Bracket %-8s: count=%d (%.2f%%) diff=%+.0f stddev=%+.2f σ",
			bracketNames[i], count, percentage, diffs[i], deviations)

		assert.InDelta(t, expectedPercentage, percentage, tolerance*100,
			"Bracket %s percentage outside tolerance", bracketNames[i])

		assert.InDelta(t, 0, math.Abs(deviations), 2.0,
			"Bracket %s deviation outside tolerance", bracketNames[i])
	}

	// Assert total count matches iterations
	assert.Equal(t, iterations, sum,
		"Sum of bracket counts (%d) doesn't match total number of generated UUIDs (%d)", sum, iterations)
}

func TestRCEventJSONOutput(t *testing.T) {
	category.Set(t, category.Unit)
	ctx := Context{
		AppVersion:   "1.2.3",
		Country:      "XX",
		ISP:          "Super Duper ISP",
		RolloutGroup: 42,
	}

	testCases := []struct {
		name            string
		event           RCEvent
		expectedResult  string
		expectedError   string
		expectedMessage string
	}{
		{
			name:            "PartialRollout Success",
			event:           NewRCPartialRollout(ctx, "test-client", "new-ui", "", 50),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: `{"feature.name":"new-ui","rollout_group":42,"feature.rollout":50}`,
		},
		{
			name:            "PartialRollout Failure",
			event:           NewRCPartialRollout(ctx, "test-client", "new-ui", "config-error", 50),
			expectedResult:  rcFailure,
			expectedError:   "config-error",
			expectedMessage: `{"feature.name":"new-ui","rollout_group":42,"feature.rollout":50}`,
		},
		{
			name:            "Download Success",
			event:           NewRCDownloadSuccess(ctx, "client", "feature"),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
		{
			name:            "Download Failure",
			event:           NewRCDownloadFailure(ctx, "client", "feature", RCDownloadErrorNetwork, "timeout"),
			expectedResult:  rcFailure,
			expectedError:   string(RCDownloadErrorNetwork),
			expectedMessage: "timeout",
		},
		{
			name:            "JSON Parse Success",
			event:           NewRCJSONParse(ctx, "client", "feature", "", ""),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
		{
			name:            "JSON Parse Failure",
			event:           NewRCJSONParse(ctx, "client", "feature", "syntax-error", "bad token"),
			expectedResult:  rcFailure,
			expectedError:   "syntax-error",
			expectedMessage: "bad token",
		},
		{
			name:            "Local Use",
			event:           NewRCLocalUse(ctx, "client", "feature"),
			expectedResult:  rcSuccess,
			expectedError:   "",
			expectedMessage: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mooseEvent := tc.event.ToMooseDebuggerEvent()
			var decodedEvent RCEvent
			err := json.Unmarshal([]byte(mooseEvent.JsonData), &decodedEvent)
			require.NoError(t, err, "JSON should be valid and parsable into RCEvent")

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

	ctx := Context{
		AppVersion:   "3.1.0",
		Country:      "Testland",
		ISP:          "TestISP",
		RolloutGroup: 42,
	}
	event := NewRCDownloadFailure(ctx, "test-client", "test-feature", RCDownloadErrorNetwork, "timeout")

	// Act: Generate the MooseDebuggerEvent.
	mooseEvent := event.ToMooseDebuggerEvent()

	// Assert: Verify the Global/General context paths.
	expectedGeneralPaths := []string{
		"device.*",
		"application.nordvpnapp.*",
		"application.nordvpnapp.version",
		"application.nordvpnapp.platform",
	}
	assert.ElementsMatch(t, expectedGeneralPaths, mooseEvent.GeneralContextPaths)

	// Assert: Verify the KeyBased context paths.
	expectedKeyBasedPaths := []events.ContextValue{
		{Path: baseKey + ".type", Value: string(RCDownloadFailure)},
		{Path: baseKey + ".app_version", Value: "3.1.0"},
		{Path: baseKey + ".country", Value: "Testland"},
		{Path: baseKey + ".isp", Value: "TestISP"},
		{Path: baseKey + ".error", Value: string(RCDownloadErrorNetwork)},
		{Path: baseKey + ".feature_name", Value: "test-feature"},
		{Path: baseKey + ".rollout_group", Value: 42},
	}
	assert.ElementsMatch(t, expectedKeyBasedPaths, mooseEvent.KeyBasedContextPaths)
}
