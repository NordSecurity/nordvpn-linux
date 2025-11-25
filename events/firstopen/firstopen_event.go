package firstopen

import (
	"fmt"
	"sync/atomic"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
)

var (
	published atomic.Bool // true only when first open event was actually published
)

// firstOpenNotifier publishes a one-time `first_open` event when
// `cm.NewInstallation` is true and the FirstTimeOpened event fires.
// It uses an `atomic.Bool` to enforce exactly-once behavior.
type firstOpenNotifier struct {
	cm                     *config.FilesystemConfigManager
	emitFirstTimeOpenEvent func() error
}

// RegisterNotifier sets up a `firstOpenNotifier`.
// If `cm.NewInstallation` is true and the first open event is not yet published, it subscribes
// [firstopen.notifyOnceAppJustInstalled] to the FirstTimeOpened events stream.
func RegisterNotifier(
	cm *config.FilesystemConfigManager,
	firstTimeOpened events.PublishSubcriber[any],
	emitFirstTimeOpenEvent func() error,
) {
	if !cm.NewInstallation || published.Load() {
		return
	}
	n := firstOpenNotifier{
		cm:                     cm,
		emitFirstTimeOpenEvent: emitFirstTimeOpenEvent,
	}
	firstTimeOpened.Subscribe(n.notifyOnceAppJustInstalled)
}

// notifyOnceAppJustInstalled checks whether this is a fresh installation,
// and, if so, publishes a `first_open` event exactly once.
func (i *firstOpenNotifier) notifyOnceAppJustInstalled(any) error {
	if !i.cm.NewInstallation || !published.CompareAndSwap(false, true) {
		return nil
	}
	if err := i.emitFirstTimeOpenEvent(); err != nil {
		return fmt.Errorf("Failure upon emitting first open event: %w", err)
	}
	return nil
}
