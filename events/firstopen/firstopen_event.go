package firstopen

import (
	"sync"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
)

var (
	defaultGuard = &sync.Once{} // ensures `publishFirstOpen` runs at most once
	published    atomic.Bool    // true once we’ve actually published
)

// FirstOpenNotifier publishes a one-time `first_open` event when
// `cm.NewInstallation` is true and the device location event fires.
// It uses a `sync.Once` guard plus an `atomic.Bool` to enforce exactly-once
// behavior.
type FirstOpenNotifier struct {
	cm           *config.FilesystemConfigManager
	uiItemsClick events.Publisher[events.UiItemsAction]
	guard        *sync.Once
}

// RegisterNotifier sets up a `FirstOpenNotifier` using the shared `defaultGuard`.
// If `cm.NewInstallation` is true and we haven’t yet published, it subscribes
// [firstopen.NotifyOnceAppJustInstalled] to the device location events stream.
func RegisterNotifier(
	cm *config.FilesystemConfigManager,
	deviceLocation events.Subscriber[core.Insights],
	uiItemsClick events.Publisher[events.UiItemsAction],
) {
	registerNotifier(cm, deviceLocation, uiItemsClick, defaultGuard)
}

// registerNotifier is the internal constructor that lets you inject a custom guard.
func registerNotifier(
	cm *config.FilesystemConfigManager,
	deviceLocation events.Subscriber[core.Insights],
	uiItemsClick events.Publisher[events.UiItemsAction],
	guard *sync.Once,
) {
	notifier := FirstOpenNotifier{
		cm:           cm,
		uiItemsClick: uiItemsClick,
		guard:        guard,
	}
	if cm.NewInstallation && !published.Load() {
		deviceLocation.Subscribe(notifier.notifyOnceAppJustInstalled)
	}
}

// NotifyOnceAppJustInstalled checks whether this is a fresh installation,
// and, if so, publishes a `first_open` event exactly once.
func (i *FirstOpenNotifier) notifyOnceAppJustInstalled(_ core.Insights) error {
	if !i.cm.NewInstallation {
		return nil
	}

	i.guard.Do(i.publishFirstOpen)
	return nil
}

// publishFirstOpen emits the `UiItemsAction` event recording
// that the app was opened for the first time after installation and
// marks published=true.
// This method is protected by `sync.Once` to guarantee a single emission.
func (i *FirstOpenNotifier) publishFirstOpen() {
	i.uiItemsClick.Publish(events.UiItemsAction{
		ItemName:      "first_open",
		ItemType:      "button",
		ItemValue:     "first_open",
		FormReference: "daemon",
	})
	published.Store(true)
}
