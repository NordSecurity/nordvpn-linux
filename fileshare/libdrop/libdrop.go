// Package libdrop wraps libdrop fileshare implementation.
package libdrop

import (
	"errors"
	"fmt"
	"log"
	"net/netip"
	"path/filepath"
	"strings"
	"sync"
	"time"

	norddrop "github.com/NordSecurity/libdrop-go/v8"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Fileshare is the main functional filesharing implementation using norddrop library.
// Thread safe.
type Fileshare struct {
	norddrop     *norddrop.NordDrop
	eventsDbPath string
	storagePath  string
	isProd       bool
	mutex        sync.Mutex
}

func logLevelToPrefix(level norddrop.LogLevel) string {
	switch level {
	case norddrop.LogLevelCritical, norddrop.LogLevelError:
		return internal.ErrorPrefix
	case norddrop.LogLevelWarning:
		return internal.WarningPrefix
	case norddrop.LogLevelDebug, norddrop.LogLevelTrace:
		return internal.DebugPrefix
	case norddrop.LogLevelInfo:
		return internal.InfoPrefix
	default:
		return internal.InfoPrefix
	}
}

type defaultKeyStore struct {
	pubkeyFunc func(string) []byte
	privKey    string
}

func (dks defaultKeyStore) OnPubkey(peer string) *[]byte {
	pubKey := dks.pubkeyFunc(peer)
	return &pubKey
}

func (dks defaultKeyStore) Privkey() []byte {
	return []byte(dks.privKey)
}

type defaultLogger struct {
	logLevel norddrop.LogLevel
}

func (dl defaultLogger) OnLog(level norddrop.LogLevel, msg string) {
	log.Println(logLevelToPrefix(level), "DROP("+norddrop.Version()+"): "+msg)
}

func (dl defaultLogger) Level() norddrop.LogLevel {
	return dl.logLevel
}

type libdropEventCallback struct {
	eventCallback fileshare.EventCallback
}

func (lec libdropEventCallback) OnEvent(nev norddrop.Event) {
	ev := libdropEventToInternalEvent(nev)
	lec.eventCallback.Event(ev)
}

func libdropEventToInternalEvent(nev norddrop.Event) fileshare.Event {
	return fileshare.Event{
		Kind:      toInternalEventKind(nev.Kind),
		Timestamp: nev.Timestamp,
	}
}

func toInternalEventKind(kind norddrop.EventKind) fileshare.EventKind {
	switch v := kind.(type) {
	case norddrop.EventKindFileDownloaded:
		return fileshare.EventKindFileDownloaded{
			TransferId: v.TransferId,
			FileId:     v.FileId,
			FinalPath:  v.FinalPath,
		}
	case norddrop.EventKindFileFailed:
		return fileshare.EventKindFileFailed{
			TransferId: v.TransferId,
			FileId:     v.FileId,
			Status: fileshare.Status{
				OsErrorCode: v.Status.OsErrorCode,
				Status:      fileshare.StatusCode(v.Status.Status),
			},
		}
	case norddrop.EventKindFileProgress:
		return fileshare.EventKindFileProgress{
			TransferId:  v.TransferId,
			FileId:      v.FileId,
			Transferred: v.Transferred,
		}
	case norddrop.EventKindFileRejected:
		return fileshare.EventKindFileRejected{
			TransferId: v.TransferId,
			FileId:     v.FileId,
			ByPeer:     v.ByPeer,
		}
	case norddrop.EventKindFileStarted:
		return fileshare.EventKindFileStarted{
			TransferId:  v.TransferId,
			FileId:      v.FileId,
			Transferred: v.Transferred,
		}
	case norddrop.EventKindFileUploaded:
		return fileshare.EventKindFileUploaded{
			TransferId: v.TransferId,
			FileId:     v.FileId,
		}
	case norddrop.EventKindRequestQueued:
		return fileshare.EventKindRequestQueued{
			Peer:       v.Peer,
			TransferId: v.TransferId,
			Files:      toInternalQueuedFiles(v.Files),
		}
	case norddrop.EventKindRequestReceived:
		return fileshare.EventKindRequestReceived{
			Peer:       v.Peer,
			TransferId: v.TransferId,
			Files:      toInternalReceivedFiles(v.Files),
		}
	case norddrop.EventKindTransferFailed:
		return fileshare.EventKindTransferFailed{
			Status: fileshare.Status{
				OsErrorCode: v.Status.OsErrorCode,
				Status:      fileshare.StatusCode(v.Status.Status),
			},
			TransferId: "",
		}
	case norddrop.EventKindTransferFinalized:
		return fileshare.EventKindTransferFinalized{
			TransferId: v.TransferId,
			ByPeer:     v.ByPeer,
		}
	default:
		log.Printf(internal.WarningPrefix+" unexpected norddrop.EventKind: %T\n", v)
		return fileshare.EventKindUnknown{}
	}
}

func toInternalQueuedFiles(files []norddrop.QueuedFile) []fileshare.QueuedFile {
	result := make([]fileshare.QueuedFile, len(files))
	for i, file := range files {
		result[i] = fileshare.QueuedFile{
			BaseDir: file.BaseDir,
			Id:      file.Id,
			Path:    file.Path,
			Size:    file.Size,
		}
	}
	return result
}

func toInternalReceivedFiles(files []norddrop.ReceivedFile) []fileshare.ReceivedFile {
	result := make([]fileshare.ReceivedFile, len(files))
	for i, file := range files {
		result[i] = fileshare.ReceivedFile{
			Id:   file.Id,
			Path: file.Path,
			Size: file.Size,
		}
	}
	return result
}

// New initializes norddrop library.
func New(
	eventCb fileshare.EventCallback,
	eventsDbPath string,
	isProd bool,
	pubkeyFunc func(string) []byte,
	privKey string,
	storagePath string,
) (*Fileshare, error) {
	keyStore := defaultKeyStore{
		pubkeyFunc: pubkeyFunc,
		privKey:    privKey,
	}
	logLevel := norddrop.LogLevelTrace
	if isProd {
		logLevel = norddrop.LogLevelError
	}

	logger := defaultLogger{logLevel}

	eventCallback := libdropEventCallback{eventCb}
	norddrop, err := norddrop.NewNordDrop(eventCallback, keyStore, logger)
	if err != nil {
		return nil, fmt.Errorf("creating norddrop instance: %w", err)
	}

	return &Fileshare{
		norddrop:     norddrop,
		eventsDbPath: eventsDbPath,
		storagePath:  storagePath,
		isProd:       isProd,
	}, nil
}

// Enable executes Start in norddrop library. Has to be called before using other Fileshare methods.
func (f *Fileshare) Enable(listenAddr netip.Addr) (err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	log.Println(internal.InfoPrefix, "libdrop version:", norddrop.Version())

	if err = f.start(listenAddr, f.eventsDbPath, f.isProd, f.storagePath); err != nil {
		if errors.Is(err, norddrop.ErrLibdropErrorAddrInUse) {
			return fileshare.ErrAddressAlreadyInUse
		}
		return fmt.Errorf("starting drop: %w", err)
	}
	return nil
}

func (f *Fileshare) start(
	listenAddr netip.Addr,
	eventsDbPath string,
	isProd bool,
	storagePath string,
) error {
	var autoRetryIntervalMs uint32 = 5000
	config := norddrop.Config{
		DirDepthLimit:       fileshare.DirDepthLimit,
		TransferFileLimit:   fileshare.TransferFileLimit,
		MooseEventPath:      eventsDbPath,
		MooseProd:           isProd,
		StoragePath:         storagePath,
		AutoRetryIntervalMs: &autoRetryIntervalMs,
	}

	return f.norddrop.Start(listenAddr.String(), config)
}

// Disable executes Stop in norddrop library. Other Fileshare methods can't be called until
// after Enable is called again.
func (f *Fileshare) Disable() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if err := f.stop(); err != nil {
		return fmt.Errorf("stopping drop: %w", err)
	}

	return nil
}

func (f *Fileshare) stop() error {
	return f.norddrop.Stop()
}

// Send file or dir to peer.
// Path must be absolute.
// Returns transfer ID.
func (f *Fileshare) Send(peer netip.Addr, paths []string) (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if !peer.Is4() {
		return "", fmt.Errorf("peer %s must be an IPv4 address", peer.String())
	}

	transferDescriptors := make([]norddrop.TransferDescriptor, len(paths))

	for i, path := range paths {
		transferDescriptors[i] = norddrop.TransferDescriptorPath{Path: path}
	}

	transfer, err := f.norddrop.NewTransfer(peer.String(), transferDescriptors)
	if err != nil {
		return "", fmt.Errorf("transfer wasn't created")
	}

	return transfer, nil
}

// Accept starts downloading provided files into dstPath.
// dstPath must be absolute.
func (f *Fileshare) Accept(transferID, dstPath string, fileID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if err := f.norddrop.DownloadFile(transferID, fileID, dstPath); err != nil {
		return err
	}

	return nil
}

// Finalize file transfer.
func (f *Fileshare) Finalize(transferID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.norddrop.FinalizeTransfer(transferID)
}

// CancelFile id in a transfer
func (f *Fileshare) CancelFile(transferID string, fileID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.norddrop.RejectFile(transferID, fileID)
}

// Load transfers from fileshare implementation storage
func (f *Fileshare) Load() (map[string]*pb.Transfer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	since := time.Time{}.Unix()
	norddropTransfers, err := f.norddrop.TransfersSince(since)
	if err != nil {
		return nil, fmt.Errorf("getting transfers since %d: %w", since, err)
	}

	transfers := make(map[string]*pb.Transfer)
	for _, transfer := range norddropTransfers {
		transfers[transfer.Id] = norddropTransferToPBTransfer(transfer)
	}

	return transfers, nil
}

// PurgeTransfersUntil provided time from fileshare implementation storage
func (f *Fileshare) PurgeTransfersUntil(until time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.norddrop.PurgeTransfersUntil(until.Unix() * 1000)
}

func norddropTransferToPBTransfer(ti norddrop.TransferInfo) *pb.Transfer {
	allFiles := filesFromTransferInfo(&ti)
	totalSize, totalTransferred := calculateTotalSizeAndTotalTransferred(allFiles)
	out := &pb.Transfer{
		Id:               ti.Id,
		Direction:        directionFromTransferInfo(&ti),
		Peer:             ti.Peer,
		Created:          timestamppb.New(time.UnixMilli(ti.CreatedAt)),
		Files:            allFiles,
		TotalSize:        totalSize,
		TotalTransferred: totalTransferred,
		Status:           determineTransferStatusAndAdjustFileStatuses(&ti, allFiles),
		Path:             determineTransferPath(&ti, allFiles),
	}

	return out
}

func directionFromTransferInfo(ti *norddrop.TransferInfo) pb.Direction {
	var direction pb.Direction
	switch ti.Kind.(type) {
	case norddrop.TransferKindOutgoing:
		direction = pb.Direction_OUTGOING
	case norddrop.TransferKindIncoming:
		direction = pb.Direction_INCOMING
	default:
		log.Printf(internal.WarningPrefix+" unknown direction found when parsing libdrop transfers: %T\n", ti.Kind)
		direction = pb.Direction_UNKNOWN_DIRECTION
	}
	return direction
}

func filesFromTransferInfo(ri *norddrop.TransferInfo) []*pb.File {
	switch ti := ri.Kind.(type) {
	case norddrop.TransferKindIncoming:
		files := make([]*pb.File, len(ti.Paths))
		for i, path := range ti.Paths {
			files[i] = norddropIncomingPathToPBFile(path)
		}
		return files
	case norddrop.TransferKindOutgoing:
		files := make([]*pb.File, len(ti.Paths))
		for i, path := range ti.Paths {
			files[i] = norddropOutgoingPathToPBFile(path)
		}
		return files
	default:
		log.Printf(internal.WarningPrefix+" unknown transfer kind: %T\n", ti)
		return []*pb.File{}
	}
}

func norddropOutgoingPathToPBFile(outPath norddrop.OutgoingPath) *pb.File {
	file := &pb.File{
		Id:     outPath.FileId,
		Path:   outPath.RelativePath,
		Size:   outPath.Bytes,
		Status: statusFromOutgoingPath(&outPath),
		// This only shows the amount from the last status change.
		// The only up to date source for data transferred are events.
		// To show the correct value we track ongoing transfers in event manager,
		// and event manager overwrites values of transferred fields if needed.
		Transferred: transferredFromOutgoingPath(&outPath),
		FullPath:    determineFullOutgoingPath(&outPath),
	}
	return file
}

func statusFromOutgoingPath(outPath *norddrop.OutgoingPath) pb.Status {
	var status pb.Status

	if len(outPath.States) == 0 {
		status = pb.Status_REQUESTED
	} else {
		lastState := outPath.States[len(outPath.States)-1]
		switch lastState := lastState.Kind.(type) {
		case norddrop.OutgoingPathStateKindStarted:
			status = pb.Status_ONGOING
		case norddrop.OutgoingPathStateKindFailed:
			status = pb.Status(lastState.Status)
		case norddrop.OutgoingPathStateKindCompleted:
			status = pb.Status_SUCCESS
		case norddrop.OutgoingPathStateKindRejected:
			status = pb.Status_CANCELED
		case norddrop.OutgoingPathStateKindPaused:
			status = pb.Status_PAUSED
		default:
			log.Printf(internal.WarningPrefix+" unknown file status in transfer: %T\n", lastState)
			status = pb.Status_BAD_STATUS
		}
	}
	return status
}

func transferredFromOutgoingPath(outPath *norddrop.OutgoingPath) uint64 {
	var transferred uint64
	for _, state := range outPath.States {
		switch st := state.Kind.(type) {
		case norddrop.OutgoingPathStateKindStarted:
			transferred = st.BytesSent
		case norddrop.OutgoingPathStateKindCompleted:
			transferred = outPath.Bytes
		case norddrop.OutgoingPathStateKindFailed:
			transferred = st.BytesSent
		case norddrop.OutgoingPathStateKindRejected:
			transferred = st.BytesSent
		case norddrop.OutgoingPathStateKindPaused:
			transferred = st.BytesSent
		}
	}
	return transferred
}

func determineFullOutgoingPath(outPath *norddrop.OutgoingPath) string {
	switch pathSource := outPath.Source.(type) {
	case norddrop.OutgoingFileSourceBasePath:
		return filepath.Join(pathSource.BasePath, outPath.RelativePath)
	default:
		log.Printf(internal.WarningPrefix+" unsupported path source: %T\n", outPath.Source)
		return ""
	}
}

func norddropIncomingPathToPBFile(inPath norddrop.IncomingPath) *pb.File {
	out := &pb.File{
		Id:     inPath.FileId,
		Path:   inPath.RelativePath,
		Size:   inPath.Bytes,
		Status: statusFromIncomingPath(&inPath),
		// This only shows the amount from the last status change.
		// The only up to date source for data transferred are events.
		// To show the correct value we track ongoing transfers in event manager,
		// and event manager overwrites values of transferred fields if needed.
		Transferred: transferredFromIncomingPath(&inPath),
		FullPath:    determineFullIncomingPath(&inPath),
	}
	return out
}

func statusFromIncomingPath(inPath *norddrop.IncomingPath) pb.Status {
	var status pb.Status
	if len(inPath.States) == 0 {
		status = pb.Status_REQUESTED
	} else {
		lastState := inPath.States[len(inPath.States)-1]
		switch lastState := lastState.Kind.(type) {
		case norddrop.IncomingPathStateKindCompleted:
			status = pb.Status_SUCCESS
		case norddrop.IncomingPathStateKindFailed:
			status = pb.Status(lastState.Status)
		case norddrop.IncomingPathStateKindPaused:
			status = pb.Status_PAUSED
		case norddrop.IncomingPathStateKindPending:
			status = pb.Status_PENDING
		case norddrop.IncomingPathStateKindRejected:
			status = pb.Status_CANCELED
		case norddrop.IncomingPathStateKindStarted:
			status = pb.Status_ONGOING
		default:
			log.Printf(internal.WarningPrefix+" unknown file status in transfer: %T\n", lastState)
			status = pb.Status_BAD_STATUS
		}
	}
	return status
}

func transferredFromIncomingPath(inPath *norddrop.IncomingPath) uint64 {
	var transferred uint64
	for _, state := range inPath.States {
		switch st := state.Kind.(type) {
		case norddrop.IncomingPathStateKindStarted:
			transferred = st.BytesReceived
		case norddrop.IncomingPathStateKindCompleted:
			transferred = inPath.Bytes
		case norddrop.IncomingPathStateKindFailed:
			transferred = st.BytesReceived
		case norddrop.IncomingPathStateKindRejected:
			transferred = st.BytesReceived
		case norddrop.IncomingPathStateKindPaused:
			transferred = st.BytesReceived
		}
	}
	return transferred
}

func determineFullIncomingPath(inPath *norddrop.IncomingPath) string {
	fullPath := inPath.RelativePath
	for _, state := range inPath.States {
		switch st := state.Kind.(type) {
		case norddrop.IncomingPathStateKindPending:
			// BasePath is provided with the very first "pending" state
			fullPath = filepath.Join(st.BaseDir, fullPath)
		case norddrop.IncomingPathStateKindCompleted:
			// FinalPath is provided with the very last "completed" state, so if it is present, then
			// it will overwrite the previously constructed path
			fullPath = st.FinalPath
		}
	}
	return fullPath
}

func calculateTotalSizeAndTotalTransferred(allFiles []*pb.File) (uint64, uint64) {
	var totalSize, totalTransferred uint64
	for _, file := range allFiles {
		if isFileTransferred(file) {
			totalSize += file.Size
			totalTransferred += file.Transferred
		}
	}
	return totalSize, totalTransferred
}

func determineTransferStatusAndAdjustFileStatuses(libdropTransfer *norddrop.TransferInfo, allFiles []*pb.File) pb.Status {
	var status pb.Status
	if len(libdropTransfer.States) != 0 {
		lastState := libdropTransfer.States[len(libdropTransfer.States)-1]
		switch lastState := lastState.Kind.(type) {
		case norddrop.TransferStateKindCancel:
			// This is annoying. We have to "finalize" finished transfers by cancelling them,
			// otherwise there's a resource leak in libdrop. Also, by doing that all finished
			// transfers have cancel status, so we need to figure out the real status.
			// So we try to determine status by files and see if the transfer was already finished
			// when it was cancelled or not.
			statusByFiles := getTransferStatusByFiles(allFiles)
			if statusByFiles == pb.Status_SUCCESS || statusByFiles == pb.Status_FINISHED_WITH_ERRORS {
				status = statusByFiles
			} else {
				if lastState.ByPeer {
					status = pb.Status_CANCELED_BY_PEER
				} else {
					status = pb.Status_CANCELED
				}
				// adjust file statuses
				for _, file := range allFiles {
					if !isFileCompleted(file) {
						file.Status = pb.Status_CANCELED
					}
				}
			}
		case norddrop.TransferStateKindFailed:
			status = pb.Status(lastState.Status)
		}
	} else {
		status = getTransferStatusByFiles(allFiles)
	}
	return status
}

func getTransferStatusByFiles(files []*pb.File) pb.Status {
	allCanceled := true
	allFinished := true
	hasNoErrors := true
	hasStarted := false
	for _, file := range files {
		if allCanceled && file.Status != pb.Status_CANCELED {
			allCanceled = false
		}
		if allFinished && !isFileCompleted(file) {
			allFinished = false
		}
		if hasNoErrors && checkFileHasErrors(file) {
			hasNoErrors = false
		}
		if !hasStarted && file.Status != pb.Status_REQUESTED {
			hasStarted = true
		}
	}

	switch {
	case allCanceled:
		return pb.Status_CANCELED
	case allFinished && hasNoErrors:
		return pb.Status_SUCCESS
	case allFinished && !hasNoErrors:
		return pb.Status_FINISHED_WITH_ERRORS
	case hasStarted:
		return pb.Status_ONGOING
	default:
		return pb.Status_REQUESTED
	}
}

func determineTransferPath(libdropTransfer *norddrop.TransferInfo, allFiles []*pb.File) string {
	var result string
	var transferPath string
	// Base path is in different place for incoming and outgoing files
	switch ti := libdropTransfer.Kind.(type) {
	case norddrop.TransferKindIncoming:
		for _, path := range ti.Paths {
			for _, state := range path.States {
				switch st := state.Kind.(type) {
				case norddrop.IncomingPathStateKindPending:
					transferPath = st.BaseDir
				}
			}

			if result == "" {
				result = transferPath
			} else if result != transferPath && transferPath != "" {
				result = "multiple files"
			}
		}
		return transferPath
	case norddrop.TransferKindOutgoing:
		for i, path := range ti.Paths {
			if filepath.Base(path.RelativePath) == path.RelativePath {
				file := allFiles[i]
				// If relative path doesn't contain a directory - means user specified a single file
				// We show the full path in that case for outgoing transfers
				transferPath = file.FullPath
			} else {
				// If user is sending a directory - we add the directory itself to the path.
				// Example: user sends /tmp/test, then file.BasePath would be "/tmp", while
				// file.RelativePath would be "test/file". So we take the dir name from RelativePath
				// and add it to BasePath to get "/tmp/test".
				dir, _, ok := strings.Cut(path.RelativePath, string(filepath.Separator))
				if ok {
					switch source := path.Source.(type) {
					case norddrop.OutgoingFileSourceBasePath:
						transferPath = filepath.Join(source.BasePath, dir)
					default:
						log.Printf(internal.WarningPrefix+" unsupported path source: %T\n", source)
					}
				}
			}

			if result == "" {
				result = transferPath
			} else if result != transferPath && transferPath != "" {
				result = "multiple files"
			}
		}
	}

	return result
}

func checkFileHasErrors(file *pb.File) bool {
	return file.Status != pb.Status_SUCCESS &&
		file.Status != pb.Status_REQUESTED &&
		file.Status != pb.Status_CANCELED &&
		file.Status != pb.Status_ONGOING &&
		file.Status != pb.Status_PENDING
}

func isFileCompleted(file *pb.File) bool {
	return file.Status != pb.Status_REQUESTED &&
		file.Status != pb.Status_ONGOING &&
		file.Status != pb.Status_PENDING
}

// Used to check if file's size should be part of transfer's total size
// Basically we don't include files that are canceled or errored out
func isFileTransferred(file *pb.File) bool {
	return !isFileCompleted(file) || file.Status == pb.Status_SUCCESS
}
