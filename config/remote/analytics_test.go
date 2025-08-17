package remote

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	ev "github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

// MockSubscriber is a test implementation of an event listener.
// It captures published debugger events into a slice for later assertion.
// It uses a sync.WaitGroup for synchronization means.
type MockSubscriber struct {
	mu     sync.Mutex
	events []string // store event JSON for assertion
	wg     *sync.WaitGroup
}

func NewMockListener() *MockSubscriber {
	return &MockSubscriber{
		events: make([]string, 0),
		wg:     &sync.WaitGroup{},
	}
}

func (s *MockSubscriber) NotifyDebuggerEvent(e ev.DebuggerEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, e.JsonData)
	s.wg.Done()
	return nil
}

func (s *MockSubscriber) ExpectEvents(count int) {
	s.wg.Add(count)
}

func (s *MockSubscriber) Wait(t *testing.T) {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for DebuggerEvents")
	}
}

type MockDebuggerEvents struct {
	subscribers []ev.Handler[ev.DebuggerEvent]
}

func (m *MockDebuggerEvents) Subscribe(s ev.Handler[ev.DebuggerEvent]) {
	m.subscribers = append(m.subscribers, s)
}

func (m *MockDebuggerEvents) Publish(e ev.DebuggerEvent) {
	for _, sub := range m.subscribers {
		sub(e)
	}
}

type MockRemoteStorage struct{}

func (m *MockRemoteStorage) GetRemoteFile(string) ([]byte, error) {
	return nil, nil
}

type analyticsTestFixture struct {
	publisher  *MockDebuggerEvents
	subscriber *MockSubscriber
	analytics  Analytics
}

func setupAnalyticsTest(rolloutGroup int) *analyticsTestFixture {
	publisher := &MockDebuggerEvents{}
	subscriber := NewMockListener()
	publisher.Subscribe(subscriber.NotifyDebuggerEvent)

	analytics := NewRemoteConfigAnalytics(publisher, rolloutGroup)

	return &analyticsTestFixture{
		publisher:  publisher,
		subscriber: subscriber,
		analytics:  analytics,
	}
}

// The tests verify that Analytics correctly publishes analytics events
// for various scenarios, including:
// - Successful and failed downloads
// - Local feature usage
// - Successful and failed JSON parsing
// - Partial rollout notifications
// Each test checks that the appropriate event name and details are included
// in the published event data.
func TestAnalytics(t *testing.T) {
	category.Set(t, category.Unit)

	const (
		client  = "cli"
		feature = "meshnet"
	)

	testCases := []struct {
		name                string
		action              func(a Analytics)
		expectedEventName   string
		expectedDetails     string
		expectedResult      string
		expectedRolloutInfo string
		expectedFeatureName string
	}{
		{
			name: "EmitDownloadEvent success",
			action: func(a Analytics) {
				a.EmitDownloadEvent(client, feature)
			},
			expectedEventName: fmt.Sprintf(`"event":"%s"`, DownloadSuccess),
			expectedResult:    fmt.Sprintf(`"result":"%s"`, rcSuccess),
		},
		{
			name: "EmitDownloadEvent failure",
			action: func(a Analytics) {
				a.EmitDownloadFailureEvent(client, feature, *NewDownloadError(DownloadErrorFileDownload, errors.New("fail")))
			},
			expectedEventName: fmt.Sprintf(`"event":"%s"`, DownloadFailure),
			expectedDetails:   `"message":"file_download_error: fail"`,
			expectedResult:    fmt.Sprintf(`"result":"%s"`, rcFailure),
		},
		{
			name: "EmitLocalUseEvent",
			action: func(a Analytics) {
				a.EmitLocalUseEvent(client, feature, nil)
			},
			expectedEventName: fmt.Sprintf(`"event":"%s"`, LocalUse),
			expectedResult:    fmt.Sprintf(`"result":"%s"`, rcSuccess),
		},
		{
			name: "EmitLocalUseEvent_failure",
			action: func(a Analytics) {
				a.EmitLocalUseEvent(client, feature, errors.New("local-use-test-error"))
			},
			expectedEventName: fmt.Sprintf(`"event":"%s"`, LocalUse),
			expectedDetails:   `"message":"local-use-test-error"`,
			expectedResult:    fmt.Sprintf(`"result":"%s"`, rcFailure),
		},
		{
			name: "EmitJsonParseEvent failure",
			action: func(a Analytics) {
				a.EmitJsonParseFailureEvent(client, feature, *NewLoadError(LoadErrorValidation, errors.New("parse error")))
			},
			expectedEventName: fmt.Sprintf(`"event":"%s"`, JSONParseFailure),
			expectedDetails:   `"message":"validation_error: parse error"`,
			expectedResult:    fmt.Sprintf(`"result":"%s"`, rcFailure),
		},
		{
			name: "EmitPartialRolloutEvent",
			action: func(a Analytics) {
				a.EmitPartialRolloutEvent(client, feature, 52, true)
			},
			expectedEventName:   fmt.Sprintf(`"event":"%s"`, Rollout),
			expectedRolloutInfo: `"rollout_info":"meshnet 42 / app 52"`,
			expectedFeatureName: `"feature_name":"meshnet"`,
			//rollout uses different expected results - yes|no
			expectedResult: fmt.Sprintf(`"result":"%s"`, rolloutYes),
		},
		{
			name: "EmitPartialRolloutEvent failure",
			action: func(a Analytics) {
				a.EmitPartialRolloutEvent(client, feature, 7, false)
			},
			expectedEventName:   fmt.Sprintf(`"event":"%s"`, Rollout),
			expectedRolloutInfo: `"rollout_info":"meshnet 42 / app 7"`,
			expectedFeatureName: `"feature_name":"meshnet"`,
			//rollout uses different expected results - yes|no
			expectedResult: fmt.Sprintf(`"result":"%s"`, rolloutNo),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(42)
			fixture.subscriber.ExpectEvents(1)

			tc.action(fixture.analytics)
			fixture.subscriber.Wait(t)

			assert.Len(t, fixture.subscriber.events, 1)
			event := fixture.subscriber.events[0]
			assert.Contains(t, event, tc.expectedEventName)
			if tc.expectedDetails != "" {
				assert.Contains(t, event, tc.expectedDetails)
			}

			if tc.expectedResult != "" {
				assert.Contains(t, event, tc.expectedResult)
			}

			if tc.expectedRolloutInfo != "" {
				assert.Contains(t, event, tc.expectedRolloutInfo)
			}

			if tc.expectedFeatureName != "" {
				assert.Contains(t, event, tc.expectedFeatureName)
			}
		})
	}
}

// TestFindMatchingRecord_EmitsOneEvent tests that findMatchingRecord emits exactly one
// EmitPartialRolloutEvent when called through IsFeatureEnabled
func TestFindMatchingRecord_EmitsOneEvent(t *testing.T) {
	category.Set(t, category.Unit)

	const (
		featureName    = FeatureMeshnet
		featureRollout = 50
	)

	testCases := []struct {
		name                string
		userRolloutGroup    int
		expectedEventName   string
		expectedRolloutInfo string
		expectedResult      string
		expectedFeatureName string
	}{
		{
			name:                "Partiall Rollout failed - rollout group above limit",
			userRolloutGroup:    60,
			expectedEventName:   `"event":"rc_rollout"`,
			expectedRolloutInfo: `"rollout_info":"meshnet 60 / app 50"`,
			expectedResult:      `"result":"no"`,
			expectedFeatureName: `"feature_name":"meshnet"`,
		},
		{
			name:                "Rollout success - rollout group under the limit",
			userRolloutGroup:    40,
			expectedEventName:   `"event":"rc_rollout"`,
			expectedRolloutInfo: `"rollout_info":"meshnet 40 / app 50"`,
			expectedResult:      `"result":"yes"`,
			expectedFeatureName: `"feature_name":"meshnet"`,
		},
		{
			name:                "Rollout success - rollout group same as the limit",
			userRolloutGroup:    50,
			expectedEventName:   `"event":"rc_rollout"`,
			expectedRolloutInfo: `"rollout_info":"meshnet 50 / app 50"`,
			expectedResult:      `"result":"yes"`,
			expectedFeatureName: `"feature_name":"meshnet"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(tc.userRolloutGroup)

			fixture.subscriber.ExpectEvents(1)

			rc := NewCdnRemoteConfig(
				config.BuildTarget{Version: "1.2.3", Environment: "test"},
				"/remote/path",
				"/local/path",
				&MockRemoteStorage{},
				fixture.analytics,
				tc.userRolloutGroup,
			)

			feature := rc.features.get(featureName)
			feature.params = map[string]*Param{
				featureName: {
					Type: fieldTypeBool,
					Settings: []ParamValue{
						{
							AppVersion:    "^1.0.0",
							Value:         true,
							Weight:        100,
							TargetRollout: featureRollout,
						},
					},
				},
			}

			// Call IsFeatureEnabled, which will trigger findMatchingRecord
			rc.IsFeatureEnabled(featureName)

			fixture.subscriber.Wait(t)

			assert.Len(t, fixture.subscriber.events, 1,
				"Expected exactly one analytics event to be emitted")

			event := fixture.subscriber.events[0]
			assert.Contains(t, event, tc.expectedEventName)
			assert.Contains(t, event, tc.expectedRolloutInfo)
			assert.Contains(t, event, tc.expectedResult)
			assert.Contains(t, event, tc.expectedFeatureName)
		})
	}
}
