package fileshare

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"golang.org/x/exp/slices"

	"google.golang.org/protobuf/proto"
)

// Handleable errors
var (
	ErrTransferNotFound               = errors.New("transfer not found")
	ErrFileNotFound                   = errors.New("file not found")
	ErrTransferAlreadyAccepted        = errors.New("can't accept already accepted transfer")
	ErrTransferAcceptOutgoing         = errors.New("can't accept outgoing transfer")
	ErrSizeLimitExceeded              = errors.New("provided size limit exceeded")
	ErrAcceptDirNotFound              = errors.New("accept directory not found")
	ErrAcceptDirIsASymlink            = errors.New("accept directory is a symlink")
	ErrAcceptDirIsNotADirectory       = errors.New("accept directory is not a directory")
	ErrNoPermissionsToAcceptDirectory = errors.New("no permissions to accept directory")
	ErrNotificationsAlreadyEnabled    = errors.New("notifications already enabled")
	ErrNotificationsAlreadyDisabled   = errors.New("notifications already disabled")
	ErrTransferCanceledByPeer         = errors.New("transfer has been canceled by peer")
)

// EventManager is responsible for libdrop event handling.
// It keeps transfer state, distributes events to further subscribers, and uses Storage for
// transfer state persistence.
// Thread safe.
type EventManager struct {
	mutex     sync.Mutex
	isProd    bool
	transfers map[string]*pb.Transfer // key is transfer ID
	// stores transfer status notification channels, added by Subscribe, removed by Unsubscribe when TransferFinished event is received
	transferSubscriptions map[string]chan TransferProgressInfo
	storage               Storage
	meshClient            meshpb.MeshnetClient
	fileshare             Fileshare
	osInfo                OsInfo
	filesystem            Filesystem
	notificationManager   *NotificationManager
	defaultDownloadDir    string
}

// NewEventManager loads transfer state from storage, or creates empty state if loading fails.
func NewEventManager(
	isProd bool,
	storage Storage,
	meshClient meshpb.MeshnetClient,
	osInfo OsInfo,
	filesystem Filesystem,
	defaultDownloadDir string) *EventManager {
	loadedTransfers, err := storage.Load()
	if err != nil {
		log.Printf("couldn't load transfer history from storage: %s", err)
		loadedTransfers = map[string]*pb.Transfer{}
	}

	return &EventManager{
		isProd:                isProd,
		transfers:             loadedTransfers,
		transferSubscriptions: map[string]chan TransferProgressInfo{},
		storage:               storage,
		meshClient:            meshClient,
		osInfo:                osInfo,
		filesystem:            filesystem,
		defaultDownloadDir:    defaultDownloadDir,
	}
}

func (em *EventManager) SetFileshare(fileshare Fileshare) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.fileshare = fileshare
}

func (em *EventManager) EnableNotifications(fileshare Fileshare) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.notificationManager != nil {
		return ErrNotificationsAlreadyEnabled
	}

	notificationManager, err := NewNotificationManager(fileshare, em)
	if err != nil {
		return err
	}
	em.notificationManager = notificationManager

	return nil
}

func (em *EventManager) DisableNotifications() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.notificationManager == nil {
		return ErrNotificationsAlreadyDisabled
	}

	if err := em.notificationManager.notifier.Close(); err != nil {
		return fmt.Errorf("failed to disable notifier: %s", err)
	}
	em.notificationManager = nil

	return nil
}

// EventFunc processes events and updates transfers state.
// It should be passed directly to libdrop to be called on events.
func (em *EventManager) EventFunc(eventJSON string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if !em.isProd {
		log.Printf("DROP EVENT: %s", eventJSON)
	}

	var genericEvent genericEvent
	err := json.Unmarshal([]byte(eventJSON), &genericEvent)
	if err != nil {
		log.Printf("unmarshalling drop event: %s\n%s", err, eventJSON)
		return
	}

	switch genericEvent.Type {
	case requestReceived:
		em.handleRequestReceivedEvent(genericEvent.Data)
	case requestQueued:
		// sending peer got event
		// libdrop emits this event later, so, before that eventManager.NewOutgoingTransfer(...) has to be invoked
		var event requestQueuedEvent
		err := json.Unmarshal(genericEvent.Data, &event)
		if err != nil {
			log.Printf("unmarshalling drop event: %s", err)
			return
		}
		transfer, ok := em.transfers[event.TransferID]
		if !ok {
			transfer = NewOutgoingTransfer(event.TransferID, event.Peer, "")
			em.transfers[event.TransferID] = transfer
		}
		transfer.Files = event.Files
		for _, file := range transfer.Files {
			file.Status = pb.Status_REQUESTED
		}
		if err := em.storage.Save(em.transfers); err != nil {
			log.Printf("writing file transfer history: %s", err)
		}
	case transferStarted:
		em.handleTransferStartedEvent(genericEvent.Data)
	case transferProgress:
		em.handleTransferProgressEvent(genericEvent.Data)
	case transferFinished:
		em.handleTransferFinishedEvent(genericEvent.Data)
	default:
		log.Println("DROP EVENT: ", eventJSON)
	}
}

func (em *EventManager) handleRequestReceivedEvent(eventJson json.RawMessage) {
	var event requestReceivedEvent
	err := json.Unmarshal(eventJson, &event)
	if err != nil {
		log.Printf("unmarshalling drop event: %s", err)
		return
	}

	peer, err := getPeerByIP(em.meshClient, event.Peer)
	if err != nil {
		log.Println("failed to retrieve peer requesting transfer: ", err.Error())
		return
	}
	if !peer.DoIAllowFileshare {
		return
	}

	em.transfers[event.TransferID] = NewIncomingTransfer(event.TransferID, event.Peer, event.Files)
	if err := em.storage.Save(em.transfers); err != nil {
		log.Printf("writing file transfer history: %s", err)
		return
	}

	if !peer.AlwaysAcceptFiles {
		if em.notificationManager != nil {
			em.notificationManager.NotifyNewTransfer(event.TransferID, peer.Hostname)
		}
		return
	}

	// default download directory not set
	if em.defaultDownloadDir == "" {
		return
	}

	transfer, err := em.acceptTransfer(event.TransferID, em.defaultDownloadDir, []string{})

	if err != nil {
		log.Println("failed to autoaccept transfer: ", err.Error())
		if em.notificationManager != nil {
			em.notificationManager.NotifyAutoacceptFailed(event.TransferID, peer.Hostname, err)
		}
		return
	}

	for _, file := range transfer.Files {
		err = em.fileshare.Accept(event.TransferID, em.defaultDownloadDir, file.Id)
	}

	if err != nil {
		log.Println("failed to autoaccept all files: ", err)
	}

	if em.notificationManager != nil {
		em.notificationManager.NotifyNewAutoacceptTransfer(event.TransferID, peer.Hostname)
	}
}

func (em *EventManager) handleTransferStartedEvent(eventJson json.RawMessage) {
	var event transferStartedEvent
	err := json.Unmarshal(eventJson, &event)
	if err != nil {
		log.Printf("unmarshalling drop event: %s", err)
		return
	}
	transfer, ok := em.transfers[event.TransferID]
	if !ok {
		log.Printf("transfer %s from transferStarted event not found", event.TransferID)
		return
	}

	if file := FindTransferFileByID(transfer, event.FileID); file != nil {
		transfer.TotalSize += file.Size
	} else {
		log.Printf("can't find file from transferStarted event in transfer %s", transfer.Id)
	}

	if err := em.storage.Save(em.transfers); err != nil {
		log.Printf("writing file transfer history: %s", err)
	}
}

func (em *EventManager) handleTransferProgressEvent(eventJson json.RawMessage) {
	// transfer progress per file
	var event transferProgressEvent
	err := json.Unmarshal(eventJson, &event)
	if err != nil {
		log.Printf("unmarshalling drop event: %s", err)
		return
	}
	transfer, ok := em.transfers[event.TransferID]
	if !ok {
		log.Printf("transfer %s from transferProgress event not found", event.TransferID)
		return
	}
	// mark corresponding file progress in transfer data structure
	if file := FindTransferFileByID(transfer, event.FileID); file != nil {
		transfer.Status = pb.Status_ONGOING
		file.Status = pb.Status_ONGOING
		transfer.TotalTransferred += event.Transferred - file.Transferred // add only delta
		file.Transferred = event.Transferred
	} else {
		// transfer does not contain reported file?!
		log.Printf("transfer %s transferProgress event reported file that is not included in transfer",
			event.TransferID)
		return
	}
	if progressCh, ok := em.transferSubscriptions[transfer.Id]; ok {
		var progressPercent uint32
		if transfer.TotalSize > 0 { // transfer progress percentage should be reported to subscriber
			progressPercent = uint32(float64(transfer.TotalTransferred) / float64(transfer.TotalSize) * 100)
		}
		progressCh <- TransferProgressInfo{
			TransferID:  event.TransferID,
			Transferred: progressPercent,
			Status:      pb.Status_ONGOING,
		}
	}
}

func (em *EventManager) handleTransferFinishedEvent(eventJSON json.RawMessage) {
	var event transferFinishedEvent
	err := json.Unmarshal(eventJSON, &event)
	if err != nil {
		log.Printf("unmarshalling drop event: %s", err)
		return
	}
	transfer, ok := em.transfers[event.TransferID]

	if !ok {
		transfer = NewOutgoingTransfer(event.TransferID, "peer", "path")
		em.transfers[event.TransferID] = transfer
	}

	// Currently libdrop will not clean up the transfer after transferring all of the files, so we have to
	// cancel it manually after all of the files have finished downloading/uploading. This will generate
	// a TransferCanceled event, which we don't care about and don't want to have any impact on internal state
	// If transfer has been finalized(canceled), we return early. We should be able to remove this check and
	// the Finalized flag once cleanup is implemented in libdrop.
	if transfer.Finalized {
		return
	}

	var newFileStatus pb.Status
	switch event.Reason {
	case transferCanceled:
		ForAllFiles(transfer.Files, func(f *pb.File) {
			if f.Status == pb.Status_REQUESTED || f.Status == pb.Status_ONGOING {
				f.Status = pb.Status_CANCELED
			}
		})
		if event.Data.ByPeer {
			transfer.Status = pb.Status_CANCELED_BY_PEER
		} else {
			transfer.Status = pb.Status_CANCELED
		}
		em.finalizeTransfer(transfer)
		return
	case transferFailed:
		transfer.Status = pb.Status_FINISHED_WITH_ERRORS
		em.finalizeTransfer(transfer)
		return
	case fileDownloaded:
		fallthrough
	case fileUploaded:
		newFileStatus = pb.Status_SUCCESS
	case fileCanceled:
		newFileStatus = pb.Status_CANCELED
	case fileFailed:
		newFileStatus = event.Data.Status
	default:
		log.Printf("Unknown reason for transfer finished event: %s", event.Reason)
		return
	}

	file := FindTransferFileByID(transfer, event.Data.File)
	if em.notificationManager != nil && file != nil {
		em.notificationManager.NotifyFile(
			filepath.Join(transfer.Path, file.Path),
			transfer.Direction,
			newFileStatus,
		)
	}

	if err := SetFileStatus(transfer.Files, event.Data.File, newFileStatus); err != nil {
		log.Printf("Failed to set file status: %s", err)
		return
	}
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)

	if isTransferFinished(transfer) {
		em.finalizeTransfer(transfer)
		if transfer.Direction == pb.Direction_INCOMING {
			if err := em.fileshare.Cancel(event.TransferID); err != nil {
				log.Printf("failed to finalize transfer %s: %s", event.TransferID, err)
			}
		}
	}
}

func (em *EventManager) finalizeTransfer(transfer *pb.Transfer) {
	if progressCh, ok := em.transferSubscriptions[transfer.Id]; ok {
		progressCh <- TransferProgressInfo{
			TransferID: transfer.Id,
			Status:     transfer.Status,
		}
		// unsubscribe finished transfer
		close(progressCh)
		delete(em.transferSubscriptions, transfer.Id)
	}
	transfer.Finalized = true
}

// NewOutgoingTransfer used when we are sending files, because libdrop only emits the event a bit later
func (em *EventManager) NewOutgoingTransfer(id, peer, path string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if transfer, ok := em.transfers[id]; ok {
		transfer.Peer = peer
		transfer.Path = path
	} else {
		em.transfers[id] = NewOutgoingTransfer(id, peer, path)
	}
}

// GetTransfers is used for listing transfers.
// Returned transfers are sorted by date created from oldest to newest.
// Returns copies of transfers, modifying them will not change EventManager state.
func (em *EventManager) GetTransfers() []*pb.Transfer {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	transfers := make([]*pb.Transfer, 0, len(em.transfers))
	for _, transfer := range em.transfers {
		transferCloned, ok := proto.Clone(transfer).(*pb.Transfer)
		if !ok {
			log.Printf("failed to cast cloned transfer %s", transfer.Id)
			continue
		}
		transfers = append(transfers, transferCloned)
	}

	sort.Slice(transfers, func(i int, j int) bool {
		return transfers[i].Created.AsTime().Before(transfers[j].Created.AsTime())
	})

	return transfers
}

// GetTransfer by ID.
// Returns a copy, modifying it will not change EventManager state.
func (em *EventManager) GetTransfer(transferID string) (*pb.Transfer, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	transfer, ok := em.transfers[transferID]
	if !ok {
		return nil, ErrTransferNotFound
	}

	transferCloned, ok := proto.Clone(transfer).(*pb.Transfer)
	if !ok {
		return transferCloned, fmt.Errorf("failed to cast cloned transfer %s", transferID)
	}
	return transferCloned, nil
}

// AcceptTransfer changes transfer status to reflect that it is being downloaded
func (em *EventManager) AcceptTransfer(
	transferID string,
	path string,
	filePaths []string,
) (*pb.Transfer, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	return em.acceptTransfer(transferID, path, filePaths)
}
func (em *EventManager) acceptTransfer(
	transferID string,
	path string,
	filePaths []string,
) (*pb.Transfer, error) {
	fileInfo, err := em.filesystem.Lstat(path)

	if err != nil {
		return nil, ErrAcceptDirNotFound
	}

	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		return nil, ErrAcceptDirIsASymlink
	}

	if !fileInfo.IsDir() {
		return nil, ErrAcceptDirIsNotADirectory
	}

	userInfo, err := em.osInfo.CurrentUser()
	if err != nil {
		log.Printf("getting user info: %s", err)
		return nil, ErrNoPermissionsToAcceptDirectory
	}

	userGroups, err := em.osInfo.GetGroupIds(userInfo)
	if err != nil {
		log.Printf("getting user groups: %s", err)
		return nil, ErrNoPermissionsToAcceptDirectory
	}

	if !isFileWriteable(fileInfo, userInfo, userGroups) {
		return nil, ErrNoPermissionsToAcceptDirectory
	}

	transfer, ok := em.transfers[transferID]
	if !ok {
		return nil, ErrTransferNotFound
	}
	if transfer.Direction != pb.Direction_INCOMING {
		return nil, ErrTransferAcceptOutgoing
	}
	if transfer.Status == pb.Status_CANCELED_BY_PEER {
		return nil, ErrTransferCanceledByPeer
	}
	if transfer.Status != pb.Status_REQUESTED {
		return nil, ErrTransferAlreadyAccepted
	}

	var files []*pb.File
	if len(filePaths) == 0 {
		files = transfer.Files // All files were accepted
	} else {
		for _, filePath := range filePaths {
			acceptedFiles := GetTransferFilesByPathPrefix(transfer, filePath)
			if acceptedFiles == nil {
				return nil, ErrFileNotFound
			}
			files = append(files, acceptedFiles...)
		}
	}

	var totalSize uint64
	ForAllFiles(files, func(f *pb.File) {
		totalSize += f.Size
	})

	statfs, err := em.filesystem.Statfs(path)
	if err != nil {
		log.Printf("doing statfs: %s", err)
		return nil, ErrSizeLimitExceeded
	}

	if totalSize > statfs.Bavail*uint64(statfs.Bsize) {
		return nil, ErrSizeLimitExceeded
	}

	transfer.Path = path
	transfer.Status = pb.Status_ONGOING

	return transfer, nil
}

func isFileWriteable(fileInfo fs.FileInfo, user *user.User, gids []string) bool {
	var ownerUID int
	var ownerGID int
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		ownerUID = int(stat.Uid)
		ownerGID = int(stat.Gid)
	} else {
		return false
	}

	uid, err := strconv.Atoi(user.Uid)

	if err != nil {
		log.Printf("Failed to convert uid %s to int: %s", user.Uid, err)
		return false
	}

	isOwner := uid == ownerUID

	if isOwner {
		return fileInfo.Mode().Perm()&os.FileMode(0200) != 0
	}

	ownerGIDStr := strconv.Itoa(ownerGID)
	gidIndex := slices.Index(gids, ownerGIDStr)
	isGroup := gidIndex != -1
	if isGroup {
		return fileInfo.Mode().Perm()&os.FileMode(0020) != 0
	}

	return fileInfo.Mode().Perm()&os.FileMode(0002) != 0
}

// SetTransferStatus manually
func (em *EventManager) SetTransferStatus(transferID string, status pb.Status) error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	transfer, ok := em.transfers[transferID]
	if !ok {
		return ErrTransferNotFound
	}

	transfer.Status = status
	return nil
}

// TransferProgressInfo info to report to the user
type TransferProgressInfo struct {
	TransferID  string
	Transferred uint32 // percent of transferred bytes
	Status      pb.Status
}

// Subscribe is used to track progress.
func (em *EventManager) Subscribe(id string) chan TransferProgressInfo {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.transferSubscriptions[id] = make(chan TransferProgressInfo)

	return em.transferSubscriptions[id]
}

// SetFileStatus manually
func (em *EventManager) SetFileStatus(transferID string, fileID string, status pb.Status) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	transfer, ok := em.transfers[transferID]

	if !ok {
		log.Printf("Cannot set file status, transfer %s not found", transferID)
		return
	}

	if err := SetFileStatus(transfer.Files, fileID, status); err != nil {
		log.Printf("Failed to set file status: %s", err)
		return
	}
}
