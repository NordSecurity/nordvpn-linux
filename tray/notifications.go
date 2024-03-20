package tray

import (
	"errors"
	"fmt"
	"sync"

	"github.com/esiqveland/notify"
	"github.com/fatih/color"
	"github.com/godbus/dbus/v5"
)

func notification(mode string, text string, a ...any) {
	text = fmt.Sprintf(text, a...)
	_, err := dbusNotifier.SendNotification("NordVPN", text)
	if err != nil {
		switch mode {
		case "info":
			color.Green(text)
		case "warning":
			color.Yellow(text)
		case "error":
			color.Red(text)
		}
	}
}

// DbusNotifier wraps github.com/esiqveland/notify notifier implementation
type DbusNotifier struct {
	mu       sync.Mutex
	notifier notify.Notifier
}

// SendNotification sends notification via dbus. Thread safe.
func (n *DbusNotifier) SendNotification(summary string, body string) (uint32, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.notifier != nil {
		notification := notify.Notification{
			AppName:       "NordVPN",
			Summary:       summary,
			AppIcon:       "nordvpn",
			Body:          body,
			ExpireTimeout: notify.ExpireTimeoutSetByNotificationServer,
			Hints: map[string]dbus.Variant{
				"transient": dbus.MakeVariant(1),
			},
		}
		return n.notifier.SendNotification(notification)
	} else {
		return 0, errors.New("DbusNotifier not connected")
	}
}

// Close dbus connection. Thread safe.
func (n *DbusNotifier) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.notifier != nil {
		return n.notifier.Close()
	} else {
		return nil
	}
}

func NewDbusNotifier() (*DbusNotifier, error) {
	dbusConn, err := dbus.SessionBusPrivate()
	defer func() {
		if err != nil {
			if err := dbusConn.Close(); err != nil {
				color.Red("failed to close dbus connection: ", err)
			}
		}
	}()

	if err != nil {
		return nil, err
	}

	if err = dbusConn.Auth(nil); err != nil {
		return nil, err
	}

	if err = dbusConn.Hello(); err != nil {
		return nil, err
	}

	notifier, err := notify.New(dbusConn)
	defer func() {
		if err != nil {
			if err := notifier.Close(); err != nil {
				color.Red("failed to close notifier: ", err)
			}
		}
	}()

	if err != nil {
		return nil, err
	}

	return &DbusNotifier{notifier: notifier}, nil
}
