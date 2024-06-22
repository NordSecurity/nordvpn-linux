package fileshare

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"sort"
	"strconv"
	"sync"
	"syscall"

	norddrop "github.com/NordSecurity/libdrop-go/v7"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"golang.org/x/exp/slices"
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
	ErrTransferCanceledByUs           = errors.New("transfer has been canceled by us")
)

// EventManager is responsible for libdrop event handling.
// It keeps transfer state, distributes events to further subscribers, and uses Storage for
// transfer state persistence.
// Thread safe.
type EventManager struct {
	mutex  sync.Mutex
	isProd bool
	// Key is transfer ID.
	// If transfer doesn't exist it may have just started or resumed.
	// Must delete transfers when they are finished.
	liveTransfers map[string]*LiveTransfer
	// stores transfer status notification channels added by Subscribe,
	// removed by Unsubscribe when TransferFinished event is received
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
	meshClient meshpb.MeshnetClient,
	osInfo OsInfo,
	filesystem Filesystem,
	defaultDownloadDir string,
) *EventManager {
	return &EventManager{
		isProd:                isProd,
		liveTransfers:         map[string]*LiveTransfer{},
		transferSubscriptions: map[string]chan TransferProgressInfo{},
		meshClient:            meshClient,
		osInfo:                osInfo,
		filesystem:            filesystem,
		defaultDownloadDir:    defaultDownloadDir,
	}
}

// SetFileshare must be called before using event manager.
// Necessary because of circular dependency between event manager and libDrop.
func (em *EventManager) SetFileshare(fileshare Fileshare) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.fileshare = fileshare
}

// SetStorage must be called before using event manager.
// Necessary because of circular dependency between event manager and libDrop.
func (em *EventManager) SetStorage(storage Storage) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	em.storage = storage
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

// OnEvent processes events and handles live transfer state.
func (em *EventManager) OnEvent(event norddrop.Event) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if !em.isProd {
		log.Printf(internal.InfoPrefix+" DROP EVENT: %s\n", internal.EventToString(event))
	}

	switch ev := event.Kind.(type) {
	case norddrop.EventKindRequestReceived:
		em.handleRequestReceivedEvent(ev)
	case norddrop.EventKindRequestQueued: // ignore
	case norddrop.EventKindFileStarted: // ignore
	case norddrop.EventKindFileProgress:
		em.handleFileProgressEvent(ev)
	case norddrop.EventKindTransferFailed:
		em.handleTransferFailedEvent(ev)
	case norddrop.EventKindTransferFinalized:
		em.handleTransferFinalizedEvent(ev)
	case norddrop.EventKindFileDownloaded:
		em.handleFileDownloadedEvent(ev)
	case norddrop.EventKindFileUploaded:
		em.handleFileUploadedEvent(ev)
	case norddrop.EventKindFileRejected:
		em.handleFileRejectedEvent(ev)
	case norddrop.EventKindFileFailed:
		em.handleFileFailedEvent(ev)
	default:
		log.Printf(internal.WarningPrefix+" unsupported libdrop event: %T\n", ev)
	}
}

func (em *EventManager) handleRequestReceivedEvent(event norddrop.EventKindRequestReceived) {
	peer, err := getPeerByIP(em.meshClient, event.Peer)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to retrieve peer requesting transfer:", err)
		return
	}
	if !peer.DoIAllowFileshare {
		// This can only happen in the case of abuse, since clients shouldn't allow sending transfers
		// to peers which don't allow that.
		if err := em.fileshare.Finalize(event.TransferId); err != nil {
			log.Printf(internal.WarningPrefix+" failed to auto-reject transfer %s: %s\n", event.TransferId, err)
		}
		return
	}
	if !peer.AlwaysAcceptFiles {
		if em.notificationManager != nil {
			em.notificationManager.NotifyNewTransfer(event.TransferId, peer.Hostname)
		}
		return
	}

	// default download directory not set
	if em.defaultDownloadDir == "" {
		return
	}

	transfer, err := em.acceptTransfer(event.TransferId, em.defaultDownloadDir, []string{})
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to autoaccept transfer:", err)
		if em.notificationManager != nil {
			em.notificationManager.NotifyAutoacceptFailed(event.TransferId, peer.Hostname, err)
		}
		return
	}

	for _, file := range transfer.Files {
		err = em.fileshare.Accept(event.TransferId, em.defaultDownloadDir, file.Id)
		if err != nil {
			log.Println(internal.WarningPrefix, "failed to autoaccept file:", err)
		}
	}

	if em.notificationManager != nil {
		em.notificationManager.NotifyNewAutoacceptTransfer(event.TransferId, peer.Hostname)
	}
}

func (em *EventManager) handleFileProgressEvent(event norddrop.EventKindFileProgress) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}

	file, ok := transfer.Files[event.FileId]
	if !ok {
		log.Printf(internal.ErrorPrefix+" file %s from FileProgress event not found in transfer %s\n",
			event.FileId, transfer.ID)
		return
	}

	transfer.TotalTransferred += event.Transferred - file.Transferred // add only delta
	file.Transferred = event.Transferred

	if progressCh, ok := em.transferSubscriptions[transfer.ID]; ok {
		var progressPercent uint32
		if transfer.TotalSize > 0 { // transfer progress percentage should be reported to subscriber
			progressPercent = uint32(float64(transfer.TotalTransferred) / float64(transfer.TotalSize) * 100)
		}
		progressCh <- TransferProgressInfo{
			TransferID:  event.TransferId,
			Transferred: progressPercent,
			Status:      pb.Status_ONGOING,
		}
	}
}

func (em *EventManager) handleFileDownloadedEvent(event norddrop.EventKindFileDownloaded) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}

	file, ok := transfer.Files[event.FileId]
	if !ok {
		log.Printf(internal.ErrorPrefix+" file %s from FileDownloaded event not found in transfer %s\n",
			event.FileId, transfer.ID)
		return
	}
	file.Finished = true

	fileStatusInNotification := pb.Status_SUCCESS
	if em.notificationManager != nil && file != nil {
		em.notificationManager.NotifyFile(
			event.FinalPath,
			transfer.Direction,
			fileStatusInNotification,
		)
	}

	em.finalizeFinishedTransfer(transfer)
}

func (em *EventManager) finalizeFinishedTransfer(transfer *LiveTransfer) {
	// Libdrop will not clean up the transfer after transferring all of the files, so we have to
	// finalize it manually - after all of the files have finished downloading/uploading or are
	// failed/rejected.
	// This will generate a [norddrop.EventKindTransferFinalized] event, which is processed in
	// [EventManager.handleTransferFinalizedEvent] and will trigger the finalization of transfer.
	if isLiveTransferFinished(transfer) && transfer.Direction == pb.Direction_INCOMING {
		if err := em.fileshare.Finalize(transfer.ID); err != nil {
			log.Printf(internal.WarningPrefix+" failed to finalize transfer %s: %s\n", transfer.ID, err)
		}
	}
}

func (em *EventManager) handleFileUploadedEvent(event norddrop.EventKindFileUploaded) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}

	file, ok := transfer.Files[event.FileId]
	if !ok {
		log.Printf(internal.ErrorPrefix+" file %s from FileUploaded event not found in transfer %s\n",
			event.FileId, transfer.ID)
		return
	}
	file.Finished = true

	fileStatusInNotification := pb.Status_SUCCESS
	if em.notificationManager != nil && file != nil {
		em.notificationManager.NotifyFile(
			file.FullPath,
			transfer.Direction,
			fileStatusInNotification,
		)
	}

	em.finalizeFinishedTransfer(transfer)
}

func (em *EventManager) handleFileFailedEvent(event norddrop.EventKindFileFailed) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}

	file, ok := transfer.Files[event.FileId]
	if !ok {
		log.Printf(internal.ErrorPrefix+" file %s from FileFailed event not found in transfer %s\n",
			event.FileId, transfer.ID)
		return
	}
	file.Finished = true

	fileStatusInNotification := pb.Status(event.Status.Status)
	removeFileFromLiveTransfer(transfer, file)
	if em.notificationManager != nil && file != nil {
		em.notificationManager.NotifyFile(
			file.FullPath,
			transfer.Direction,
			fileStatusInNotification,
		)
	}

	em.finalizeFinishedTransfer(transfer)
}

func (em *EventManager) handleFileRejectedEvent(event norddrop.EventKindFileRejected) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}

	file, ok := transfer.Files[event.FileId]
	if !ok {
		log.Printf(internal.ErrorPrefix+" file %s from FileRejected event not found in transfer %s\n",
			event.FileId, transfer.ID)
		return
	}
	file.Finished = true

	fileStatusInNotification := pb.Status_CANCELED
	removeFileFromLiveTransfer(transfer, file)
	if em.notificationManager != nil && file != nil {
		em.notificationManager.NotifyFile(
			file.FullPath,
			transfer.Direction,
			fileStatusInNotification,
		)
	}

	em.finalizeFinishedTransfer(transfer)
}

func (em *EventManager) handleTransferFailedEvent(event norddrop.EventKindTransferFailed) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}
	em.finalizeTransfer(transfer, pb.Status(event.Status.Status))
}

func (em *EventManager) handleTransferFinalizedEvent(event norddrop.EventKindTransferFinalized) {
	transfer, err := em.getLiveTransfer(event.TransferId)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to get live transfer:", err)
		return
	}

	var status pb.Status
	switch {
	case isLiveTransferFinished(transfer):
		// Automatic cancel due to transfer finalization
		storageTransfer, err := getTransferFromStorage(event.TransferId, em.storage)
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to get transfer from storage:", err)
			return
		}
		status = storageTransfer.Status
	case event.ByPeer:
		status = pb.Status_CANCELED_BY_PEER
	default:
		status = pb.Status_CANCELED
	}
	em.finalizeTransfer(transfer, status)
}

func (em *EventManager) finalizeTransfer(transfer *LiveTransfer, status pb.Status) {
	if progressCh, ok := em.transferSubscriptions[transfer.ID]; ok {
		progressCh <- TransferProgressInfo{
			TransferID: transfer.ID,
			Status:     status,
		}
		// unsubscribe finished transfer
		close(progressCh)
		delete(em.transferSubscriptions, transfer.ID)
	}

	delete(em.liveTransfers, transfer.ID)
}

// GetTransfers is used for listing transfers.
// Returned transfers are sorted by date created from oldest to newest.
func (em *EventManager) GetTransfers() ([]*pb.Transfer, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	storageTransfers, err := em.storage.Load()
	if err != nil {
		return nil, fmt.Errorf("loading transfers from storage: %s", err)
	}

	transfers := make([]*pb.Transfer, 0, len(storageTransfers))
	for _, storageTransfer := range storageTransfers {
		updatedTransfer := updateTransferWithLiveData(storageTransfer, em.liveTransfers)
		transfers = append(transfers, updatedTransfer)
	}

	sort.Slice(transfers, func(i int, j int) bool {
		return transfers[i].Created.AsTime().Before(transfers[j].Created.AsTime())
	})

	return transfers, nil
}

// CancelLiveTransfers cancels all ongoing transfers.
func (em *EventManager) CancelLiveTransfers() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	for transferID := range em.liveTransfers {
		err := em.fileshare.Finalize(transferID)
		if err != nil {
			log.Println(internal.WarningPrefix, "failed to cancel live transfer:", err)
		}
	}
}

// GetTransfer by ID.
func (em *EventManager) GetTransfer(transferID string) (*pb.Transfer, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	return em.getTransfer(transferID)
}

func (em *EventManager) getTransfer(transferID string) (*pb.Transfer, error) {
	transfer, err := getTransferFromStorage(transferID, em.storage)
	if err != nil {
		return nil, err
	}
	transfer = updateTransferWithLiveData(transfer, em.liveTransfers)
	return transfer, nil
}

func getTransferFromStorage(id string, storage Storage) (*pb.Transfer, error) {
	storageTransfers, err := storage.Load()
	if err != nil {
		return nil, fmt.Errorf("loading transfers from storage: %s", err)
	}
	storageTransfer, ok := storageTransfers[id]
	if !ok {
		return nil, ErrTransferNotFound
	}
	return storageTransfer, nil
}

// updateTransferWithLiveData used for ongoing transfers because Storage doesn't contain momentary info about transfer
// progress, so update it from liveTransfers
func updateTransferWithLiveData(transfer *pb.Transfer, liveTransfers map[string]*LiveTransfer) *pb.Transfer {
	liveTransfer, ok := liveTransfers[transfer.Id]
	if !ok {
		return transfer
	}

	transfer.TotalTransferred = liveTransfer.TotalTransferred
	for _, file := range transfer.Files {
		liveFile, ok := liveTransfer.Files[file.Id]
		if ok {
			file.Transferred = liveFile.Transferred
		}
	}

	return transfer
}

// AcceptTransfer validates the transfer to ensure it can be accepted
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
		log.Printf(internal.ErrorPrefix+" getting user info: %s\n", err)
		return nil, ErrNoPermissionsToAcceptDirectory
	}

	userGroups, err := em.osInfo.GetGroupIds(userInfo)
	if err != nil {
		log.Printf(internal.ErrorPrefix+" getting user groups: %s\n", err)
		return nil, ErrNoPermissionsToAcceptDirectory
	}

	if !isFileWriteable(fileInfo, userInfo, userGroups) {
		return nil, ErrNoPermissionsToAcceptDirectory
	}

	transfer, err := em.getTransfer(transferID)
	if err != nil {
		return nil, err
	}
	if transfer.Direction != pb.Direction_INCOMING {
		return nil, ErrTransferAcceptOutgoing
	}
	if transfer.Status == pb.Status_CANCELED_BY_PEER {
		return nil, ErrTransferCanceledByPeer
	}
	if transfer.Status == pb.Status_CANCELED {
		return nil, ErrTransferCanceledByUs
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
		log.Printf(internal.ErrorPrefix+" doing statfs: %s\n", err)
		return nil, ErrSizeLimitExceeded
	}

	if totalSize > statfs.Bavail*uint64(statfs.Bsize) {
		return nil, ErrSizeLimitExceeded
	}

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
		log.Printf(internal.ErrorPrefix+" failed to convert uid %s to int: %s\n", user.Uid, err)
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

// TransferProgressInfo info to report to the user
type TransferProgressInfo struct {
	TransferID  string
	Transferred uint32 // percent of transferred bytes
	Status      pb.Status
}

// Subscribe is used to track progress.
func (em *EventManager) Subscribe(id string) <-chan TransferProgressInfo {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.transferSubscriptions[id] = make(chan TransferProgressInfo)

	return em.transferSubscriptions[id]
}

// LiveTransfer to track ongoing transfers live in app based on events
type LiveTransfer struct {
	ID               string
	Direction        pb.Direction
	TotalSize        uint64
	TotalTransferred uint64
	Files            map[string]*LiveFile // Key is ID
}

// LiveFile is part of LiveTransfer
type LiveFile struct {
	ID          string
	FullPath    string
	Size        uint64
	Transferred uint64
	Finished    bool
}

// Returns an existing live transfer or creates a new one if necessary
func (em *EventManager) getLiveTransfer(id string) (*LiveTransfer, error) {
	transfer, ok := em.liveTransfers[id]
	if ok {
		return transfer, nil
	}

	storageTransfer, err := getTransferFromStorage(id, em.storage)
	if err != nil {
		return nil, err
	}

	transfer = &LiveTransfer{
		ID:               storageTransfer.Id,
		Direction:        storageTransfer.Direction,
		TotalSize:        storageTransfer.TotalSize,
		TotalTransferred: storageTransfer.TotalTransferred,
		Files:            map[string]*LiveFile{},
	}
	for _, file := range storageTransfer.Files {
		if isFileTransferred(file) {
			liveFile := &LiveFile{
				ID:          file.Id,
				FullPath:    file.FullPath,
				Size:        file.Size,
				Transferred: file.Transferred,
				Finished:    isFileCompleted(file),
			}
			transfer.Files[file.Id] = liveFile
		}
	}

	em.liveTransfers[transfer.ID] = transfer
	return transfer, nil
}

func isLiveTransferFinished(tr *LiveTransfer) bool {
	for _, file := range tr.Files {
		if !file.Finished {
			return false
		}
	}
	return true
}

// Used when file is canceled or errors out to exclude it from progress calculations
func removeFileFromLiveTransfer(transfer *LiveTransfer, file *LiveFile) {
	transfer.TotalSize -= file.Size
	transfer.TotalTransferred -= file.Transferred
	delete(transfer.Files, file.ID)
}
