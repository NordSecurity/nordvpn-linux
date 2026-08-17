// Package alert delivers API for sending OS notification.
package alert

import (
	"fmt"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

type AlertID = uint32

type Urgency int

const (
	// UrgencyNormal indicates that the alert should respect the user's settings.
	UrgencyNormal Urgency = iota
	// UrgencyCritical indicates that the alert should be shown regardless of the user's settings.
	UrgencyCritical
)

func (u Urgency) String() string {
	switch u {
	case UrgencyNormal:
		return "normal"
	case UrgencyCritical:
		return "critical"
	default:
		return "unknown"
	}
}

type Notifier interface {
	Alert(body string) *AlertBuilder
	Mute()
	Unmute()
	Close() error
}

type Alert struct {
	Summary string
	Body    string
	Actions []Action
	Urgency Urgency
}

func (a Alert) String() string {
	return fmt.Sprintf("[%s] %q (%s)", a.Summary, a.Body, a.Urgency)
}

type Action struct {
	Key      string
	Label    string
	Callback func()
}

type DbusNotifier struct {
	mu        sync.Mutex
	isActive  bool
	transient bool
	notifier  notify.Notifier
	actions   map[AlertID]map[string]func()
}

type Option func(*DbusNotifier)

func WithTransient() Option {
	return func(n *DbusNotifier) {
		n.transient = true
	}
}

func NewDbusNotifier(opts ...Option) (*DbusNotifier, error) {
	dbusConn, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, fmt.Errorf("creating D-Bus connection: %w", err)
	}

	var success bool
	defer func() {
		if !success {
			_ = dbusConn.Close()
		}
	}()

	if err = dbusConn.Auth(nil); err != nil {
		return nil, fmt.Errorf("authenticating to D-Bus: %w", err)
	}

	if err = dbusConn.Hello(); err != nil {
		return nil, fmt.Errorf("sending D-Bus hello: %w", err)
	}

	n := &DbusNotifier{
		actions:  make(map[AlertID]map[string]func()),
		isActive: true,
	}
	for _, opt := range opts {
		opt(n)
	}

	notifier, err := notify.New(
		dbusConn,
		notify.WithOnAction(n.dispatchAction),
		notify.WithOnClosed(n.forget),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new notifier: %w", err)
	}

	n.notifier = notifier
	success = true

	return n, nil
}

func (n *DbusNotifier) Alert(body string) *AlertBuilder {
	return NewAlertBuilder(n.doNotify, body)
}

func (n *DbusNotifier) doNotify(alert Alert) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.isActive && alert.Urgency == UrgencyNormal {
		log.Infof("notification suppressed: %s", alert)
		return
	}

	alertActions := make([]notify.Action, 0, len(alert.Actions))
	for _, action := range alert.Actions {
		alertActions = append(alertActions, notify.Action{Key: action.Key, Label: action.Label})
	}

	notif := notify.Notification{
		AppName:       internal.AppName,
		Summary:       alert.Summary,
		AppIcon:       GetIconPath("nordvpn"),
		Body:          alert.Body,
		ExpireTimeout: notify.ExpireTimeoutSetByNotificationServer,
		Actions:       alertActions,
	}

	if n.transient {
		notif.Hints = map[string]dbus.Variant{"transient": dbus.MakeVariant(1)}
	}

	log.Infof("sending notification: %s", alert)
	id, err := n.notifier.SendNotification(notif)
	if err != nil {
		log.Errorf("failed to send notification '%s': %v", notif, err)
		return
	}

	if len(alert.Actions) > 0 {
		callbacks := make(map[string]func(), len(alert.Actions))
		for _, action := range alert.Actions {
			if action.Callback != nil {
				callbacks[action.Key] = action.Callback
			}
		}
		n.actions[id] = callbacks
	}
}

func (n *DbusNotifier) dispatchAction(action *notify.ActionInvokedSignal) {
	n.mu.Lock()
	callbacks, ok := n.actions[action.ID]
	delete(n.actions, action.ID)
	n.mu.Unlock()

	if !ok {
		return
	}

	callback, ok := callbacks[action.ActionKey]
	if !ok || callback == nil {
		log.Error("Unknown action key: ", action.ActionKey)
		return
	}

	callback()
}

func (n *DbusNotifier) forget(ncs *notify.NotificationClosedSignal) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.actions, ncs.ID)
}

func (n *DbusNotifier) Mute() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.isActive = false
}

func (n *DbusNotifier) Unmute() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.isActive = true
}

func (n *DbusNotifier) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.isActive = false

	return n.notifier.Close()
}
