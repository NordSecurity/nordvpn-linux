package firstopen

import (
	"slices"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func Test_NoPublish_WhenNotNewInstallation(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: false}
	sub := &stubSubscriber{}
	pub := &stubPublisher{}
	guard := &sync.Once{}

	registerNotifier(cm, sub, pub, guard)

	sub.Publish(core.Insights{})

	assert.Equal(t, pub.count(), 0)
}

func Test_PublishesExactlyOnce_OnFirstEvent(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	pub := &stubPublisher{}
	guard := &sync.Once{}

	registerNotifier(cm, sub, pub, guard)

	sub.Publish(core.Insights{})
	assert.Equal(t, pub.count(), 1)

	// second install event - no new publish
	sub.Publish(core.Insights{})
	assert.Equal(t, pub.count(), 1)

	want := events.UiItemsAction{
		ItemName:      "first_open",
		ItemType:      "button",
		ItemValue:     "first_open",
		FormReference: "daemon",
	}
	assert.DeepEqual(t, pub.last(), want)
}

func Test_NoResubscribe_AfterPublished(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	pub := &stubPublisher{}
	guard := &sync.Once{}

	registerNotifier(cm, sub, pub, guard)
	sub.Publish(core.Insights{})
	assert.Equal(t, pub.count(), 1)

	// a second registration attempt should not add another handler
	registerNotifier(cm, sub, pub, guard)
	assert.Equal(t, len(sub.handlers), 1)

	// even if we fire again, still only one publish
	sub.Publish(core.Insights{})
	assert.Equal(t, pub.count(), 1)
}

func Test_MultipleNotifiers_WithSingleGuard(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	pub := &stubPublisher{}

	// two registrations, but both share defaultGuard
	RegisterNotifier(cm, sub, pub)
	RegisterNotifier(cm, sub, pub)

	sub.Publish(core.Insights{})
	// only one publish despite two handlers
	assert.Equal(t, pub.count(), 1)

	// next emits are no-ops
	sub.Publish(core.Insights{})
	assert.Equal(t, pub.count(), 1)
}

func Test_ConcurrentInstallEvents_OnlyOnePublish(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	pub := &stubPublisher{}
	guard := &sync.Once{}

	registerNotifier(cm, sub, pub, guard)

	var wg sync.WaitGroup
	const n = 20
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			sub.Publish(core.Insights{})
		}()
	}
	wg.Wait()

	assert.Equal(t, pub.count(), 1, "expected exactly 1 publish")
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

type stubSubscriber struct {
	mu       sync.Mutex
	handlers []events.Handler[core.Insights]
}

func (s *stubSubscriber) Subscribe(h events.Handler[core.Insights]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, h)
}

func (s *stubSubscriber) Publish(ins core.Insights) {
	s.mu.Lock()
	hs := slices.Clone(s.handlers)
	s.mu.Unlock()

	for _, h := range hs {
		_ = h(ins)
	}
}

func resetGlobals() {
	published.Store(false)
}
