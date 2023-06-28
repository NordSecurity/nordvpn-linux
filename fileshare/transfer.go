package fileshare

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
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
	}
	FileStatus = map[pb.Status]string{
		pb.Status_SUCCESS:                  "completed",
		pb.Status_CANCELED:                 "canceled",
		pb.Status_INTERRUPTED:              "interrupted",
		pb.Status_BAD_PATH:                 "bad path",
		pb.Status_BAD_FILE:                 "bad file",
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
		pb.Status_TRANSFER_LIMITS_EXCEEDED: "limits exceeded",
		pb.Status_MISMATCHED_SIZE:          "file was changed",
		pb.Status_UNEXPECTED_DATA:          "file was changed",
		pb.Status_INVALID_ARGUMENT:         "internal error",
		pb.Status_TRANSFER_TIMEOUT:         "transfer timeout",
		pb.Status_WS_SERVER:                "waiting for peer to come online",
		pb.Status_WS_CLIENT:                "waiting for peer to come online",
		pb.Status_FILE_MODIFIED:            "file was changed",
		pb.Status_FILENAME_TOO_LONG:        "filename too long",
		pb.Status_AUTHENTICATION_FAILED:    "authentication failed",
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

// NewOutgoingTransfer creates new transfer and initializes internals
func NewOutgoingTransfer(id, peer, path string) *pb.Transfer {
	return &pb.Transfer{
		Id:        id,
		Peer:      peer,
		Direction: pb.Direction_OUTGOING,
		Status:    pb.Status_REQUESTED,
		Created:   timestamppb.Now(),
		Path:      path, // path to be sent: single file or dir
		Finalized: false,
		// Not adding files because the user might not have permission to access them
		// and we don't want to leak information about them. Add files only when they
		// are started to be sent.
	}
}

// NewIncomingTransfer creates new transfer and initializes internals
func NewIncomingTransfer(id, peer string, files []*pb.File) *pb.Transfer {
	tr := &pb.Transfer{
		Id:        id,
		Peer:      peer,
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
		Created:   timestamppb.Now(),
		Files:     files,
		Finalized: false,
	}
	for _, file := range tr.Files {
		file.Status = pb.Status_REQUESTED
	}
	return tr
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

// SetFileStatus finds file with fileID in files and updates its status
func SetFileStatus(files []*pb.File, fileID string, status pb.Status) error {
	for _, file := range files {
		if file.Id == fileID {
			file.Status = status
			return nil
		}
	}
	return fmt.Errorf("status %s reported for nonexistent file", status.String())
}

func FindTransferFileByID(tr *pb.Transfer, fileID string) *pb.File {
	predicate := func(f *pb.File) bool { return f.Id == fileID }
	return findTransferFile(tr, predicate)
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

// GetNewTransferStatus returns new transfer status based on files
//
// If at least one file did not finish transferring(status is REQUESTED or ONGOING), transfer status
// is unchanged, otherwise:
//
//   - if status of all of the files CANCELED, new transfer status is CANCELED
//
//   - if at least one file status is SUCCESS, new transfer status is SUCCESS
//
//   - if at least one file status is SUCCESS and at leas one file status is erroneous,
//     new transfer status is FINISHED_WITH_ERRORS
func GetNewTransferStatus(files []*pb.File, currentStatus pb.Status) pb.Status {
	allCanceled := true
	allFinished := true
	hasNoErrors := true
	for _, file := range files {
		if allCanceled && file.Status != pb.Status_CANCELED {
			allCanceled = false
		}
		if allFinished && !isFileCompleted(file.Status) {
			allFinished = false
		}
		if hasNoErrors && checkFileHasErrors(file) {
			hasNoErrors = false
		}
	}

	if allCanceled {
		return pb.Status_CANCELED
	} else if allFinished {
		if hasNoErrors {
			return pb.Status_SUCCESS
		} else {
			return pb.Status_FINISHED_WITH_ERRORS
		}
	}

	return currentStatus
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

func checkFileHasErrors(file *pb.File) bool {
	return file.Status != pb.Status_SUCCESS &&
		file.Status != pb.Status_REQUESTED &&
		file.Status != pb.Status_CANCELED &&
		file.Status != pb.Status_ONGOING
}

func isFileCompleted(fileStatus pb.Status) bool {
	return fileStatus != pb.Status_REQUESTED &&
		fileStatus != pb.Status_ONGOING
}

// isTransferFinished check transfer status to be one of
func isTransferFinished(tr *pb.Transfer) bool {
	return tr.Status == pb.Status_FINISHED_WITH_ERRORS ||
		tr.Status == pb.Status_SUCCESS ||
		tr.Status == pb.Status_CANCELED ||
		tr.Status == pb.Status_CANCELED_BY_PEER
}

// TransferProgressInfo info to report to the user
type TransferProgressInfo struct {
	TransferID  string
	Transferred uint32 // percent of transferred bytes
	Status      pb.Status
}
