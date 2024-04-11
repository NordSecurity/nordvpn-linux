package tray

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

type logPriority int

const (
	pInfo logPriority = iota
	pWarning
	pError
)

func (ti *Instance) notify(text string, a ...any) {
	text = fmt.Sprintf(text, a...)
	ti.state.mu.RLock()
	notifyEnabled := ti.state.notifyEnabled
	ti.state.mu.RUnlock()
	err := errors.New("notifications disabled")
	if notifyEnabled {
		_, err = ti.notifier.sendNotification("NordVPN", text)
	}
	if err != nil {
		log.Println("failed to send notification: ", err)
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
		log.Println("Started dbus notifier")
		n.mu.Lock()
		n.notifier = ntf
		n.mu.Unlock()
	} else {
		log.Println("Failed to start dbus notifier: ", err)
	}
}

// sendNotification sends notification via dbus. Thread safe.
func (n *dbusNotifier) sendNotification(summary string, body string) (uint32, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.notifier != nil {
		notification := notify.Notification{
			AppName:       "NordVPN",
			Summary:       summary,
			AppIcon:       getIconPath("nordvpn"),
			Body:          body,
			ExpireTimeout: notify.ExpireTimeoutSetByNotificationServer,
			Hints: map[string]dbus.Variant{
				"transient": dbus.MakeVariant(1),
			},
		}
		return n.notifier.SendNotification(notification)
	} else {
		return 0, errors.New("dbus notifier not connected")
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
				log.Println("Failed to close dbus connection: ", err)
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
				log.Println("Failed to close notifier: ", err)
			}
		}
	}()

	return ntf, nil
}
