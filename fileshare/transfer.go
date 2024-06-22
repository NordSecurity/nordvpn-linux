package fileshare

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	norddrop "github.com/NordSecurity/libdrop-go/v7"
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

func LibdropTransferToInternalTransfer(ti norddrop.TransferInfo) *pb.Transfer {
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

func filesFromTransferInfo(ti *norddrop.TransferInfo) []*pb.File {
	switch ti := ti.Kind.(type) {
	case norddrop.TransferKindIncoming:
		files := make([]*pb.File, len(ti.Paths))
		for i, path := range ti.Paths {
			files[i] = norddropIncomingPathToInternalFile(path)
		}
		return files
	case norddrop.TransferKindOutgoing:
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

func norddropOutgoingPathToInternalFile(outPath norddrop.OutgoingPath) *pb.File {
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

func norddropIncomingPathToInternalFile(inPath norddrop.IncomingPath) *pb.File {
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

func determineTransferStatusAndAdjustFileStatuses(transferInfo *norddrop.TransferInfo, allFiles []*pb.File) pb.Status {
	var status pb.Status
	if len(transferInfo.States) != 0 {
		lastState := transferInfo.States[len(transferInfo.States)-1]
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

func determineTransferPath(transferInfo *norddrop.TransferInfo, allFiles []*pb.File) string {
	var result string
	var transferPath string
	// Base path is in different place for incoming and outgoing files
	switch ti := transferInfo.Kind.(type) {
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
