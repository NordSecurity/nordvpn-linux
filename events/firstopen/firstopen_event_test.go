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
	count := 0
	testNotifier := func() error {
		count++
		return nil
	}

	RegisterNotifier(cm, sub, testNotifier)

	sub.Publish(core.Insights{})

	assert.Equal(t, count, 0)
}

func Test_PublishesExactlyOnce_OnFirstEvent(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	count := 0
	testNotifier := func() error {
		count++
		return nil
	}

	RegisterNotifier(cm, sub, testNotifier)

	sub.Publish(core.Insights{})
	assert.Equal(t, count, 1)

	// second install event - no new publish
	sub.Publish(core.Insights{})
	assert.Equal(t, count, 1)
}

func Test_NoResubscribe_AfterPublished(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	count := 0
	testNotifier := func() error {
		count++
		return nil
	}

	RegisterNotifier(cm, sub, testNotifier)
	sub.Publish(core.Insights{})
	assert.Equal(t, count, 1)

	// a second registration attempt should not add another handler
	RegisterNotifier(cm, sub, testNotifier)
	assert.Equal(t, len(sub.handlers), 1)

	// even if we fire again, still only one publish
	sub.Publish(core.Insights{})
	assert.Equal(t, count, 1)
}

func Test_WithMultipleRegistrations_EventIsEmittedOnlyOnce(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	count := 0
	testNotifier := func() error {
		count++
		return nil
	}

	RegisterNotifier(cm, sub, testNotifier)
	RegisterNotifier(cm, sub, testNotifier)
	RegisterNotifier(cm, sub, testNotifier)

	sub.Publish(core.Insights{})
	assert.Equal(t, count, 1)
}

func Test_ConcurrentInstallEvents_OnlyOnePublish(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	cm := &config.FilesystemConfigManager{NewInstallation: true}
	sub := &stubSubscriber{}
	count := 0
	testNotifier := func() error {
		count++
		return nil
	}

	RegisterNotifier(cm, sub, testNotifier)
	var wg sync.WaitGroup
	const n = 20
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			sub.Publish(core.Insights{})
		}()
	}
	wg.Wait()

	assert.Equal(t, count, 1, "expected exactly 1 publish")
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
