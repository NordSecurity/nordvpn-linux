package fileshare

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/protobuf/types/known/timestamppb"
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

// LibdropTransfer as represented in libdrop storage
type LibdropTransfer struct {
	ID        string                 `json:"id"`
	Peer      string                 `json:"peer_id"`
	CreatedAt int64                  `json:"created_at"`
	States    []LibdropTransferState `json:"states"`
	Direction string                 `json:"type"`
	Files     []LibdropFile          `json:"paths"`
}

type LibdropTransferState struct {
	CreatedAt  uint64 `json:"created_at"`
	State      string `json:"state"`
	ByPeer     bool   `json:"by_peer"`
	StatusCode int    `json:"status_code"`
}

type LibdropFile struct {
	ID           string             `json:"file_id"`
	TransferID   string             `json:"transfer_id"`
	BasePath     string             `json:"base_path"`
	RelativePath string             `json:"relative_path"`
	TotalSize    uint64             `json:"bytes"`
	CreatedAt    uint64             `json:"created_at"`
	States       []LibdropFileState `json:"states"`
}

type LibdropFileState struct {
	CreatedAt     uint64 `json:"created_at"`
	State         string `json:"state"`
	BytesSent     uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
	BasePath      string `json:"base_dir"`
	FinalPath     string `json:"final_path"`
	StatusCode    int    `json:"status_code"`
}

// Converts libdrop transfer representation to our own
func LibdropTransferToInternalTransfer(in LibdropTransfer) *pb.Transfer {
	out := &pb.Transfer{}

	out.Id = in.ID
	out.Peer = in.Peer
	out.Created = timestamppb.New(time.UnixMilli(in.CreatedAt))

	switch in.Direction {
	case "outgoing":
		out.Direction = pb.Direction_OUTGOING
	case "incoming":
		out.Direction = pb.Direction_INCOMING
	default:
		log.Printf("%s unknown direction found when parsing libdrop transfers: %s",
			internal.WarningPrefix, in.Direction)
		out.Direction = pb.Direction_UNKNOWN_DIRECTION
	}

	for _, file := range in.Files {
		outFile := libdropFileToInternalFile(file)
		out.Files = append(out.Files, outFile)
		if isFileTransferred(outFile) {
			out.TotalSize += outFile.Size
			out.TotalTransferred += outFile.Transferred
		}

		// Determine transfer path.

		var fileBasePath string
		// Base path is in different place for incoming and outgoing files
		if file.BasePath != "" {
			fileBasePath = file.BasePath
		} else if len(file.States) != 0 && file.States[0].BasePath != "" {
			fileBasePath = file.States[0].BasePath
		}

		if out.Direction == pb.Direction_OUTGOING {
			if filepath.Base(file.RelativePath) == file.RelativePath {
				// If relative path doesn't contain a directory - means user specified a single file
				// We show the full path in that case for outgoing transfers
				fileBasePath = outFile.FullPath
			} else {
				// If user is sending a directory - we add the directory itself to the path.
				// Example: user sends /tmp/test, then file.BasePath would be "/tmp", while
				// file.RelativePath would be "test/file". So we take the dir name from RelativePath
				// and add it to BasePath to get "/tmp/test".
				dir, _, ok := strings.Cut(file.RelativePath, string(filepath.Separator))
				if ok {
					fileBasePath = filepath.Join(file.BasePath, dir)
				}
			}
		}

		if out.Path == "" {
			out.Path = fileBasePath
		} else if out.Path != fileBasePath && fileBasePath != "" {
			out.Path = "multiple files"
		}
	}

	if len(in.States) != 0 {
		lastState := in.States[len(in.States)-1]
		switch lastState.State {
		case "cancel":
			// This is annoying. We have to "finalize" finished transfers by cancelling them,
			// otherwise there's a resource leak in libdrop. Also, by doing that all finished
			// transfers have cancel status, so we need to figure out the real status.
			// So we try to determine status by files and see if the transfer was already finished
			// when it was cancelled or not.
			statusByFiles := getTransferStatus(out.Files)
			if statusByFiles == pb.Status_SUCCESS || statusByFiles == pb.Status_FINISHED_WITH_ERRORS {
				out.Status = statusByFiles
			} else {
				if lastState.ByPeer {
					out.Status = pb.Status_CANCELED_BY_PEER
				} else {
					out.Status = pb.Status_CANCELED
				}
				for _, file := range out.Files {
					if !isFileCompleted(file) {
						file.Status = pb.Status_CANCELED
					}
				}
			}
		case "failed":
			out.Status = pb.Status(lastState.StatusCode)
		}
	} else {
		out.Status = getTransferStatus(out.Files)
	}

	return out
}

func libdropFileToInternalFile(in LibdropFile) *pb.File {
	out := &pb.File{
		Id:       in.ID,
		Path:     in.RelativePath,
		FullPath: filepath.Join(in.BasePath, in.RelativePath),
		Size:     in.TotalSize,
		// This only shows the amount from the last status change.
		// The only up to date source for data transferred are events.
		// To show the correct value we track ongoing transfers in event manager,
		// and event manager overwrites values of transferred fields if needed.
		Transferred: 0,
	}

	for _, state := range in.States {
		// BasePath is provided with the very first "pending" state
		if state.BasePath != "" {
			out.FullPath = filepath.Join(state.BasePath, out.FullPath)
		}
		// FinalPath is provided with the very last "completed" state, so if it is present, then
		// it will overwrite the previously constructed path
		if state.FinalPath != "" {
			out.FullPath = state.FinalPath
		}

		if state.BytesReceived != 0 {
			out.Transferred = state.BytesReceived
		}
		if state.BytesSent != 0 {
			out.Transferred = state.BytesSent
		}
	}

	if len(in.States) == 0 {
		out.Status = pb.Status_REQUESTED
	} else {
		lastState := in.States[len(in.States)-1]
		switch lastState.State {
		case "completed":
			out.Transferred = out.Size
			out.Status = pb.Status_SUCCESS
		case "failed":
			out.Status = pb.Status(lastState.StatusCode)
		case "paused":
			out.Status = pb.Status_PAUSED
		case "pending":
			out.Status = pb.Status_PENDING
		case "rejected":
			out.Status = pb.Status_CANCELED
		case "started":
			out.Status = pb.Status_ONGOING
		default:
			log.Printf("%s unknown file status found when parsing libdrop transfers: %s",
				internal.WarningPrefix, lastState.State)
			out.Status = pb.Status_BAD_STATUS
		}
	}

	return out
}

func getTransferStatus(files []*pb.File) pb.Status {
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
