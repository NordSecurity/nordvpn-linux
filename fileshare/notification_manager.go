package fileshare

import (
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

const (
	actionKeyOpenFile       = "open-file"
	actionKeyAcceptTransfer = "accept-transfer"
	actionKeyCancelTransfer = "cancel-transfer"

	transferAcceptAction = "Accept"
	transferCancelAction = "Decline"

	notifyNewTransferSummary    = "New file transfer!"
	notifyNewTransferBody       = "Transfer ID: %s\nFrom: %s"
	notifyNewAutoacceptTransfer = "New transfer accepted automatically"
	notifyAutoacceptFailed      = "Failed to autoaccept transfer"

	acceptFailedNotificationSummary     = "Failed to accept transfer"
	acceptFileFailedNotificationSummary = "Failed to download file"
	downloadDirNotFoundError            = "The download directory doesn't exist."
	downloadDirIsASymlinkError          = "The download path can’t be a symbolic link."
	downloadDirIsNotADirError           = "The download path must be a directory."
	downloadDirNoPermissions            = "You don’t have write permissions for the download directory."
	notEnoughSpaceOnDeviceError         = "There’s not enough storage on your device."

	cancelFailedNotificationSummary = "Failed to decline transfer"

	transferInvalidated = "You’ve already accepted or declined this transfer."
	genericError        = "Something went wrong."
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
	mu       sync.Mutex
	notifier notify.Notifier
}

// SendNotification sends notification via dbus. Thread safe.
func (n *DbusNotifier) SendNotification(summary string, body string, actions []Action) (uint32, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

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

// Close dbus connection. Thread safe.
func (n *DbusNotifier) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

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
		switch action.ActionKey {
		case actionKeyOpenFile:
			notificationManager.OpenFile(action.ID)
		case actionKeyAcceptTransfer:
			notificationManager.AcceptTransfer(action.ID)
		case actionKeyCancelTransfer:
			notificationManager.CancelTransfer(action.ID)
		default:
			log.Println("Unknown action key: ", action.ActionKey)
		}
	}

	notifier, err := notify.New(
		dbusConn,
		notify.WithOnAction(onAction),
		notify.WithOnClosed(func(ncs *notify.NotificationClosedSignal) {
			notificationManager.CloseNotification(ncs.ID)
		}),
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

// openFileXdg opens a file with xdg-open command
func openFileXdg(path string) {
	if err := exec.Command("xdg-open", path).Start(); err != nil {
		log.Println("failed to open file from notification ", err)
	}
}

type notificationsStorage struct {
	// maps Open action id to file path for downloaded files
	downloadedFiles map[uint32]string
	// maps Accept action id to transfer id for incoming transfers
	transfers map[uint32]string
	mu        sync.Mutex
}

func newNotificationStorage() notificationsStorage {
	return notificationsStorage{
		downloadedFiles: make(map[uint32]string),
		transfers:       make(map[uint32]string),
	}
}

// AddTransferNotification, thread safe
func (ns *notificationsStorage) AddTransferNotification(notificationID uint32, transferID string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.transfers[notificationID] = transferID
}

// GetAndDeleteTransferNotification, returns transfer id associated with give notification id and
// removes it from the storage. Second return value denotes if given notification id was found in
// the storage. Thread safe.
func (ns *notificationsStorage) GetAndDeleteTransferNotification(notificationID uint32) (string, bool) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	transferID, ok := ns.transfers[notificationID]

	delete(ns.transfers, notificationID)

	return transferID, ok
}

// AddFileNotification, thread safe
func (ns *notificationsStorage) AddFileNotification(notificationID uint32, file string) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	ns.downloadedFiles[notificationID] = file
}

// GetAndDeleteFileNotification, returns filename associated with given notification id and removes it
// from the storage. Second return value denotes if given notification id was found in the storage.
// Thread safe.
func (ns *notificationsStorage) GetAndDeleteFileNotification(notificationID uint32) (string, bool) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	file, ok := ns.downloadedFiles[notificationID]

	delete(ns.downloadedFiles, notificationID)

	return file, ok
}

// NotificationManager is responsible for creating gui pop-up notifications for changes in transfer file status
type NotificationManager struct {
	notifications      notificationsStorage
	notifier           Notifier
	eventManager       *EventManager
	fileshare          Fileshare
	openFileFunc       func(string)
	defaultDownloadDir string
}

// NewNotificationManager creates a new notification
func NewNotificationManager(fileshare Fileshare, eventManager *EventManager) (*NotificationManager, error) {
	defaultDownloadDir, err := GetDefaultDownloadDirectory()

	if err != nil {
		log.Println("Failed to find default download directory: ", err.Error())
	}

	notificationManager := NotificationManager{
		notifications:      newNotificationStorage(),
		fileshare:          fileshare,
		openFileFunc:       openFileXdg,
		defaultDownloadDir: defaultDownloadDir,
		eventManager:       eventManager,
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

// OpenFile associated with notificationID
func (nm *NotificationManager) OpenFile(notificationID uint32) {
	if filename, ok := nm.notifications.GetAndDeleteFileNotification(notificationID); ok {
		nm.openFileFunc(filename)
	}
}

func destinationDirectoryErrorToNotificationBody(err error) string {
	switch err {
	case ErrSizeLimitExceeded:
		return notEnoughSpaceOnDeviceError
	case ErrTransferAlreadyAccepted:
		return transferInvalidated
	case ErrAcceptDirNotFound:
		return downloadDirNotFoundError
	case ErrAcceptDirIsASymlink:
		return downloadDirIsASymlinkError
	case ErrAcceptDirIsNotADirectory:
		return downloadDirIsNotADirError
	case ErrNoPermissionsToAcceptDirectory:
		return downloadDirNoPermissions
	default:
		log.Println("Unknown error: ", err.Error())
		return genericError
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

// NotifyFile creates a pop-up gui notification, in case of incoming files, filename should be a full path
// (download path + filename), so that it can be opened by the user.
func (nm *NotificationManager) NotifyFile(filename string, direction pb.Direction, status pb.Status) {
	summary := fileStatusToNotificationSummary(direction, status)

	if direction == pb.Direction_INCOMING && status == pb.Status_SUCCESS {
		if notificationID, err :=
			nm.notifier.SendNotification(summary, filename, []Action{{actionKeyOpenFile, "Open"}}); err == nil {
			nm.notifications.AddFileNotification(notificationID, filename)
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

func (nm *NotificationManager) sendGenericNotification(summary string, body string) {
	_, err := nm.notifier.SendNotification(summary, body, nil)
	if err != nil {
		log.Println("failed to send generic notification: ", err)
	}
}

// AcceptTransfer associated with notificationID, generates notifications on failure
func (nm *NotificationManager) AcceptTransfer(notificationID uint32) {
	transferID, ok := nm.notifications.GetAndDeleteTransferNotification(notificationID)

	if !ok {
		return
	}

	transfer, err := nm.eventManager.AcceptTransfer(transferID,
		nm.defaultDownloadDir,
		[]string{})

	if err != nil {
		notificationSummary := destinationDirectoryErrorToNotificationBody(err)
		nm.sendGenericNotification(acceptFailedNotificationSummary, notificationSummary)
		return
	}

	for _, file := range transfer.Files {
		if err = nm.fileshare.Accept(transferID, nm.defaultDownloadDir, file.Id); err != nil {
			nm.sendGenericNotification(acceptFileFailedNotificationSummary, file.Id)
		}
	}

	if err != nil {
		log.Println("Failed to accept some files: ", err)
	}
}

// CancelTransfer associated with notificationID, generates error notifiacation on failure
func (nm *NotificationManager) CancelTransfer(notificationID uint32) {
	transferID, ok := nm.notifications.GetAndDeleteTransferNotification(notificationID)

	if !ok {
		return
	}

	transfer, err := nm.eventManager.GetTransfer(transferID)

	if err != nil {
		log.Println("Failed to cancel transfer from notification manager: ", err)
		nm.sendGenericNotification(cancelFailedNotificationSummary, genericError)
		return
	}

	if transfer.Status != pb.Status_ONGOING && transfer.Status != pb.Status_REQUESTED {
		nm.sendGenericNotification(cancelFailedNotificationSummary, transferInvalidated)
		return
	}

	if err := nm.fileshare.Cancel(transferID); err != nil {
		log.Println("Failed to cancel transfer from notification manager: ", err)
		nm.sendGenericNotification(cancelFailedNotificationSummary, err.Error())
	}
}

// NotifyTransfer creates a pop-up gui notification
func (nm *NotificationManager) NotifyNewTransfer(transferID string, peer string) {
	body := fmt.Sprintf(notifyNewTransferBody, transferID, peer)

	notificationID, err := nm.notifier.SendNotification(
		notifyNewTransferSummary,
		body,
		[]Action{
			{actionKeyAcceptTransfer, transferAcceptAction},
			{actionKeyCancelTransfer, transferCancelAction}})

	if err != nil {
		log.Println("failed to send notification for new transfer: ", err)
	}

	nm.notifications.AddTransferNotification(notificationID, transferID)
}

// NotifyNewAutoacceptTransfer creates a pop-up gui notification
func (nm *NotificationManager) NotifyNewAutoacceptTransfer(transferID string, peer string) {
	body := fmt.Sprintf(notifyNewTransferBody, transferID, peer)

	nm.sendGenericNotification(notifyNewAutoacceptTransfer, body)
}

// NotifyAutoacceptFailed creates a pop-up gui notification
func (nm *NotificationManager) NotifyAutoacceptFailed(transferID string, peer string, reason error) {
	transferInfo := fmt.Sprintf(notifyNewTransferBody, transferID, peer)
	body := fmt.Sprintf("%s\n%s", destinationDirectoryErrorToNotificationBody(reason), transferInfo)

	nm.sendGenericNotification(notifyAutoacceptFailed, body)
}

// CloseNotification cleans up any data associated with notificationID
func (nm *NotificationManager) CloseNotification(notificationID uint32) {
	nm.notifications.GetAndDeleteFileNotification(notificationID)
	nm.notifications.GetAndDeleteTransferNotification(notificationID)
}
