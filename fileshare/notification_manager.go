package fileshare

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	transferCancelAction = "Cancel"

	notifyNewTransferSummary = "New transfer request"
	notifyNewTransferBody    = "Transfer ID: %s\nFrom: %s"

	acceptFailedNotificationSummary     = "Failed to accept transfer"
	acceptFileFailedNotificationSummary = "Failed to download file"
	downloadDirNotFoundError            = "Default download directory not found"
	downloadDirIsASymlinkError          = "Default download directory is a symlink"
	downloadDirIsNotADirError           = "Default download directory is a symlink"
	notEnoughSpaceOnDeviceError         = "Not enough space on the device"
	transferAleradyAccepted             = "Transfer has been already acceptd"
	acceptErrorGeneric                  = "Failed to accept transfer, try command line, and if the issue repeats, contact our customer support"

	cancelFailedNotificationSummary = "Failed to cancel transfer"
	transferNotCancelableError      = "Transfer is already canceled or accepted"
	cancelErrorGeneric              = "Failed to cancel transfer, try command line, and if the issue repeats, contact our customer support"
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

// NotificationManager is responsible for creating gui pop-up notifications for changes in transfer file status
type NotificationManager struct {
	// maps Open action id to file path for downloaded files
	downloadedFiles map[uint32]string
	// maps Accept action id to transfer id for incoming transfers
	transfers          map[uint32]string
	stateMutex         sync.Mutex
	notifier           Notifier
	filesystem         Filesystem
	eventManager       *EventManager
	fileshare          Fileshare
	openFileFunc       func(string)
	defaultDownloadDir string
}

// NewNotificationManager creates a new notification
func NewNotificationManager(fileshare Fileshare, eventManager *EventManager) (*NotificationManager, error) {
	defaultDownloadDir := ""
	home, err := os.UserHomeDir()
	if err == nil {
		defaultDownloadDir = filepath.Join(home, "Downloads")
		if _, err = os.Stat(defaultDownloadDir); err != nil {
			log.Println("Failed to determine default download dir: ", err)
			defaultDownloadDir = ""
		}
	} else {
		log.Println("Failed to determine default download dir: ", err)
	}

	notificationManager := NotificationManager{
		downloadedFiles:    make(map[uint32]string),
		transfers:          make(map[uint32]string),
		filesystem:         StdFilesystem{},
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
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	if err := nm.notifier.Close(); err != nil {
		log.Println("Failed to close notifier: ", err)
	}
}

// OpenFile associated with actionID
func (nm *NotificationManager) OpenFile(actionID uint32) {
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	if filename, ok := nm.downloadedFiles[actionID]; ok {
		nm.openFileFunc(filename)
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

// NotifyFile creates a pop-up gui notification, in case of incoming files, filename should be a full path
// (download path + filename), so that it can be opened by the user.
func (nm *NotificationManager) NotifyFile(filename string, direction pb.Direction, status pb.Status) {
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	summary := fileStatusToNotificationSummary(direction, status)

	if direction == pb.Direction_INCOMING && status == pb.Status_SUCCESS {
		if notificationID, err :=
			nm.notifier.SendNotification(summary, filename, []Action{{actionKeyOpenFile, "Open"}}); err == nil {
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

func (nm *NotificationManager) sendGenericNotification(summary string, body string) {
	_, err := nm.notifier.SendNotification(summary, body, nil)
	if err != nil {
		log.Println("failed to send generic notification: ", err)
	}
}

// AcceptTransfer associated with actionID, generates notifications on failure
func (nm *NotificationManager) AcceptTransfer(actionID uint32) {
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	transferID, ok := nm.transfers[actionID]
	if !ok {
		log.Println("Failed to accept transfer from notification manager, actionID not found")
		return
	}

	delete(nm.transfers, actionID)

	destinationFileInfo, err := nm.filesystem.Lstat(nm.defaultDownloadDir)

	if err != nil {
		log.Println("Failed to lstat a file: ", err)
		nm.sendGenericNotification(acceptFailedNotificationSummary,
			downloadDirNotFoundError)
		return
	}

	if destinationFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		log.Println("Default accept destination is a symlink")
		nm.sendGenericNotification(acceptFailedNotificationSummary,
			downloadDirIsASymlinkError)
		return
	}

	if !destinationFileInfo.IsDir() {
		log.Println("Default accept destination is not a directory")
		nm.sendGenericNotification(acceptFailedNotificationSummary,
			downloadDirIsNotADirError)
		return
	}

	statfs, err := nm.filesystem.Statfs(nm.defaultDownloadDir)
	if err != nil {
		log.Println("doing statfs: ", err)
		nm.sendGenericNotification(acceptFailedNotificationSummary,
			notEnoughSpaceOnDeviceError)
		return
	}

	transfer, err := nm.eventManager.AcceptTransfer(transferID,
		nm.defaultDownloadDir,
		[]string{},
		statfs.Bavail*uint64(statfs.Bsize))

	switch err {
	case ErrTransferAlreadyAccepted:
		nm.sendGenericNotification(acceptFailedNotificationSummary, transferAleradyAccepted)
		return
	case ErrSizeLimitExceeded:
		nm.sendGenericNotification(acceptFailedNotificationSummary, notEnoughSpaceOnDeviceError)
		return
	case nil:
		break
	default:
		log.Println("Unexpected error when accepting transfer from notification manager: ", err)
		nm.sendGenericNotification(acceptFailedNotificationSummary, acceptErrorGeneric)
		return
	}

	for _, file := range GetAllTransferFiles(transfer) {
		if err = nm.fileshare.Accept(transferID, nm.defaultDownloadDir, file.Id); err != nil {
			nm.sendGenericNotification(acceptFileFailedNotificationSummary, file.Id)
		}
	}

	if err != nil {
		log.Println("Failed to accept some files: ", err)
	}
}

// CancelTransfer associated with actionID, generates error notifiacation on failure
func (nm *NotificationManager) CancelTransfer(actionID uint32) {
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	transferID, ok := nm.transfers[actionID]
	if !ok {
		log.Println("Failed to cancel transfer from notification manager, actionID not found")
		return
	}

	delete(nm.transfers, actionID)

	transfer, err := nm.eventManager.GetTransfer(transferID)

	if err != nil {
		log.Println("Failed to cancel transfer from notification manager: ", err)
		nm.sendGenericNotification(cancelFailedNotificationSummary, cancelErrorGeneric)
		return
	}

	if transfer.Status != pb.Status_ONGOING && transfer.Status != pb.Status_REQUESTED {
		nm.sendGenericNotification(cancelFailedNotificationSummary, transferNotCancelableError)
		return
	}

	if err := nm.fileshare.Cancel(transferID); err != nil {
		log.Println("Failed to cancel transfer from notification manager: ", err)
		nm.sendGenericNotification(cancelFailedNotificationSummary, err.Error())
	}
}

// NotifyTransfer creates a pop-up gui notification
func (nm *NotificationManager) NotifyNewTransfer(transferID string, peer string) {
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	body := fmt.Sprintf(notifyNewTransferBody, transferID, peer)

	actionID, err := nm.notifier.SendNotification(
		notifyNewTransferSummary,
		body,
		[]Action{
			{actionKeyAcceptTransfer, transferAcceptAction},
			{actionKeyCancelTransfer, transferCancelAction}})

	if err != nil {
		log.Println("failed to send notification for new transfer: ", err)
	}

	nm.transfers[actionID] = transferID
}

// CloseNotification cleans up any data associated with actionID
func (nm *NotificationManager) CloseNotification(actoionID uint32) {
	nm.stateMutex.Lock()
	defer nm.stateMutex.Unlock()

	delete(nm.downloadedFiles, actoionID)
	delete(nm.transfers, actoionID)
}
