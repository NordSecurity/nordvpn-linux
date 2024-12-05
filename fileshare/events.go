package fileshare

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type EventCallback interface {
	AsyncEvent(event ...Event)
	SyncEvent(event ...Event)
}

type Event struct {
	Kind      EventKind
	Timestamp int64
}

type EventKind interface{}

type EventKindRequestReceived struct {
	Peer       string
	TransferId string
	Files      []ReceivedFile
}

type ReceivedFile struct {
	Id   string
	Path string
	Size uint64
}

type EventKindRequestQueued struct {
	Peer       string
	TransferId string
	Files      []QueuedFile
}

type QueuedFile struct {
	BaseDir *string
	Id      string
	Path    string
	Size    uint64
}

type EventKindFileStarted struct {
	TransferId  string
	FileId      string
	Transferred uint64
}

type EventKindFileProgress struct {
	TransferId  string
	FileId      string
	Transferred uint64
}

type EventKindTransferFailed struct {
	Status     Status
	TransferId string
}

type Status struct {
	OsErrorCode *int32
	Status      StatusCode
}

type EventKindTransferFinalized struct {
	TransferId string
	ByPeer     bool
}

type EventKindFileDownloaded struct {
	TransferId string
	FileId     string
	FinalPath  string
}

type EventKindFileUploaded struct {
	TransferId string
	FileId     string
}

type EventKindFileRejected struct {
	TransferId string
	FileId     string
	ByPeer     bool
}

type EventKindFileFailed struct {
	TransferId string
	FileId     string
	Status     Status
}

type EventKindUnknown struct{}

type StatusCode uint

const (
	// Not an error per se; indicates finalized transfers.
	StatusCodeFinalized StatusCode = 1
	// An invalid path was provided.
	// File path contains invalid components (e.g. parent `..`).
	StatusCodeBadPath StatusCode = 2
	// Failed to open the file or file doesn’t exist when asked to download. Might
	// indicate bad API usage. For Unix platforms using file descriptors, it might
	// indicate invalid FD being passed to libdrop.
	StatusCodeBadFile StatusCode = 3
	// Invalid input transfer ID passed.
	StatusCodeBadTransfer StatusCode = 4
	// An error occurred during the transfer and it cannot continue. The most probable
	// reason is the error occurred on the peer’s device or other error that cannot be
	// categorize elsewhere.
	StatusCodeBadTransferState StatusCode = 5
	// Invalid input file ID passed when.
	StatusCodeBadFileId StatusCode = 6
	// General IO error. Check the logs and contact libdrop team.
	StatusCodeIoError StatusCode = 7
	// Transfer limits exceeded. Limit is in terms of depth and breadth for
	// directories.
	StatusCodeTransferLimitsExceeded StatusCode = 8
	// The file size has changed since adding it to the transfer. The original file was
	// modified while not in flight in such a way that its size changed.
	StatusCodeMismatchedSize StatusCode = 9
	// An invalid argument was provided either as a function argument or
	// invalid config value.
	StatusCodeInvalidArgument StatusCode = 10
	// The WebSocket server failed to bind because of an address collision.
	StatusCodeAddrInUse StatusCode = 11
	// The file was modified while being uploaded.
	StatusCodeFileModified StatusCode = 12
	// The filename is too long which might be due to the fact the sender uses
	// a filesystem supporting longer filenames than the one which’s downloading the
	// file.
	StatusCodeFilenameTooLong StatusCode = 13
	// A peer couldn’t validate our authentication request.
	StatusCodeAuthenticationFailed StatusCode = 14
	// Persistence error.
	StatusCodeStorageError StatusCode = 15
	// The persistence database is lost. A new database will be created.
	StatusCodeDbLost StatusCode = 16
	// Downloaded file checksum differs from the advertised one. The downloaded
	// file is deleted by libdrop.
	StatusCodeFileChecksumMismatch StatusCode = 17
	// Download is impossible of the rejected file.
	StatusCodeFileRejected StatusCode = 18
	// Action is blocked because the failed condition has been reached.
	StatusCodeFileFailed StatusCode = 19
	// Action is blocked because the file is already transferred.
	StatusCodeFileFinished StatusCode = 20
	// Transfer requested with empty file list.
	StatusCodeEmptyTransfer StatusCode = 21
	// Transfer resume attempt was closed by peer for no reason. It might indicate
	// temporary issues on the peer’s side. It is safe to continue to resume the
	// transfer.
	StatusCodeConnectionClosedByPeer StatusCode = 22
	// Peer’s DDoS protection kicked in.
	// Transfer should be resumed after some cooldown period.
	StatusCodeTooManyRequests StatusCode = 23
	// This error code is intercepted from the OS errors. Indicate lack of
	// privileges to do certain operation.
	StatusCodePermissionDenied StatusCode = 24
)

var defaultMarshaler = jsonMarshaler{}

func EventToString(event Event) string {
	return eventToString(event, defaultMarshaler)
}

func eventToString(event Event, m marshaler) string {
	json, err := m.Marshal(event)
	if err != nil {
		log.Printf(internal.WarningPrefix+" failed to marshall event: %T, returning just its type\n", event.Kind)
		return fmt.Sprintf("%T", event.Kind)
	}
	return string(json)
}

type marshaler interface {
	Marshal(v any) ([]byte, error)
}

type jsonMarshaler struct{}

func (jm jsonMarshaler) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}
