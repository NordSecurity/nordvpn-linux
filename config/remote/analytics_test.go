package remote

import (
	"errors"
	"sync"
	"testing"
	"time"

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

type analyticsTestFixture struct {
	publisher  *MockDebuggerEvents
	subscriber *MockSubscriber
	analytics  Analytics
}

func setupAnalyticsTest() *analyticsTestFixture {
	publisher := &MockDebuggerEvents{}
	subscriber := NewMockListener()
	publisher.Subscribe(subscriber.NotifyDebuggerEvent)

	analytics := NewRemoteConfigAnalytics(publisher, "1.2.3", 42)

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
	}{
		{
			name: "NotifyDownload success",
			action: func(a Analytics) {
				a.NotifyDownload(client, feature)
			},
			expectedEventName: `"event":"rc_download_success"`,
		},
		{
			name: "NotifyDownload failure",
			action: func(a Analytics) {
				a.NotifyDownloadFailure(client, feature, DownloadError{Kind: DownloadErrorFileDownload, Cause: errors.New("fail")})
			},
			expectedEventName: `"event":"rc_download_failure"`,
			expectedDetails:   `"message":"file_download_error: fail"`,
		},
		{
			name: "NotifyLocalUse",
			action: func(a Analytics) {
				a.NotifyLocalUse(client, feature, nil)
			},
			expectedEventName: `"event":"rc_local_use"`,
		},
		{
			name: "NotifyJsonParse success",
			action: func(a Analytics) {
				a.NotifyJsonParse(client, feature, nil)
			},
			expectedEventName: `"event":"rc_json_parse_success"`,
		},
		{
			name: "NotifyJsonParse failure",
			action: func(a Analytics) {
				a.NotifyJsonParse(client, feature, errors.New("parse error"))
			},
			expectedEventName: `"event":"rc_json_parse_failure"`,
			expectedDetails:   `"message":"parse error"`,
		},
		{
			name: "NotifyPartialRollout",
			action: func(a Analytics) {
				a.NotifyPartialRollout(client, feature, 7, true)
			},
			expectedEventName: `"event":"rc_rollout"`,
			expectedDetails:   `"error":"meshnet 42 / 7"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := setupAnalyticsTest()
			fixture.subscriber.ExpectEvents(1)

			tc.action(fixture.analytics)
			fixture.subscriber.Wait(t)

			assert.Len(t, fixture.subscriber.events, 1)
			event := fixture.subscriber.events[0]
			assert.Contains(t, event, tc.expectedEventName)
			if tc.expectedDetails != "" {
				assert.Contains(t, event, tc.expectedDetails)
			}
		})
	}
}
