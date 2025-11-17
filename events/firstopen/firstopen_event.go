package firstopen

import (
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
)

var (
	published atomic.Bool // true only when first open event was actually published
)

// firstOpenNotifier publishes a one-time `first_open` event when
// `cm.NewInstallation` is true and the device location event fires.
// It uses an `atomic.Bool` to enforce exactly-once behavior.
type firstOpenNotifier struct {
	cm                     *config.FilesystemConfigManager
	emitFirstTimeOpenEvent func() error
}

// RegisterNotifier sets up a `firstOpenNotifier`.
// If `cm.NewInstallation` is true and the first open event is not yet published, it subscribes
// [firstopen.notifyOnceAppJustInstalled] to the device location events stream.
func RegisterNotifier(
	cm *config.FilesystemConfigManager,
	deviceLocation events.Subscriber[core.Insights],
	emitFirstTimeOpenEvent func() error,
) {
	if !cm.NewInstallation || published.Load() {
		return
	}
	n := firstOpenNotifier{
		cm:                     cm,
		emitFirstTimeOpenEvent: emitFirstTimeOpenEvent,
	}
	deviceLocation.Subscribe(n.notifyOnceAppJustInstalled)
}

// notifyOnceAppJustInstalled checks whether this is a fresh installation,
// and, if so, publishes a `first_open` event exactly once.
func (i *firstOpenNotifier) notifyOnceAppJustInstalled(_ core.Insights) error {
	if !i.cm.NewInstallation || !published.CompareAndSwap(false, true) {
		return nil
	}
	return i.emitFirstTimeOpenEvent()
}
