package fileshare

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/NordSecurity/nordvpn-linux/alert"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/log"
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

	transferCanceledByPeerNotificationSummary = "Transfer no longer exists"
	transferCanceledByPeerNotificationBody    = "The sender has canceled this transfer."

	transferInvalidated = "You’ve already accepted or declined this transfer."
	genericError        = "Something went wrong."
)

// openFileXdg opens a file with xdg-open command
func openFileXdg(path string) {
	if err := exec.Command("xdg-open", path).Start(); err != nil {
		log.Error("failed to open file from notification ", err)
	}
}

// NotificationManager is responsible for creating gui pop-up notifications for changes in transfer file status
type NotificationManager struct {
	n                  alert.Notifier
	eventManager       *EventManager
	fileshare          Fileshare
	openFileFunc       func(string)
	defaultDownloadDir string
}

// NewNotificationManager creates a new notification
func NewNotificationManager(fileshare Fileshare, eventManager *EventManager) (*NotificationManager, error) {
	defaultDownloadDir, err := GetDefaultDownloadDirectory()
	if err != nil {
		log.Error("Failed to find default download directory: ", err.Error())
	}

	notifier, err := alert.NewDbusNotifier()
	if err != nil {
		return nil, err
	}

	return &NotificationManager{
		n:                  notifier,
		fileshare:          fileshare,
		openFileFunc:       openFileXdg,
		defaultDownloadDir: defaultDownloadDir,
		eventManager:       eventManager,
	}, nil
}

func acceptErrorToNotificationBody(err error) string {
	switch {
	case errors.Is(err, ErrSizeLimitExceeded):
		return notEnoughSpaceOnDeviceError
	case errors.Is(err, ErrTransferAlreadyAccepted):
		return transferInvalidated
	case errors.Is(err, ErrAcceptDirNotFound):
		return downloadDirNotFoundError
	case errors.Is(err, ErrAcceptDirIsASymlink):
		return downloadDirIsASymlinkError
	case errors.Is(err, ErrAcceptDirIsNotADirectory):
		return downloadDirIsNotADirError
	case errors.Is(err, ErrNoPermissionsToAcceptDirectory):
		return downloadDirNoPermissions
	case errors.Is(err, ErrTransferCanceledByPeer):
		return transferCanceledByPeerNotificationBody
	default:
		log.Error("Unknown error: ", err.Error())
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
		log.Warnf("failed to convert file status %s for direction %s to text summary",
			status.String(), direction.String())
	}

	return summary
}

// NotifyFile creates a pop-up gui notification, in case of incoming files, filename should be a full path
// (download path + filename), so that it can be opened by the user.
func (nm *NotificationManager) NotifyFile(filename string, direction pb.Direction, status pb.Status) {
	summary := fileStatusToNotificationSummary(direction, status)
	b := nm.n.Alert(filename).Summary(summary)
	if direction == pb.Direction_INCOMING && status == pb.Status_SUCCESS {
		b = b.Action(actionKeyOpenFile, "Open", func() { nm.openFileFunc(filename) })
	}
	b.Show()
}

// acceptTransfer accepts transferID, generates notifications on failure
func (nm *NotificationManager) acceptTransfer(transferID string) {
	transfer, err := nm.eventManager.AcceptTransfer(transferID,
		nm.defaultDownloadDir,
		[]string{})

	notificationSummary := acceptFailedNotificationSummary
	if err == ErrTransferCanceledByPeer {
		notificationSummary = transferCanceledByPeerNotificationSummary
	}

	if err != nil {
		nm.n.
			Alert(acceptErrorToNotificationBody(err)).
			Summary(notificationSummary).
			Show()
		return
	}

	for _, file := range transfer.Files {
		if err = nm.fileshare.Accept(transferID, nm.defaultDownloadDir, file.Id); err != nil {
			nm.n.
				Alert(file.Id).
				Summary(acceptFileFailedNotificationSummary).
				Show()
		}
	}

	if err != nil {
		log.Error("Failed to accept some files: ", err)
	}
}

// cancelTransfer cancels transferID, generates error notification on failure
func (nm *NotificationManager) cancelTransfer(transferID string) {
	transfer, err := nm.eventManager.GetTransfer(transferID)
	if err != nil {
		log.Error("Failed to cancel transfer from notification manager: ", err)
		nm.n.
			Alert(genericError).
			Summary(cancelFailedNotificationSummary).
			Show()
		return
	}

	if transfer.Status != pb.Status_ONGOING && transfer.Status != pb.Status_REQUESTED {
		if transfer.Status == pb.Status_CANCELED_BY_PEER {
			nm.n.
				Alert(transferCanceledByPeerNotificationBody).
				Summary(transferCanceledByPeerNotificationSummary).
				Show()
			return
		}
		nm.n.Alert(transferInvalidated).Summary(cancelFailedNotificationSummary).Show()
		return
	}

	if err := nm.fileshare.Finalize(transferID); err != nil {
		log.Error("Failed to cancel transfer from notification manager: ", err)
		nm.n.Alert(err.Error()).Summary(cancelFailedNotificationSummary).Show()
	}
}

// NotifyNewTransfer creates a pop-up gui notification
func (nm *NotificationManager) NotifyNewTransfer(transferID string, peer string) {
	nm.n.Alert(fmt.Sprintf(notifyNewTransferBody, transferID, peer)).
		Summary(notifyNewTransferSummary).
		Action(actionKeyAcceptTransfer, transferAcceptAction, func() { nm.acceptTransfer(transferID) }).
		Action(actionKeyCancelTransfer, transferCancelAction, func() { nm.cancelTransfer(transferID) }).
		Show()
}

// NotifyNewAutoacceptTransfer creates a pop-up gui notification
func (nm *NotificationManager) NotifyNewAutoacceptTransfer(
	transferID string,
	peer string,
) {
	body := fmt.Sprintf(notifyNewTransferBody, transferID, peer)
	nm.n.Alert(body).Summary(notifyNewAutoacceptTransfer).Show()
}

// NotifyAutoacceptFailed creates a pop-up gui notification
func (nm *NotificationManager) NotifyAutoacceptFailed(transferID string, peer string, reason error) {
	transferInfo := fmt.Sprintf(notifyNewTransferBody, transferID, peer)
	body := fmt.Sprintf("%s\n%s", acceptErrorToNotificationBody(reason), transferInfo)
	nm.n.Alert(body).Summary(notifyAutoacceptFailed).Show()
}
