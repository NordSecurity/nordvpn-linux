package firstopen_test

import (
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/firstopen"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestNotifyOnceAppJustInstalled(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name               string
		newInstallation    bool // config.NewInstallation
		numCalls           int  // how many times to call NotifyOnceAppJustInstalled
		wantPublishCount   int  // expected total publishes
		checkActionDetails bool // whether to verify the fields of the published action
	}{
		{
			name:             "no publish when not new installation",
			newInstallation:  false,
			numCalls:         1,
			wantPublishCount: 0,
		},
		{
			name:               "publish exactly once when first call",
			newInstallation:    true,
			numCalls:           1,
			wantPublishCount:   1,
			checkActionDetails: true,
		},
		{
			name:               "multiple calls still publish only once",
			newInstallation:    true,
			numCalls:           5,
			wantPublishCount:   1,
			checkActionDetails: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cm := &config.FilesystemConfigManager{NewInstallation: tc.newInstallation}
			pub := &stubPublisher{}
			n := firstopen.NewNotifier(cm, pub)

			for i := 0; i < tc.numCalls; i++ {
				err := n.NotifyOnceAppJustInstalled(core.Insights{})
				assert.NilError(t, err, "unexpected error on call %d", i+1)
			}

			got := pub.count()
			assert.Equal(t, got, tc.wantPublishCount, "expected %d publishes, got %d", tc.wantPublishCount, got)

			if tc.checkActionDetails && tc.wantPublishCount > 0 {
				action := pub.last()
				want := events.UiItemsAction{
					ItemName:      "first_open",
					ItemType:      "button",
					ItemValue:     "first_open",
					FormReference: "daemon",
				}
				assert.DeepEqual(t, action, want)
			}
		})
	}
}

func TestConcurrency_NotifyOnceAppJustInstalled_IsThreadSafe(t *testing.T) {
	category.Set(t, category.Unit)
	cm := &config.FilesystemConfigManager{NewInstallation: true}
	pub := &stubPublisher{}
	n := firstopen.NewNotifier(cm, pub)

	var wg sync.WaitGroup
	const numGoroutines = 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := n.NotifyOnceAppJustInstalled(core.Insights{})
			assert.NilError(t, err, "unexpected error in goroutine")
		}()
	}
	wg.Wait()

	got := pub.count()
	assert.Equal(t, got, 1, "expected exactly 1 publish under concurrency, got %d", got)
}

type stubPublisher struct {
	mu      sync.Mutex
	actions []events.UiItemsAction
}

func (s *stubPublisher) Publish(a events.UiItemsAction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actions = append(s.actions, a)
}

func (s *stubPublisher) Subscribe(_ events.Handler[events.UiItemsAction]) {}

func (s *stubPublisher) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.actions)
}

func (s *stubPublisher) last() events.UiItemsAction {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.actions[len(s.actions)-1]
}
