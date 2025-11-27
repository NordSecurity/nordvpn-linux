package firstopen

import (
	"slices"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

type testFixture struct {
	count             int
	fileConfigManager *config.FilesystemConfigManager
	notifier          func() error
	subscriber        *stubSubscriber
}

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()
	f := testFixture{
		count:             0,
		fileConfigManager: &config.FilesystemConfigManager{NewInstallation: true},
		subscriber:        &stubSubscriber{},
	}
	f.notifier = func() error {
		f.count++
		return nil
	}
	return &f
}

func Test_NoPublish_WhenNotNewInstallation(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	tf := newTestFixture(t)
	tf.fileConfigManager.NewInstallation = false
	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)

	tf.subscriber.Publish(struct{}{})

	assert.Equal(t, tf.count, 0)
}

func Test_PublishesExactlyOnce_OnFirstEvent(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	tf := newTestFixture(t)
	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)

	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)

	tf.subscriber.Publish(struct{}{})
	assert.Equal(t, tf.count, 1)

	// second install event - no new publish
	tf.subscriber.Publish(struct{}{})
	assert.Equal(t, tf.count, 1)
}

func Test_NoResubscribe_AfterPublished(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	tf := newTestFixture(t)

	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)
	tf.subscriber.Publish(struct{}{})
	assert.Equal(t, tf.count, 1)
	// a second registration attempt should not add another handler
	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)
	assert.Equal(t, len(tf.subscriber.handlers), 1)

	// even if we fire again, still only one publish
	tf.subscriber.Publish(struct{}{})
	assert.Equal(t, tf.count, 1)
}

func Test_WithMultipleRegistrations_EventIsEmittedOnlyOnce(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	tf := newTestFixture(t)

	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)
	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)
	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)

	tf.subscriber.Publish(struct{}{})
	assert.Equal(t, tf.count, 1)
}

func Test_ConcurrentInstallEvents_OnlyOnePublish(t *testing.T) {
	category.Set(t, category.Unit)
	resetGlobals()

	tf := newTestFixture(t)

	RegisterNotifier(tf.fileConfigManager, tf.subscriber, tf.notifier)
	var wg sync.WaitGroup
	const n = 20
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			tf.subscriber.Publish(struct{}{})
		}()
	}
	wg.Wait()

	assert.Equal(t, tf.count, 1, "expected exactly 1 publish")
}

type stubSubscriber struct {
	mu       sync.Mutex
	handlers []events.Handler[any]
}

func (s *stubSubscriber) Subscribe(h events.Handler[any]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers = append(s.handlers, h)
}

func (s *stubSubscriber) Publish(ins any) {
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
