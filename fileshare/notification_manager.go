package fileshare

import (
	"log"
	"os/exec"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

// Action represents an action available to the user when notification is displayed
type Action struct {
	Key    string
	Action string
}

// Notifier is responsible for sending notifications to the user
type Notifier interface {
	SendNotification(summary string, body string, actions []Action) (uint32, error)
	Close() error
}

// DbusNotifier wraps github.com/esiqveland/notify notifier implementation
type DbusNotifier struct {
	notifier notify.Notifier
}

// SendNotification sends notification via dbus
func (n DbusNotifier) SendNotification(summary string, body string, actions []Action) (uint32, error) {
	notifyActions := []notify.Action{}

	for _, action := range actions {
		notifyActions = append(notifyActions, notify.Action{Key: action.Key, Label: action.Action})
	}

	notification := notify.Notification{
		AppName:       "NordVPN",
		Summary:       summary,
		AppIcon:       "nordvpn",
		Body:          body,
		ExpireTimeout: notify.ExpireTimeoutSetByNotificationServer,
		Actions:       notifyActions,
	}

	return n.notifier.SendNotification(notification)
}

func (n DbusNotifier) Close() error {
	return n.notifier.Close()
}

func newDbusNotifier(notificationManager *NotificationManager) (*DbusNotifier, error) {
	dbusConn, err := dbus.SessionBusPrivate()
	defer func() {
		if err != nil {
			if err := dbusConn.Close(); err != nil {
				log.Println("failed to close dbus connection: ", err)
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

	onAction := func(action *notify.ActionInvokedSignal) {
		log.Println(action)
		if action.ActionKey == "open" {
			notificationManager.openFile(action.ID)
		}
	}

	notifier, err := notify.New(
		dbusConn,
		notify.WithOnAction(onAction),
		notify.WithOnClosed(func(ncs *notify.NotificationClosedSignal) {}),
	)
	defer func() {
		if err != nil {
			if err := notifier.Close(); err != nil {
				log.Println("failed to close notifier: ", err)
			}
		}
	}()

	if err != nil {
		return nil, err
	}

	return &DbusNotifier{notifier: notifier}, nil
}

// FileOpener opens file using the default aplication for that file type
type FileOpener interface {
	OpenFile(path string)
}

// XdgFileOpener opens files in the default application with xdg-open command
type XdgFileOpener struct {
}

// OpenFile opens a file with xdg-open command
func (XdgFileOpener) OpenFile(path string) {
	if err := exec.Command("xdg-open", path).Start(); err != nil {
		log.Println("failed to open file from notification ", err)
	}
}

// NotificationManager is responsible for creating gui pop-up notifications for changes in transfer file status
type NotificationManager struct {
	// maps Open action id to file path for downloaded files
	downloadedFiles      map[uint32]string
	downloadedFilesMutex sync.Mutex
	notifier             Notifier
	fileOpener           FileOpener
}

// NewNotificationManager creates a new notification
func NewNotificationManager() (*NotificationManager, error) {
	notificationManager := NotificationManager{
		downloadedFiles: make(map[uint32]string),
		fileOpener:      XdgFileOpener{},
	}

	notifier, err := newDbusNotifier(&notificationManager)
	if err != nil {
		return nil, err
	}

	notificationManager.notifier = notifier

	return &notificationManager, nil
}

func (nm *NotificationManager) Disable() {
	if err := nm.notifier.Close(); err != nil {
		log.Println("Failed to close notifier: ", err)
	}
}

func (nm *NotificationManager) openFile(actionID uint32) {
	nm.downloadedFilesMutex.Lock()
	defer nm.downloadedFilesMutex.Unlock()

	if filename, ok := nm.downloadedFiles[actionID]; ok {
		nm.fileOpener.OpenFile(filename)
		delete(nm.downloadedFiles, actionID)
	}
}

func fileStatusToNotificationSummary(direction pb.Direction, status pb.Status) string {
	//exhaustive:ignore
	switch direction {
	case pb.Direction_INCOMING:
		if summary, ok := IncomingFileStatus[status]; ok {
			return summary
		}
	case pb.Direction_OUTGOING:
		if summary, ok := OutgoingFileStatus[status]; ok {
			return summary
		}
	}

	summary, ok := FileStatus[status]
	if !ok {
		log.Printf("failed to convert file status %s for direction %s to text summary",
			status.String(), direction.String())
	}

	return summary
}

// Notify creates a pop-up gui notification, in case of incoming files, filename should be a full path
// (download path + filename), so that it can be opened by the user.
func (nm *NotificationManager) Notify(filename string, status pb.Status, direction pb.Direction) {
	nm.downloadedFilesMutex.Lock()
	defer nm.downloadedFilesMutex.Unlock()

	summary := fileStatusToNotificationSummary(direction, status)

	if direction == pb.Direction_INCOMING && status == pb.Status_SUCCESS {
		if notificationID, err :=
			nm.notifier.SendNotification(summary, filename, []Action{{"open", "Open"}}); err == nil {
			nm.downloadedFiles[notificationID] = filename
		} else {
			log.Printf("failed to send notification for file %s: %s", filename, err)
		}
		return
	}

	_, err := nm.notifier.SendNotification(summary, filename, nil)
	if err != nil {
		log.Printf("failed to send notification for file %s: %s", filename, err)
	}
}
