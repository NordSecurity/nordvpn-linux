package remote

import (
	"errors"
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

// The tests verify that MooseAnalytics correctly publishes analytics events
// for various scenarios, including:
// - Successful and failed downloads
// - Local feature usage
// - Successful and failed JSON parsing
// - Partial rollout notifications
// Each test checks that the appropriate event name and details are included
// in the published event data.
func TestMooseAnalytics(t *testing.T) {
	category.Set(t, category.Unit)

	const (
		client  = "cli"
		feature = "meshnet"
	)

	testCases := []struct {
		name              string
		action            func(a Analytics)
		expectedEventName string
		expectedDetails   string
		expectedResult    string
	}{
		{
			name: "EmitDownloadEvent success",
			action: func(a Analytics) {
				a.EmitDownloadEvent(client, feature)
			},
			expectedEventName: `"event":"rc_download_success"`,
			expectedResult:    `"result":"success"`,
		},
		{
			name: "EmitDownloadEvent failure",
			action: func(a Analytics) {
				a.EmitDownloadFailureEvent(client, feature, DownloadError{Kind: DownloadErrorFileDownload, Cause: errors.New("fail")})
			},
			expectedEventName: `"event":"rc_download_failure"`,
			expectedDetails:   `"message":"file_download_error: fail"`,
			expectedResult:    `"result":"failure"`,
		},
		{
			name: "EmitLocalUseEvent",
			action: func(a Analytics) {
				a.EmitLocalUseEvent(client, feature, nil)
			},
			expectedEventName: `"event":"rc_local_use"`,
			expectedResult:    `"result":"success"`,
		},
		{
			name: "EmitLocalUseEvent_failure",
			action: func(a Analytics) {
				a.EmitLocalUseEvent(client, feature, errors.New("local-use-test-error"))
			},
			expectedEventName: `"event":"rc_local_use"`,
			expectedDetails:   `"message":"local-use-test-error"`,
			expectedResult:    `"result":"failure"`,
		},
		{
			name: "EmitJsonParseEvent success",
			action: func(a Analytics) {
				a.EmitJsonParseEvent(client, feature, nil)
			},
			expectedEventName: `"event":"rc_json_parse_success"`,
			expectedResult:    `"result":"success"`,
		},
		{
			name: "EmitJsonParseEvent failure",
			action: func(a Analytics) {
				a.EmitJsonParseEvent(client, feature, errors.New("parse error"))
			},
			expectedEventName: `"event":"rc_json_parse_failure"`,
			expectedDetails:   `"message":"parse error"`,
			expectedResult:    `"result":"failure"`,
		},
		{
			name: "EmitPartialRolloutEvent",
			action: func(a Analytics) {
				a.EmitPartialRolloutEvent(client, feature, 7, true)
			},
			expectedEventName: `"event":"rc_rollout"`,
			expectedDetails:   `"error":"meshnet 42 / 7"`,
			//rollout uses different expected results - yes|no
			expectedResult: `"result":"yes"`,
		},
		{
			name: "EmitPartialRolloutEvent failure",
			action: func(a Analytics) {
				a.EmitPartialRolloutEvent(client, feature, 7, false)
			},
			expectedEventName: `"event":"rc_rollout"`,
			expectedDetails:   `"error":"meshnet 42 / 7"`,
			//rollout uses different expected results - yes|no
			expectedResult: `"result":"no"`,
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
		})
	}
}

// TestFindMatchingRecord_EmitsOneEvent tests that findMatchingRecord emits exactly one
// EmitPartialRolloutEvent when called through IsFeatureEnabled
func TestFindMatchingRecord_EmitsOneEvent(t *testing.T) {
	category.Set(t, category.Unit)

	const (
		featureName = FeatureMeshnet
		rollout     = 50
	)

	testCases := []struct {
		name              string
		rolloutGroup      int
		expectedEventName string
		expectedDetails   string
		expectedResult    string
	}{
		{
			name:              "Partiall Rollout failed - rollout group above limit",
			rolloutGroup:      60,
			expectedEventName: `"event":"rc_rollout"`,
			expectedDetails:   `"error":"meshnet 60 / 50"`,
			expectedResult:    `"result":"no"`,
		},
		{
			name:              "Rollout success - rollout group under the limit",
			rolloutGroup:      40,
			expectedEventName: `"event":"rc_rollout"`,
			expectedDetails:   `"error":"meshnet 40 / 50"`,
			expectedResult:    `"result":"yes"`,
		},
		{
			name:              "Rollout success - rollout group same as the limit",
			rolloutGroup:      50,
			expectedEventName: `"event":"rc_rollout"`,
			expectedDetails:   `"error":"meshnet 50 / 50"`,
			expectedResult:    `"result":"no"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupAnalyticsTest(tc.rolloutGroup)

			fixture.subscriber.ExpectEvents(1)

			rc := NewCdnRemoteConfig(
				config.BuildTarget{Version: "1.2.3", Environment: "test"},
				"/remote/path",
				"/local/path",
				&MockRemoteStorage{},
				fixture.analytics,
				tc.rolloutGroup,
			)

			feature := rc.features.get(featureName)
			feature.params = map[string]*Param{
				featureName: {
					Type: fieldTypeBool,
					Settings: []ParamValue{
						{
							AppVersion: "^1.0.0",
							Value:      true,
							Weight:     100,
							Rollout:    rollout,
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
			assert.Contains(t, event, tc.expectedDetails)
			assert.Contains(t, event, tc.expectedResult)
		})
	}
}
