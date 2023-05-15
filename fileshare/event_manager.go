package fileshare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"golang.org/x/exp/slices"

	"google.golang.org/protobuf/proto"
)

// Handleable errors
var (
	ErrTransferNotFound        = errors.New("transfer not found")
	ErrFileNotFound            = errors.New("file not found")
	ErrTransferAlreadyAccepted = errors.New("can't accept already accepted transfer")
	ErrTransferAcceptOutgoing  = errors.New("can't accept outgoing transfer")
	ErrSizeLimitExceeded       = errors.New("provided size limit exceeded")
)

// EventManager is responsible for libdrop event handling.
// It keeps transfer state, distributes events to further subscribers, and uses Storage for
// transfer state persistence.
// Thread safe.
type EventManager struct {
	mutex     sync.Mutex
	transfers map[string]*pb.Transfer // key is transfer ID
	// stores transfer status notification channels, added by Subscribe, removed by Unsubscribe when TransferFinished event is received
	transferSubscriptions map[string]chan TransferProgressInfo
	storage               Storage
	meshClient            meshpb.MeshnetClient
	// CancelFunc is called when transfer is completed in order to finalize the transfer
	CancelFunc          func(transferID string) error
	notificationManager *NotificationManager
}

// NewEventManager loads transfer state from storage, or creates empty state if loading fails.
func NewEventManager(storage Storage, meshClient meshpb.MeshnetClient, notificationManager *NotificationManager) *EventManager {
	loadedTransfers, err := storage.Load()
	if err != nil {
		log.Printf("couldn't load transfer history from storage: %s", err)
		loadedTransfers = map[string]*pb.Transfer{}
	}

	return &EventManager{
		transfers:             loadedTransfers,
		transferSubscriptions: map[string]chan TransferProgressInfo{},
		storage:               storage,
		meshClient:            meshClient,
		notificationManager:   notificationManager,
	}
}

func (em *EventManager) AreNotificationsEnabled() bool {
	return em.notificationManager != nil
}

func (em *EventManager) EnableNotifications() error {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	if em.notificationManager != nil {
		return nil
	}

	notificationManager, err := NewNotificationManager()
	if err != nil {
		return err
	}
	em.notificationManager = notificationManager

	return nil
}

func (em *EventManager) DisableNotifications() {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	if err := em.notificationManager.notifier.Close(); err != nil {
		log.Println(err)
	}
	em.notificationManager = nil
}

func (em *EventManager) isFileshareFromPeerAllowed(peerIP string) bool {
	resp, err := em.meshClient.GetPeers(context.Background(), &meshpb.Empty{})
	if err != nil {
		log.Printf("failed to get peers when validating permissions: %s", err)
		return false
	}

	switch resp := resp.Response.(type) {
	case *meshpb.GetPeersResponse_Peers:
		peers := resp.Peers.External
		peers = append(peers, resp.Peers.Local...)
		peerIndex := slices.IndexFunc(peers, func(peer *meshpb.Peer) bool {
			return peer.Ip == peerIP
		})

		if peerIndex == -1 {
			log.Printf("unknown peer %s found when validating permissions", peerIP)
			return false
		}

		if !peers[peerIndex].DoIAllowFileshare {
			return false
		}

		return true
	case *meshpb.GetPeersResponse_ServiceErrorCode:
		log.Printf("GetPeers failed, service error: %s", meshpb.ServiceErrorCode_name[int32(resp.ServiceErrorCode)])
		return false
	case *meshpb.GetPeersResponse_MeshnetErrorCode:
		log.Printf("GetPeers failed, meshnet error: %s", meshpb.ServiceErrorCode_name[int32(resp.MeshnetErrorCode)])
		return false
	default:
		log.Printf("GetPeers failed, unknown error")
		return false
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
			transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
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

	if em.notificationManager != nil {
		em.notificationManager.Notify(filepath.Join(transfer.Path, event.Data.File), newFileStatus, transfer.Direction)
	}

	if err := SetFileStatus(transfer.Files, event.Data.File, newFileStatus); err != nil {
		log.Printf("Failed to set file status: %s", err)
		return
	}
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)

	if isTransferFinished(transfer) {
		em.finalizeTransfer(transfer)
		if transfer.Direction == pb.Direction_INCOMING {
			if err := em.CancelFunc(event.TransferID); err != nil {
				log.Printf("failed to finalize transfer %s: %s", event.TransferID, err)
			}
		}
	}
}

// EventFunc processes events and updates transfers state.
// It should be passed directly to libdrop to be called on events.
func (em *EventManager) EventFunc(eventJSON string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	var genericEvent genericEvent
	err := json.Unmarshal([]byte(eventJSON), &genericEvent)
	if err != nil {
		log.Printf("unmarshalling drop event: %s\n%s", err, eventJSON)
		return
	}

	switch genericEvent.Type {
	case requestReceived:
		// receiving peer got event
		var event requestReceivedEvent
		err := json.Unmarshal(genericEvent.Data, &event)
		if err != nil {
			log.Printf("unmarshalling drop event: %s", err)
			return
		}
		if em.isFileshareFromPeerAllowed(event.Peer) {
			em.transfers[event.TransferID] = NewIncomingTransfer(event.TransferID, event.Peer, event.Files)
			if err := em.storage.Save(em.transfers); err != nil {
				log.Printf("writing file transfer history: %s", err)
			}
		}
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
			log.Printf("transfer %s from requestQueued event not found", event.TransferID)
			return
		}
		SetTransferFiles(transfer, event.Files)
		if err := em.storage.Save(em.transfers); err != nil {
			log.Printf("writing file transfer history: %s", err)
		}
	case transferStarted:
		var event transferStartedEvent
		err := json.Unmarshal(genericEvent.Data, &event)
		if err != nil {
			log.Printf("unmarshalling drop event: %s", err)
			return
		}
		transfer, ok := em.transfers[event.TransferID]
		if !ok {
			log.Printf("transfer %s from transferStarted event not found", event.TransferID)
			return
		}

		if file := FindTransferFile(transfer, event.FileID); file != nil {
			transfer.TotalSize += file.Size
		} else {
			log.Printf("can't find file from transferStarted event in transfer %s", transfer.Id)
		}

		if err := em.storage.Save(em.transfers); err != nil {
			log.Printf("writing file transfer history: %s", err)
		}
	case transferProgress:
		// transfer progress per file
		var event transferProgressEvent
		err := json.Unmarshal(genericEvent.Data, &event)
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
		if file := FindTransferFile(transfer, event.FileID); file != nil {
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
	case transferFinished:
		em.handleTransferFinishedEvent(genericEvent.Data)
	default:
		log.Println("DROP EVENT: ", eventJSON)
	}
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
	fileIDs []string,
	sizeLimit uint64,
) (*pb.Transfer, error) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	transfer, ok := em.transfers[transferID]
	if !ok {
		return nil, ErrTransferNotFound
	}
	if transfer.Direction != pb.Direction_INCOMING {
		return nil, ErrTransferAcceptOutgoing
	}
	if transfer.Status != pb.Status_REQUESTED {
		return nil, ErrTransferAlreadyAccepted
	}

	var files []*pb.File
	for _, fileID := range fileIDs {
		file := FindTransferFile(transfer, fileID)
		if file == nil {
			return nil, ErrFileNotFound
		}
		files = append(files, file)
	}

	if len(fileIDs) == 0 {
		files = transfer.Files // All files were accepted
	}
	var totalSize uint64
	ForAllFiles(files, func(f *pb.File) {
		totalSize += f.Size
	})
	if totalSize > sizeLimit {
		return nil, ErrSizeLimitExceeded
	}

	transfer.Path = path
	transfer.Status = pb.Status_ONGOING

	return transfer, nil
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
