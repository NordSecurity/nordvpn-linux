package fileshare

import (
	"strings"

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
		pb.Status_FILE_REJECTED:        "the receiver has declined the file transfer",
	}
	FileStatus = map[pb.Status]string{
		pb.Status_SUCCESS:                  "completed",
		pb.Status_CANCELED:                 "canceled",
		pb.Status_INTERRUPTED:              "interrupted",
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

func isTransferFinished(tr *LiveTransfer) bool {
	for _, file := range tr.Files {
		if !file.Finished {
			return false
		}
	}
	return true
}
