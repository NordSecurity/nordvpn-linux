// Package libdrop wraps libdrop fileshare implementation.
package libdrop

import (
	"fmt"
	"log"
	"net/netip"
	"sync"
	"time"

	norddrop "github.com/NordSecurity/libdrop-go/v7"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

func (nec libdropEventCallback) OnEvent(nev norddrop.Event) {
	ev := libdropEventToInternalEvent(nev)
	nec.eventCallback.OnEvent(ev)
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
	config := norddrop.Config{
		DirDepthLimit:     fileshare.DirDepthLimit,
		TransferFileLimit: fileshare.TransferFileLimit,
		MooseEventPath:    eventsDbPath,
		MooseProd:         isProd,
		StoragePath:       storagePath,
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

	if err := f.norddrop.RejectFile(transferID, fileID); err != nil {
		return err
	}

	return nil
}

// GetTransfersSince provided time from fileshare implementation storage
func (f *Fileshare) GetTransfersSince(t time.Time) ([]fileshare.LibdropTransfer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	since := t.Unix()
	norddropTransfers, err := f.norddrop.TransfersSince(since)
	if err != nil {
		return nil, fmt.Errorf("getting transfers since %d: %w", since, err)
	}

	transfers := make([]fileshare.LibdropTransfer, len(norddropTransfers))

	for i, transfer := range norddropTransfers {
		transfers[i] = toInternalTransfer(&transfer)
	}

	return transfers, nil
}

// PurgeTransfersUntil provided time from fileshare implementation storage
func (f *Fileshare) PurgeTransfersUntil(until time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// TODO: In the calculation below: `until.Unix() * 100` it should be
	// multiplied by 1000 to get number of milliseconds. The issue is that there
	// is a bug on the libdrop side here: https://github.com/NordSecurity/libdrop/blob/v7.0.0/norddrop/src/uni.rs#L100
	// It converts milliseconds to seconds by dividing by 100 instead of 1000
	// resulting in incorrect dates in the year ~2515 and purging of all transfers.
	// This will be fixed with migration to v8.0.0 of libdrop.
	return f.norddrop.PurgeTransfersUntil(until.Unix() * 100)
}

func toInternalTransfer(ti *norddrop.TransferInfo) fileshare.LibdropTransfer {
	return fileshare.LibdropTransfer{
		Kind:      toInternalTransferKind(ti.Kind),
		Id:        ti.Id,
		Peer:      ti.Peer,
		States:    toInternalTransferStates(ti.States),
		CreatedAt: ti.CreatedAt,
	}
}

func toInternalTransferKind(transferKind norddrop.TransferKind) fileshare.TransferKind {
	switch v := transferKind.(type) {
	case norddrop.TransferKindIncoming:
		return fileshare.TransferKindIncoming{
			Paths: toInternalIncomingPaths(v.Paths),
		}
	case norddrop.TransferKindOutgoing:
		return fileshare.TransferKindOutgoing{
			Paths: toInternalOutgoingPaths(v.Paths),
		}
	default:
		log.Printf(internal.WarningPrefix+" unexpected norddrop.TransferKind: %T\n", v)
		return fileshare.TransferKindUnknown{}
	}
}

func toInternalIncomingPaths(paths []norddrop.IncomingPath) []fileshare.IncomingPath {
	result := make([]fileshare.IncomingPath, len(paths))
	for i, path := range paths {
		result[i] = fileshare.IncomingPath{
			FileId:        path.FileId,
			RelativePath:  path.RelativePath,
			Bytes:         path.Bytes,
			BytesReceived: path.BytesReceived,
			States:        toInternalIncomingPathStates(path.States),
		}
	}
	return result
}

func toInternalIncomingPathStates(states []norddrop.IncomingPathState) []fileshare.IncomingPathState {
	result := make([]fileshare.IncomingPathState, len(states))
	for i, state := range states {
		result[i] = fileshare.IncomingPathState{
			Kind:      toInternalIncomingPathStateKind(state.Kind),
			CreatedAt: state.CreatedAt,
		}
	}
	return result
}

func toInternalIncomingPathStateKind(kind norddrop.IncomingPathStateKind) fileshare.IncomingPathStateKind {
	switch v := kind.(type) {
	case norddrop.IncomingPathStateKindCompleted:
		return fileshare.IncomingPathStateKindCompleted{
			FinalPath: v.FinalPath,
		}
	case norddrop.IncomingPathStateKindFailed:
		return fileshare.IncomingPathStateKindFailed{
			Status:        fileshare.StatusCode(v.Status),
			BytesReceived: v.BytesReceived,
		}
	case norddrop.IncomingPathStateKindPaused:
		return fileshare.IncomingPathStateKindPaused{
			BytesReceived: v.BytesReceived,
		}
	case norddrop.IncomingPathStateKindPending:
		return fileshare.IncomingPathStateKindPending{
			BaseDir: v.BaseDir,
		}
	case norddrop.IncomingPathStateKindRejected:
		return fileshare.IncomingPathStateKindRejected{
			ByPeer:        v.ByPeer,
			BytesReceived: v.BytesReceived,
		}
	case norddrop.IncomingPathStateKindStarted:
		return fileshare.IncomingPathStateKindStarted{
			BytesReceived: v.BytesReceived,
		}
	default:
		log.Printf(internal.WarningPrefix+" unexpected norddrop.IncomingPathStateKind: %T\n", v)
		return fileshare.IncomingPathStateKindUnknown{}
	}
}

func toInternalTransferStates(transferStates []norddrop.TransferState) []fileshare.TransferState {
	states := make([]fileshare.TransferState, len(transferStates))
	for i, state := range transferStates {
		states[i] = fileshare.TransferState{
			Kind:      toInternalTransferStateKind(state.Kind),
			CreatedAt: state.CreatedAt,
		}
	}
	return states
}

func toInternalTransferStateKind(kind norddrop.TransferStateKind) fileshare.TransferStateKind {
  switch v := kind.(type) {
	case norddrop.TransferStateKindCancel:
  return fileshare.TransferStateKindCancel{
  	ByPeer: v.ByPeer,
  }
	case norddrop.TransferStateKindFailed:
  return fileshare.TransferStateKindFailed{
  	Status: fileshare.StatusCode(v.Status),
  }
	default:
    log.Printf(internal.WarningPrefix+" unexpected norddrop.TransferStateKind: %T\n", v)
  return fileshare.TransferStateKindUnknown{}
	}
}

func toInternalOutgoingPaths(paths []norddrop.OutgoingPath) []fileshare.OutgoingPath {
	result := make([]fileshare.OutgoingPath, len(paths))
	for i, path := range paths {
		result[i] = fileshare.OutgoingPath{
			Source:       toInternalPathSource(path.Source),
			FileId:       path.FileId,
			RelativePath: path.RelativePath,
			States:       toInternalOutgoingPathStates(path.States),
			Bytes:        path.Bytes,
			BytesSent:    path.BytesSent,
		}
	}
	return result
}

func toInternalPathSource(source norddrop.OutgoingFileSource) fileshare.OutgoingFileSource {
	switch v := source.(type) {
	case norddrop.OutgoingFileSourceBasePath:
		return fileshare.OutgoingFileSourceBasePath{
			BasePath: v.BasePath,
		}
	default:
		log.Printf(internal.WarningPrefix+" unexpected norddrop.OutgoingFileSource: %T\n", v)
		return fileshare.OutgoingFileSourceUnknown{}
	}
}

func toInternalOutgoingPathStates(states []norddrop.OutgoingPathState) []fileshare.OutgoingPathState {
	result := make([]fileshare.OutgoingPathState, len(states))
	for i, state := range states {
		result[i] = fileshare.OutgoingPathState{
			Kind:      toInternalPathStateKind(state.Kind),
			CreatedAt: state.CreatedAt,
		}
	}
	return result
}

func toInternalPathStateKind(kind norddrop.OutgoingPathStateKind) fileshare.OutgoingPathStateKind {
	switch v := kind.(type) {
	case norddrop.OutgoingPathStateKindCompleted:
		return fileshare.OutgoingPathStateKindCompleted{}
	case norddrop.OutgoingPathStateKindFailed:
		return fileshare.OutgoingPathStateKindFailed{
			Status:    fileshare.StatusCode(v.Status),
			BytesSent: v.BytesSent,
		}
	case norddrop.OutgoingPathStateKindPaused:
		return fileshare.OutgoingPathStateKindPaused{
			BytesSent: v.BytesSent,
		}
	case norddrop.OutgoingPathStateKindRejected:
		return fileshare.OutgoingPathStateKindRejected{
			ByPeer:    v.ByPeer,
			BytesSent: v.BytesSent,
		}
	case norddrop.OutgoingPathStateKindStarted:
		return fileshare.OutgoingPathStateKindStarted{
			BytesSent: v.BytesSent,
		}
	default:
		log.Printf(internal.WarningPrefix+" unexpected norddrop.OutgoingPathStateKind: %T\n", v)
		return fileshare.OutgoingPathStateKindUnknown{}
	}
}
