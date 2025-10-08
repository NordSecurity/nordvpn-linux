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

func (ti *Instance) notify(text string, a ...any) {
	text = fmt.Sprintf(text, a...)
	ti.state.mu.RLock()
	notificationsStatus := ti.state.notificationsStatus
	ti.state.mu.RUnlock()
	if notificationsStatus == Enabled {
		if err := ti.notifier.sendNotification("NordVPN", text); err != nil {
			if !errors.Is(err, dbusNotifierNotConnectedError) {
				log.Println(internal.ErrorPrefix, "Failed to send notification:", err)
			}
		}
	}
}

// notifyForce sends a notification, ignoring users notify setting
func (ti *Instance) notifyForce(text string, a ...any) {
	text = fmt.Sprintf(text, a...)
	if err := ti.notifier.sendNotification("NordVPN", text); err != nil {
		if !errors.Is(err, dbusNotifierNotConnectedError) {
			log.Println(internal.ErrorPrefix, "Failed to send forced notification:", err)
		}
	}
}

// dbusNotifier wraps github.com/esiqveland/notify notifier implementation
type dbusNotifier struct {
	mu       sync.Mutex
	notifier notify.Notifier
}

func (n *dbusNotifier) start() {
	log.Println(internal.InfoPrefix, "Starting dbus notifier")
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
	log.Println(internal.ErrorPrefix, "######    creating new notifier")
	dbusConn, err := dbus.SessionBusPrivate()
	if err != nil {
		log.Println(internal.ErrorPrefix, "######    error with getting dbus", err)
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
		log.Println(internal.ErrorPrefix, "######    Error in auth", err)
		return nil, err
	}

	if err = dbusConn.Hello(); err != nil {
		log.Println(internal.ErrorPrefix, "######    Error in hello", err)
		return nil, err
	}

	ntf, err := notify.New(dbusConn)
	if err != nil {
		log.Println(internal.ErrorPrefix, "######    in newnotify", err)
		return nil, err
	}

	defer func() {
		if err != nil {
			log.Println(internal.ErrorPrefix, "######    some error happened so closing", err)
			if err := ntf.Close(); err != nil {
				log.Println(internal.ErrorPrefix, "Failed to close notifier:", err)
			}
		}
		log.Println(internal.ErrorPrefix, "######    creating new notifier finished with success in defer")
	}()

	log.Println(internal.ErrorPrefix, "######    creating new notifier finished")
	return ntf, nil
}
