package remote

import (
	"sync"
	"testing"
	"time"

	ev "github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const testRolloutGroup = 42

// TestEventFlagsSet tests the eventFlagsSet functionality
func TestEventFlagsSet(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name   string
		setup  func(flags eventFlagsSet)
		verify func(t *testing.T, flags eventFlagsSet)
	}{
		{
			name: "set and has operations for single feature",
			setup: func(flags eventFlagsSet) {
				flags.set(FeatureMeshnet, Rollout)
				flags.set(FeatureMeshnet, Download)
			},
			verify: func(t *testing.T, flags eventFlagsSet) {
				assert.True(t, flags.has(FeatureMeshnet, Rollout))
				assert.True(t, flags.has(FeatureMeshnet, Download))
				assert.False(t, flags.has(FeatureMeshnet, LocalUse))
			},
		},
		{
			name: "set and has operations for multiple features",
			setup: func(flags eventFlagsSet) {
				flags.set(FeatureMeshnet, Rollout)
				flags.set(FeatureLibtelio, Rollout)
				flags.set(FeatureMeshnet, Download)
				flags.set(FeatureMain, LocalUse)
			},
			verify: func(t *testing.T, flags eventFlagsSet) {
				// Each feature tracks its own events independently
				assert.True(t, flags.has(FeatureMeshnet, Rollout))
				assert.True(t, flags.has(FeatureLibtelio, Rollout))
				assert.False(t, flags.has(FeatureMain, Rollout))

				assert.True(t, flags.has(FeatureMeshnet, Download))
				assert.False(t, flags.has(FeatureLibtelio, Download))

				assert.True(t, flags.has(FeatureMain, LocalUse))
				assert.False(t, flags.has(FeatureMeshnet, LocalUse))
			},
		},
		{
			name: "clear operation removes all flags for all features",
			setup: func(flags eventFlagsSet) {
				flags.set(FeatureMeshnet, Rollout)
				flags.set(FeatureLibtelio, Download)
				flags.set(FeatureMain, LocalUse)
				flags.clear()
			},
			verify: func(t *testing.T, flags eventFlagsSet) {
				assert.False(t, flags.has(FeatureMeshnet, Rollout))
				assert.False(t, flags.has(FeatureLibtelio, Download))
				assert.False(t, flags.has(FeatureMain, LocalUse))
			},
		},
		{
			name: "empty flags return false for all checks",
			setup: func(flags eventFlagsSet) {
				// No setup needed - testing empty state
			},
			verify: func(t *testing.T, flags eventFlagsSet) {
				assert.False(t, flags.has(FeatureMeshnet, Rollout))
				assert.False(t, flags.has(FeatureMeshnet, Download))
				assert.False(t, flags.has(FeatureMeshnet, LocalUse))
				assert.False(t, flags.has(FeatureLibtelio, JSONParseSuccess))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := make(eventFlagsSet)
			tt.setup(flags)
			tt.verify(t, flags)
		})
	}
}

// TestEmitPartialRolloutEvent_Deduplication tests that rollout events are deduplicated per feature
func TestEmitPartialRolloutEvent_Deduplication(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		actions        []func(analytics Analytics)
		expectedEvents int
		verifyEvent    func(t *testing.T, events []string)
	}{
		{
			name: "first rollout event is emitted",
			actions: []func(analytics Analytics){
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true) },
			},
			expectedEvents: 1,
			verifyEvent: func(t *testing.T, events []string) {
				assert.Contains(t, events[0], `"event":"rc_rollout"`)
				assert.Contains(t, events[0], `"result":"yes"`)
			},
		},
		{
			name: "subsequent rollout events for same feature are not emitted",
			actions: []func(analytics Analytics){
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 60, true) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 30, false) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 100, true) },
			},
			expectedEvents: 1,
			verifyEvent: func(t *testing.T, events []string) {
				// Only first event should be present
				assert.Contains(t, events[0], `"rollout_info":"meshnet 42 / app 50"`)
			},
		},
		{
			name: "different features can emit rollout events independently",
			actions: []func(analytics Analytics){
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureLibtelio, 60, true) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMain, 70, false) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 80, true) }, // duplicate, should not emit
			},
			expectedEvents: 3,
			verifyEvent: func(t *testing.T, events []string) {
				assert.Contains(t, events[0], `"feature_name":"meshnet"`)
				assert.Contains(t, events[1], `"feature_name":"libtelio"`)
				assert.Contains(t, events[2], `"feature_name":"nordvpn"`)
			},
		},
		{
			name: "rollout event emitted after ClearEventFlags",
			actions: []func(analytics Analytics){
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true) },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 60, false) },
				func(a Analytics) { a.ClearEventFlags() },
				func(a Analytics) { a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 70, true) },
			},
			expectedEvents: 2,
			verifyEvent: func(t *testing.T, events []string) {
				assert.Contains(t, events[0], `"rollout_info":"meshnet 42 / app 50"`)
				assert.Contains(t, events[1], `"rollout_info":"meshnet 42 / app 70"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(testRolloutGroup)
			fixture.subscriber.ExpectEvents(tt.expectedEvents)

			for _, action := range tt.actions {
				action(fixture.analytics)
			}

			fixture.subscriber.Wait(t)
			assert.Len(t, fixture.subscriber.events, tt.expectedEvents)
			if tt.verifyEvent != nil {
				tt.verifyEvent(t, fixture.subscriber.events)
			}
		})
	}
}

// TestOtherAnalyticsEvents_NotAffectedByDeduplication tests that other event types
// are not affected by the deduplication logic
func TestOtherAnalyticsEvents_NotAffectedByDeduplication(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		eventType     string
		emitEvents    func(analytics Analytics)
		expectedCount int
		expectedEvent string
	}{
		{
			name:      "download events are not deduplicated",
			eventType: "download",
			emitEvents: func(a Analytics) {
				a.EmitDownloadEvent(ClientCli, FeatureMeshnet)
				a.EmitDownloadEvent(ClientCli, FeatureMeshnet)
				a.EmitDownloadEvent(ClientCli, FeatureMeshnet)
			},
			expectedCount: 3,
			expectedEvent: `"event":"rc_download_success"`,
		},
		{
			name:      "local use events are not deduplicated",
			eventType: "local_use",
			emitEvents: func(a Analytics) {
				a.EmitLocalUseEvent(ClientCli, FeatureMeshnet, nil)
				a.EmitLocalUseEvent(ClientCli, FeatureMeshnet, nil)
			},
			expectedCount: 2,
			expectedEvent: `"event":"rc_local_use"`,
		},
		{
			name:      "download failure events are not deduplicated",
			eventType: "download_failure",
			emitEvents: func(a Analytics) {
				err := *NewDownloadError(DownloadErrorFileDownload, nil)
				a.EmitDownloadFailureEvent(ClientCli, FeatureMeshnet, err)
				a.EmitDownloadFailureEvent(ClientCli, FeatureMeshnet, err)
			},
			expectedCount: 2,
			expectedEvent: `"event":"rc_download_failure"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(testRolloutGroup)
			fixture.subscriber.ExpectEvents(tt.expectedCount)

			tt.emitEvents(fixture.analytics)

			fixture.subscriber.Wait(t)
			assert.Len(t, fixture.subscriber.events, tt.expectedCount)

			for _, event := range fixture.subscriber.events {
				assert.Contains(t, event, tt.expectedEvent)
			}
		})
	}
}

// TestMultipleFeatures_Deduplication tests rollout event deduplication across different features
func TestMultipleFeatures_Deduplication(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		actions []struct {
			feature    string
			rollout    int
			shouldEmit bool
		}
		clearAfterIndex int // -1 means no clear
		expectedEvents  int
		verifyEvents    func(t *testing.T, events []string)
	}{
		{
			name: "each feature can emit one rollout event independently",
			actions: []struct {
				feature    string
				rollout    int
				shouldEmit bool
			}{
				{feature: FeatureMeshnet, rollout: 50, shouldEmit: true},
				{feature: FeatureLibtelio, rollout: 60, shouldEmit: true},
				{feature: FeatureMain, rollout: 70, shouldEmit: true},
				{feature: FeatureMeshnet, rollout: 80, shouldEmit: false}, // duplicate for meshnet
			},
			clearAfterIndex: -1,
			expectedEvents:  3,
			verifyEvents: func(t *testing.T, events []string) {
				assert.Contains(t, events[0], `"feature_name":"meshnet"`)
				assert.Contains(t, events[0], `"rollout_info":"meshnet 42 / app 50"`)
				assert.Contains(t, events[1], `"feature_name":"libtelio"`)
				assert.Contains(t, events[1], `"rollout_info":"libtelio 42 / app 60"`)
				assert.Contains(t, events[2], `"feature_name":"nordvpn"`)
				assert.Contains(t, events[2], `"rollout_info":"nordvpn 42 / app 70"`)
			},
		},
		{
			name: "clear allows all features to emit again",
			actions: []struct {
				feature    string
				rollout    int
				shouldEmit bool
			}{
				{feature: FeatureMeshnet, rollout: 50, shouldEmit: true},
				{feature: FeatureLibtelio, rollout: 60, shouldEmit: true},
				{feature: FeatureMeshnet, rollout: 55, shouldEmit: false}, // duplicate before clear
				{feature: FeatureMeshnet, rollout: 80, shouldEmit: true},  // After clear
				{feature: FeatureLibtelio, rollout: 90, shouldEmit: true}, // After clear
			},
			clearAfterIndex: 2, // Clear after third action
			expectedEvents:  4,
			verifyEvents: func(t *testing.T, events []string) {
				// First two events before clear
				assert.Contains(t, events[0], `"rollout_info":"meshnet 42 / app 50"`)
				assert.Contains(t, events[1], `"rollout_info":"libtelio 42 / app 60"`)
				// Two events after clear
				assert.Contains(t, events[2], `"rollout_info":"meshnet 42 / app 80"`)
				assert.Contains(t, events[3], `"rollout_info":"libtelio 42 / app 90"`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(testRolloutGroup)
			fixture.subscriber.ExpectEvents(tt.expectedEvents)

			for i, action := range tt.actions {
				fixture.analytics.EmitPartialRolloutEvent(ClientCli, action.feature, action.rollout, true)
				if tt.clearAfterIndex == i {
					fixture.analytics.ClearEventFlags()
				}
			}

			fixture.subscriber.Wait(t)
			assert.Len(t, fixture.subscriber.events, tt.expectedEvents)
			if tt.verifyEvents != nil {
				tt.verifyEvents(t, fixture.subscriber.events)
			}
		})
	}
}

// TestEventAlreadyEmitted tests the internal eventAlreadyEmitted method
func TestEventAlreadyEmitted(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		setup        func(rca *RemoteConfigAnalytics)
		checkFeature string
		checkType    EventType
		expected     bool
	}{
		{
			name:         "initially no events are marked as emitted",
			setup:        func(rca *RemoteConfigAnalytics) {},
			checkFeature: FeatureMeshnet,
			checkType:    Rollout,
			expected:     false,
		},
		{
			name: "rollout marked as emitted for specific feature",
			setup: func(rca *RemoteConfigAnalytics) {
				rca.eventFlags.set(FeatureMeshnet, Rollout)
			},
			checkFeature: FeatureMeshnet,
			checkType:    Rollout,
			expected:     true,
		},
		{
			name: "rollout for different feature not affected",
			setup: func(rca *RemoteConfigAnalytics) {
				rca.eventFlags.set(FeatureMeshnet, Rollout)
			},
			checkFeature: FeatureLibtelio,
			checkType:    Rollout,
			expected:     false,
		},
		{
			name: "other event types not affected",
			setup: func(rca *RemoteConfigAnalytics) {
				rca.eventFlags.set(FeatureMeshnet, Rollout)
			},
			checkFeature: FeatureMeshnet,
			checkType:    Download,
			expected:     false,
		},
		{
			name: "clear removes all flags",
			setup: func(rca *RemoteConfigAnalytics) {
				rca.eventFlags.set(FeatureMeshnet, Rollout)
				rca.eventFlags.set(FeatureLibtelio, Download)
				rca.ClearEventFlags()
			},
			checkFeature: FeatureMeshnet,
			checkType:    Rollout,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(testRolloutGroup)
			rca, ok := fixture.analytics.(*RemoteConfigAnalytics)
			assert.True(t, ok, "analytics should be of type *RemoteConfigAnalytics")

			tt.setup(rca)
			result := rca.eventAlreadyEmitted(tt.checkFeature, tt.checkType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConcurrentEventEmission tests thread safety of the deduplication mechanism
func TestConcurrentEventEmission(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		concurrentEmits   int
		expectedEvents    int
		includeOtherTypes bool
		multipleFeatures  bool
	}{
		{
			name:            "10 concurrent rollout attempts for same feature emit 1 event",
			concurrentEmits: 10,
			expectedEvents:  1,
		},
		{
			name:              "mixed event types with concurrent rollouts",
			concurrentEmits:   5,
			expectedEvents:    6, // 1 rollout + 5 download events
			includeOtherTypes: true,
		},
		{
			name:             "concurrent rollouts for different features",
			concurrentEmits:  3,
			expectedEvents:   3, // Each feature gets its own event
			multipleFeatures: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom event collector that doesn't use WaitGroup
			var mu sync.Mutex
			var collectedEvents []string

			publisher := &MockDebuggerEvents{}
			publisher.Subscribe(func(e ev.DebuggerEvent) error {
				mu.Lock()
				defer mu.Unlock()
				collectedEvents = append(collectedEvents, e.JsonData)
				return nil
			})

			analytics := NewRemoteConfigAnalytics(publisher, testRolloutGroup)

			// Launch concurrent goroutines
			var wg sync.WaitGroup
			wg.Add(tt.concurrentEmits)

			features := []string{FeatureMeshnet, FeatureLibtelio, FeatureMain}

			for i := 0; i < tt.concurrentEmits; i++ {
				go func(idx int) {
					defer wg.Done()
					feature := FeatureMeshnet
					if tt.multipleFeatures {
						feature = features[idx%len(features)]
					}
					analytics.EmitPartialRolloutEvent(ClientCli, feature, 40+idx, true)
					if tt.includeOtherTypes {
						// These should not be deduplicated
						analytics.EmitDownloadEvent(ClientCli, feature)
					}
				}(i)
			}

			wg.Wait()

			// Give a small delay for events to be processed
			time.Sleep(10 * time.Millisecond)

			// Check the collected events
			mu.Lock()
			eventCount := len(collectedEvents)
			mu.Unlock()

			assert.Equal(t, tt.expectedEvents, eventCount, "Expected %d events but got %d", tt.expectedEvents, eventCount)
		})
	}
}

// TestClearEventFlags_Scenarios tests various scenarios with ClearEventFlags
func TestClearEventFlags_Scenarios(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		scenario      func(analytics Analytics) []int // returns expected events at each step
		expectedTotal int
	}{
		{
			name: "clear between different event types",
			scenario: func(a Analytics) []int {
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true) // 1 event
				a.EmitDownloadEvent(ClientCli, FeatureMeshnet)                 // 2 events
				a.ClearEventFlags()
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 60, true) // 3 events
				a.EmitLocalUseEvent(ClientCli, FeatureMeshnet, nil)            // 4 events
				return []int{1, 2, 3, 4}
			},
			expectedTotal: 4,
		},
		{
			name: "multiple clears allow multiple rollout events for same feature",
			scenario: func(a Analytics) []int {
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 30, true) // 1 event
				a.ClearEventFlags()
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 40, true) // 2 events
				a.ClearEventFlags()
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true) // 3 events
				return []int{1, 2, 3}
			},
			expectedTotal: 3,
		},
		{
			name: "clear affects all features",
			scenario: func(a Analytics) []int {
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 30, true)  // 1 event
				a.EmitPartialRolloutEvent(ClientCli, FeatureLibtelio, 40, true) // 2 events
				a.EmitPartialRolloutEvent(ClientCli, FeatureMain, 50, true)     // 3 events
				a.ClearEventFlags()
				a.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 60, true)  // 4 events
				a.EmitPartialRolloutEvent(ClientCli, FeatureLibtelio, 70, true) // 5 events
				a.EmitPartialRolloutEvent(ClientCli, FeatureMain, 80, true)     // 6 events
				return []int{1, 2, 3, 4, 5, 6}
			},
			expectedTotal: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(testRolloutGroup)
			fixture.subscriber.ExpectEvents(tt.expectedTotal)

			tt.scenario(fixture.analytics)

			fixture.subscriber.Wait(t)
			assert.Len(t, fixture.subscriber.events, tt.expectedTotal)
		})
	}
}

// TestPerFeatureDeduplication tests that deduplication works independently per feature
func TestPerFeatureDeduplication(t *testing.T) {
	category.Set(t, category.Unit)

	fixture := setupAnalyticsTest(testRolloutGroup)

	// Expect 6 events total
	fixture.subscriber.ExpectEvents(6)

	// Emit for meshnet - should succeed
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 50, true)

	// Emit for libtelio - should succeed (different feature)
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureLibtelio, 60, true)

	// Emit for meshnet again - should be deduplicated
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 55, false)

	// Emit for nordvpn - should succeed (different feature)
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureMain, 70, true)

	// Clear flags
	fixture.analytics.ClearEventFlags()

	// Now all features can emit again
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureMeshnet, 80, true)
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureLibtelio, 85, true)
	fixture.analytics.EmitPartialRolloutEvent(ClientCli, FeatureMain, 90, true)

	fixture.subscriber.Wait(t)

	events := fixture.subscriber.events
	assert.Len(t, events, 6)

	// Verify the order and content
	assert.Contains(t, events[0], `"feature_name":"meshnet"`)
	assert.Contains(t, events[0], `"rollout_info":"meshnet 42 / app 50"`)

	assert.Contains(t, events[1], `"feature_name":"libtelio"`)
	assert.Contains(t, events[1], `"rollout_info":"libtelio 42 / app 60"`)

	assert.Contains(t, events[2], `"feature_name":"nordvpn"`)
	assert.Contains(t, events[2], `"rollout_info":"nordvpn 42 / app 70"`)

	// After clear
	assert.Contains(t, events[3], `"feature_name":"meshnet"`)
	assert.Contains(t, events[3], `"rollout_info":"meshnet 42 / app 80"`)

	assert.Contains(t, events[4], `"feature_name":"libtelio"`)
	assert.Contains(t, events[4], `"rollout_info":"libtelio 42 / app 85"`)

	assert.Contains(t, events[5], `"feature_name":"nordvpn"`)
	assert.Contains(t, events[5], `"rollout_info":"nordvpn 42 / app 90"`)
}
