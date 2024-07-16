package fileshare

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	IncomingStatus = map[pb.Status]string{
		pb.Status_REQUESTED:            "waiting for download",
		pb.Status_ONGOING:              "downloading",
		pb.Status_SUCCESS:              "completed",
		pb.Status_INTERRUPTED:          "interrupted",
		pb.Status_FINISHED_WITH_ERRORS: "completed with errors",
		pb.Status_ACCEPT_FAILURE:       "accepted with errors",
		pb.Status_CANCELED:             "canceled",
		pb.Status_CANCELED_BY_PEER:     "canceled by peer",
		pb.Status_PENDING:              "pending",
	}
	OutgoingStatus = map[pb.Status]string{
		pb.Status_REQUESTED:            "request sent",
		pb.Status_ONGOING:              "uploading",
		pb.Status_SUCCESS:              "completed",
		pb.Status_INTERRUPTED:          "interrupted",
		pb.Status_FINISHED_WITH_ERRORS: "completed with errors",
		pb.Status_ACCEPT_FAILURE:       "accepted with errors",
		pb.Status_CANCELED:             "canceled",
		pb.Status_CANCELED_BY_PEER:     "canceled by peer",
		pb.Status_FILE_REJECTED:        "the receiver has declined the file transfer",
		pb.Status_PENDING:              "pending",
	}
	FileStatus = map[pb.Status]string{
		pb.Status_SUCCESS:                  "completed",
		pb.Status_CANCELED:                 "canceled",
		pb.Status_INTERRUPTED:              "interrupted",
		pb.Status_PENDING:                  "pending",
		pb.Status_BAD_PATH:                 "the download path is not valid",
		pb.Status_BAD_FILE:                 "the file is no longer found",
		pb.Status_TRANSPORT:                "transport problem",
		pb.Status_BAD_STATUS:               "bad status",
		pb.Status_SERVICE_STOP:             "service not active",
		pb.Status_BAD_TRANSFER:             "bad transfer",
		pb.Status_BAD_TRANSFER_STATE:       "bad transfer state",
		pb.Status_BAD_FILE_ID:              "bad file id",
		pb.Status_BAD_SYSTEM_TIME:          "bad system time",
		pb.Status_TRUNCATED_FILE:           "truncated file",
		pb.Status_EVENT_SEND:               "internal error",
		pb.Status_BAD_UUID:                 "internal error",
		pb.Status_CHANNEL_CLOSED:           "internal error",
		pb.Status_IO:                       "io error",
		pb.Status_DATA_SEND:                "data send error",
		pb.Status_DIRECTORY_NOT_EXPECTED:   "directory not expected",
		pb.Status_EMPTY_TRANSFER:           "empty transfer",
		pb.Status_TRANSFER_CLOSED_BY_PEER:  "transfer closed by peer",
		pb.Status_TRANSFER_LIMITS_EXCEEDED: "directory depth or breadth limits are exceeded",
		pb.Status_MISMATCHED_SIZE:          "the file has been modified",
		pb.Status_UNEXPECTED_DATA:          "the file has been modified",
		pb.Status_INVALID_ARGUMENT:         "internal error",
		pb.Status_TRANSFER_TIMEOUT:         "transfer timeout",
		pb.Status_WS_SERVER:                "waiting for peer to come online",
		pb.Status_WS_CLIENT:                "waiting for peer to come online",
		pb.Status_FILE_MODIFIED:            "the file has been modified",
		pb.Status_FILENAME_TOO_LONG:        "filename too long",
		pb.Status_AUTHENTICATION_FAILED:    "authentication failed",
		pb.Status_FILE_CHECKSUM_MISMATCH:   "the file is corrupted",
	}
	IncomingFileStatus = map[pb.Status]string{
		pb.Status_SUCCESS:              "downloaded",
		pb.Status_CANCELED:             "canceled",
		pb.Status_REQUESTED:            "waiting for download",
		pb.Status_ONGOING:              "downloading",
		pb.Status_FINISHED_WITH_ERRORS: "downloaded with errors",
		pb.Status_ACCEPT_FAILURE:       "accepted with errors",
	}
	OutgoingFileStatus = map[pb.Status]string{
		pb.Status_SUCCESS:              "uploaded",
		pb.Status_CANCELED:             "canceled",
		pb.Status_REQUESTED:            "request sent",
		pb.Status_ONGOING:              "uploading",
		pb.Status_FINISHED_WITH_ERRORS: "uploaded with errors",
		pb.Status_ACCEPT_FAILURE:       "accepted with errors",
	}
)

type LibdropTransfer struct {
	Kind      TransferKind
	Id        string
	Peer      string
	States    []TransferState
	CreatedAt int64
}

type TransferKind interface{}

type TransferKindOutgoing struct {
	Paths []OutgoingPath
}

type OutgoingPath struct {
	Source       OutgoingFileSource
	FileId       string
	RelativePath string
	States       []OutgoingPathState
	Bytes        uint64
	BytesSent    uint64
}

type OutgoingFileSource interface{}

type OutgoingFileSourceBasePath struct {
	BasePath string
}

type OutgoingFileSourceUnknown struct{}

type OutgoingPathState struct {
	Kind      OutgoingPathStateKind
	CreatedAt int64
}

type OutgoingPathStateKind interface{}

type TransferState struct {
	Kind      TransferStateKind
	CreatedAt int64
}

type TransferStateKind interface{}

type TransferKindIncoming struct {
	Paths []IncomingPath
}

type IncomingPath struct {
	FileId        string
	RelativePath  string
	States        []IncomingPathState
	Bytes         uint64
	BytesReceived uint64
}

type IncomingPathState struct {
	Kind      IncomingPathStateKind
	CreatedAt int64
}

type IncomingPathStateKind interface{}

type IncomingPathStateKindPending struct {
	BaseDir string
}

type IncomingPathStateKindStarted struct {
	BytesReceived uint64
}

type IncomingPathStateKindRejected struct {
	ByPeer        bool
	BytesReceived uint64
}

type IncomingPathStateKindFailed struct {
	Status        StatusCode
	BytesReceived uint64
}

type IncomingPathStateKindUnknown struct{}

type TransferKindUnknown struct{}

type OutgoingPathStateKindStarted struct {
	BytesSent uint64
}

type OutgoingPathStateKindUnknown struct {
	BytesSent uint64
}

type OutgoingPathStateKindFailed struct {
	Status    StatusCode
	BytesSent uint64
}

type OutgoingPathStateKindPaused struct {
	BytesSent uint64
}

type TransferStateKindCancel struct {
	ByPeer bool
}

type TransferStateKindFailed struct {
	Status StatusCode
}

type TransferStateKindUnknown struct{}

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

type LibdropFile struct {
	ID           string
	TransferID   string
	BasePath     string
	RelativePath string
	States       []LibdropFileState
	TotalSize    uint64
	CreatedAt    uint64
}

type LibdropFileState struct {
	State         string
	BasePath      string
	FinalPath     string
	CreatedAt     uint64
	BytesSent     uint64
	BytesReceived uint64
	StatusCode    int
}

type OutgoingPathStateKindCompleted struct{}

type OutgoingPathStateKindRejected struct {
	ByPeer    bool
	BytesSent uint64
}

type IncomingPathStateKindPaused struct {
	BytesReceived uint64
}

type IncomingPathStateKindCompleted struct {
	FinalPath string
}

// GetTransferStatus translate transfer status into human readable form
func GetTransferStatus(tr *pb.Transfer) string {
	if tr.Direction == pb.Direction_INCOMING {
		if status, ok := IncomingStatus[tr.Status]; ok {
			return status
		}
		return "-"
	}
	if status, ok := OutgoingStatus[tr.Status]; ok {
		return status
	}
	return "-"
}

// SetTransferAllFileStatus reset all files to status
func SetTransferAllFileStatus(tr *pb.Transfer, status pb.Status) {
	for _, file := range tr.Files {
		file.Status = status
	}
}

// ForAllFiles executes op for all files in files
func ForAllFiles(files []*pb.File, op func(*pb.File)) {
	for _, file := range files {
		op(file)
	}
}

func FindTransferFileByPath(tr *pb.Transfer, filePath string) *pb.File {
	predicate := func(f *pb.File) bool { return f.Path == filePath }
	return findTransferFile(tr, predicate)
}

func findTransferFile(tr *pb.Transfer, predicate func(*pb.File) bool) *pb.File {
	for _, file := range tr.Files {
		if predicate(file) {
			return file
		}
	}
	return nil
}

// Gets all files with the given path prefix (returns all children of directory)
func GetTransferFilesByPathPrefix(tr *pb.Transfer, filePath string) []*pb.File {
	var files []*pb.File
	for _, file := range tr.Files {
		if strings.HasPrefix(file.Path, filePath) {
			files = append(files, file)
		}
	}
	return files
}

// GetTransferFileStatus file status to human readable string
func GetTransferFileStatus(file *pb.File, in bool) (status string) {
	if in {
		if status, ok := IncomingFileStatus[file.GetStatus()]; ok {
			return status
		}
	} else {
		if status, ok := OutgoingFileStatus[file.GetStatus()]; ok {
			return status
		}
	}
	if status, ok := FileStatus[file.GetStatus()]; ok {
		return status
	}
	return "-"
}

func InternalTransferToPBTransfer(lt LibdropTransfer) *pb.Transfer {
	allFiles := filesFromTransferInfo(&lt)
	totalSize, totalTransferred := calculateTotalSizeAndTotalTransferred(allFiles)
	out := &pb.Transfer{
		Id:               lt.Id,
		Direction:        directionFromTransferInfo(&lt),
		Peer:             lt.Peer,
		Created:          timestamppb.New(time.UnixMilli(lt.CreatedAt)),
		Files:            allFiles,
		TotalSize:        totalSize,
		TotalTransferred: totalTransferred,
		Status:           determineTransferStatusAndAdjustFileStatuses(&lt, allFiles),
		Path:             determineTransferPath(&lt, allFiles),
	}

	return out
}

func directionFromTransferInfo(lt *LibdropTransfer) pb.Direction {
	var direction pb.Direction
	switch lt.Kind.(type) {
	case TransferKindOutgoing:
		direction = pb.Direction_OUTGOING
	case TransferKindIncoming:
		direction = pb.Direction_INCOMING
	default:
		log.Printf(internal.WarningPrefix+" unknown direction found when parsing libdrop transfers: %T\n", lt.Kind)
		direction = pb.Direction_UNKNOWN_DIRECTION
	}
	return direction
}

func filesFromTransferInfo(lt *LibdropTransfer) []*pb.File {
	switch ti := lt.Kind.(type) {
	case TransferKindIncoming:
		files := make([]*pb.File, len(ti.Paths))
		for i, path := range ti.Paths {
			files[i] = norddropIncomingPathToInternalFile(path)
		}
		return files
	case TransferKindOutgoing:
		files := make([]*pb.File, len(ti.Paths))
		for i, path := range ti.Paths {
			files[i] = norddropOutgoingPathToInternalFile(path)
		}
		return files
	default:
		log.Printf(internal.WarningPrefix+" unknown transfer kind: %T\n", ti)
		return []*pb.File{}
	}
}

func norddropOutgoingPathToInternalFile(outPath OutgoingPath) *pb.File {
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

func statusFromOutgoingPath(outPath *OutgoingPath) pb.Status {
	var status pb.Status

	if len(outPath.States) == 0 {
		status = pb.Status_REQUESTED
	} else {
		lastState := outPath.States[len(outPath.States)-1]
		switch lastState := lastState.Kind.(type) {
		case OutgoingPathStateKindStarted:
			status = pb.Status_ONGOING
		case OutgoingPathStateKindFailed:
			status = pb.Status(lastState.Status)
		case OutgoingPathStateKindCompleted:
			status = pb.Status_SUCCESS
		case OutgoingPathStateKindRejected:
			status = pb.Status_CANCELED
		case OutgoingPathStateKindPaused:
			status = pb.Status_PAUSED
		default:
			log.Printf(internal.WarningPrefix+" unknown file status in transfer: %T\n", lastState)
			status = pb.Status_BAD_STATUS
		}
	}
	return status
}

func transferredFromOutgoingPath(outPath *OutgoingPath) uint64 {
	var transferred uint64
	for _, state := range outPath.States {
		switch st := state.Kind.(type) {
		case OutgoingPathStateKindStarted:
			transferred = st.BytesSent
		case OutgoingPathStateKindCompleted:
			transferred = outPath.Bytes
		case OutgoingPathStateKindFailed:
			transferred = st.BytesSent
		case OutgoingPathStateKindRejected:
			transferred = st.BytesSent
		case OutgoingPathStateKindPaused:
			transferred = st.BytesSent
		}
	}
	return transferred
}

func determineFullOutgoingPath(outPath *OutgoingPath) string {
	switch pathSource := outPath.Source.(type) {
	case OutgoingFileSourceBasePath:
		return filepath.Join(pathSource.BasePath, outPath.RelativePath)
	default:
		log.Printf(internal.WarningPrefix+" unsupported path source: %T\n", outPath.Source)
		return ""
	}
}

func norddropIncomingPathToInternalFile(inPath IncomingPath) *pb.File {
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

func statusFromIncomingPath(inPath *IncomingPath) pb.Status {
	var status pb.Status
	if len(inPath.States) == 0 {
		status = pb.Status_REQUESTED
	} else {
		lastState := inPath.States[len(inPath.States)-1]
		switch lastState := lastState.Kind.(type) {
		case IncomingPathStateKindCompleted:
			status = pb.Status_SUCCESS
		case IncomingPathStateKindFailed:
			status = pb.Status(lastState.Status)
		case IncomingPathStateKindPaused:
			status = pb.Status_PAUSED
		case IncomingPathStateKindPending:
			status = pb.Status_PENDING
		case IncomingPathStateKindRejected:
			status = pb.Status_CANCELED
		case IncomingPathStateKindStarted:
			status = pb.Status_ONGOING
		default:
			log.Printf(internal.WarningPrefix+" unknown file status in transfer: %T\n", lastState)
			status = pb.Status_BAD_STATUS
		}
	}
	return status
}

func transferredFromIncomingPath(inPath *IncomingPath) uint64 {
	var transferred uint64
	for _, state := range inPath.States {
		switch st := state.Kind.(type) {
		case IncomingPathStateKindStarted:
			transferred = st.BytesReceived
		case IncomingPathStateKindCompleted:
			transferred = inPath.Bytes
		case IncomingPathStateKindFailed:
			transferred = st.BytesReceived
		case IncomingPathStateKindRejected:
			transferred = st.BytesReceived
		case IncomingPathStateKindPaused:
			transferred = st.BytesReceived
		}
	}
	return transferred
}

func determineFullIncomingPath(inPath *IncomingPath) string {
	fullPath := inPath.RelativePath
	for _, state := range inPath.States {
		switch st := state.Kind.(type) {
		case IncomingPathStateKindPending:
			// BasePath is provided with the very first "pending" state
			fullPath = filepath.Join(st.BaseDir, fullPath)
		case IncomingPathStateKindCompleted:
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

func determineTransferStatusAndAdjustFileStatuses(libdropTransfer *LibdropTransfer, allFiles []*pb.File) pb.Status {
	var status pb.Status
	if len(libdropTransfer.States) != 0 {
		lastState := libdropTransfer.States[len(libdropTransfer.States)-1]
		switch lastState := lastState.Kind.(type) {
		case TransferStateKindCancel:
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
		case TransferStateKindFailed:
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

func determineTransferPath(libdropTransfer *LibdropTransfer, allFiles []*pb.File) string {
	var result string
	var transferPath string
	// Base path is in different place for incoming and outgoing files
	switch ti := libdropTransfer.Kind.(type) {
	case TransferKindIncoming:
		for _, path := range ti.Paths {
			for _, state := range path.States {
				switch st := state.Kind.(type) {
				case IncomingPathStateKindPending:
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
	case TransferKindOutgoing:
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
					case OutgoingFileSourceBasePath:
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
