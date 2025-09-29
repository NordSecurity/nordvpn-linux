package tray

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"

	"github.com/NordSecurity/nordvpn-linux/internal"
	inotify "github.com/NordSecurity/nordvpn-linux/notify"
)

var dbusNotifierNotConnectedError = errors.New("dbus notifier not connected")

func (ti *Instance) notify(ntype NotificationType, text string, a ...any) {
	ti.state.mu.RLock()
	notificationsStatus := ti.state.notificationsStatus
	ti.state.mu.RUnlock()

	if notificationsStatus == Enabled || ntype == Force {
		text = fmt.Sprintf(text, a...)
		log.Printf("%s Sending notification: %s\n", logTag, text)
		if err := ti.notifier.sendNotification("NordVPN", text); err != nil {
			if !errors.Is(err, dbusNotifierNotConnectedError) {
				log.Println(internal.ErrorPrefix, "Failed to send notification:", err)
			}
		}
	} else {
		log.Printf("%s Notification suppressed: %s (status: %d, force: %v)\n", logTag, text, notificationsStatus, ntype == Force)
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
		log.Println(internal.InfoPrefix, "Started dbus notifier")
		n.mu.Lock()
		n.notifier = ntf
		n.mu.Unlock()
	} else {
		log.Println(internal.ErrorPrefix, "Failed to start dbus notifier:", err)
	}
}

// sendNotification sends notification via dbus. Thread safe.
func (n *dbusNotifier) sendNotification(summary string, body string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.notifier != nil {
		notification := notify.Notification{
			AppName:       "NordVPN",
			Summary:       summary,
			AppIcon:       inotify.GetIconPath("nordvpn"),
			Body:          body,
			ExpireTimeout: notify.ExpireTimeoutSetByNotificationServer,
			Hints: map[string]dbus.Variant{
				"transient": dbus.MakeVariant(1),
			},
		}
		if _, err := n.notifier.SendNotification(notification); err != nil {
			return err
		}
		return nil
	} else {
		return dbusNotifierNotConnectedError
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
				log.Println(internal.ErrorPrefix, "Failed to close dbus connection:", err)
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
				log.Println(internal.ErrorPrefix, "Failed to close notifier:", err)
			}
		}
	}()

	return ntf, nil
}

type NotificationType bool

const (
	// NoForce indicates that the notification should respect the user's settings.
	NoForce NotificationType = false
	// Force indicates that the notification should be shown regardless of the user's settings.
	Force NotificationType = true
)

