package firstopen

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// FirstOpenNotifier publishes a one-time “first_open” UI event
// when the application is detected to have been installed.
// It uses sync.Once to ensure the event is only published once
// per process lifecycle.
type FirstOpenNotifier struct {
	once         sync.Once
	cm           *config.FilesystemConfigManager
	uiItemsClick events.PublishSubcriber[events.UiItemsAction]
}

func NewNotifier(
	cm *config.FilesystemConfigManager,
	uiItemsClick events.PublishSubcriber[events.UiItemsAction],
) FirstOpenNotifier {
	return FirstOpenNotifier{
		cm:           cm,
		uiItemsClick: uiItemsClick,
	}
}

// NotifyOnceAppJustInstalled checks whether this is a fresh installation,
// and, if so, publishes a “first_open” UI event exactly once.
func (i *FirstOpenNotifier) NotifyOnceAppJustInstalled(_ core.Insights) error {
	if i.cm.NewInstallation {
		i.once.Do(i.publishFirstOpen)
	}
	return nil
}

// publishFirstOpen emits the UiItemsAction event recording
// that the app was opened for the first time after installation.
// This method is protected by sync.Once to guarantee a single emission.
func (i *FirstOpenNotifier) publishFirstOpen() {
	i.uiItemsClick.Publish(events.UiItemsAction{
		ItemName:      "first_open",
		ItemType:      "button",
		ItemValue:     "first_open",
		FormReference: "daemon",
	})
}
