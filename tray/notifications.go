package tray

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
	inotify "github.com/NordSecurity/nordvpn-linux/notify"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

func (ti *Instance) notify(text string, a ...any) {
	ti.sendNotification(notify.ExpireTimeoutSetByNotificationServer, text, a...)
}

func (ti *Instance) notifyWithShortDuration(text string, a ...any) {
	// on Ubuntu 22.04 the duration seams to be ignored by the OS
	ti.sendNotification(3*time.Second, text, a...)
}

func (ti *Instance) sendNotification(duration time.Duration, text string, a ...any) {
	text = fmt.Sprintf(text, a...)
	ti.state.mu.RLock()
	notificationsStatus := ti.state.notificationsStatus
	ti.state.mu.RUnlock()
	if notificationsStatus == Enabled {
		if err := ti.notifier.sendNotification("NordVPN", text, duration); err != nil {
			log.Println(internal.ErrorPrefix+" failed to send notification: ", err)
		}
	}
}

// notifyForce sends a notification, ignoring users notify setting
func (ti *Instance) notifyForce(text string, a ...any) {
	text = fmt.Sprintf(text, a...)
	if err := ti.notifier.sendNotification("NordVPN", text, notify.ExpireTimeoutSetByNotificationServer); err != nil {
		log.Println(internal.ErrorPrefix+" failed to send forced notification: ", err)
	}
}

// dbusNotifier wraps github.com/esiqveland/notify notifier implementation
type dbusNotifier struct {
	mu       sync.Mutex
	notifier notify.Notifier
}

func (n *dbusNotifier) start() {
	ntf, err := newNotifier()
	if err == nil {
		log.Println(internal.InfoPrefix + " Started dbus notifier")
		n.mu.Lock()
		n.notifier = ntf
		n.mu.Unlock()
	} else {
		log.Println(internal.ErrorPrefix+" Failed to start dbus notifier: ", err)
	}
}

// sendNotification sends notification via dbus. Thread safe.
func (n *dbusNotifier) sendNotification(summary string, body string, duration time.Duration) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.notifier != nil {
		notification := notify.Notification{
			AppName:       "NordVPN",
			Summary:       summary,
			AppIcon:       inotify.GetIconPath("nordvpn"),
			Body:          body,
			ExpireTimeout: duration,
			Hints: map[string]dbus.Variant{
				"transient": dbus.MakeVariant(1),
			},
		}
		if _, err := n.notifier.SendNotification(notification); err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("dbus notifier not connected")
	}
}

func newNotifier() (notify.Notifier, error) {
	dbusConn, err := dbus.SessionBusPrivate()

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err := dbusConn.Close(); err != nil {
				log.Println(internal.ErrorPrefix+" Failed to close dbus connection: ", err)
			}
		}
	}()

	if err = dbusConn.Auth(nil); err != nil {
		return nil, err
	}

	if err = dbusConn.Hello(); err != nil {
		return nil, err
	}

	ntf, err := notify.New(dbusConn)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			if err := ntf.Close(); err != nil {
				log.Println(internal.ErrorPrefix+" Failed to close notifier: ", err)
			}
		}
	}()

	return ntf, nil
}
