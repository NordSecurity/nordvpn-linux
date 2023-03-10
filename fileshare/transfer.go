package fileshare

import (
	"fmt"

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
		Finalized: false,
	}
	SetTransferFiles(tr, files)
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

// SetTransferFiles set files to transfer and initialize status
func SetTransferFiles(tr *pb.Transfer, files []*pb.File) {
	tr.Files = files
	SetTransferAllFileStatus(tr, pb.Status_REQUESTED)
	setTransferAllFilePath(tr) // TODO: will be not needed when libdrop fix
}

// SetTransferAllFileStatus reset all files to status
func SetTransferAllFileStatus(tr *pb.Transfer, status pb.Status) {
	for _, file := range tr.Files {
		setAllFileStatus(file, status)
	}
}

// ForAllFiles executes op for all files in files
func ForAllFiles(files []*pb.File, op func(*pb.File)) {
	for _, fileTree := range files {
		forAllFilesInTree(fileTree, op)
	}
}

func forAllFilesInTree(file *pb.File, op func(*pb.File)) {
	op(file)
	if len(file.Children) > 0 {
		for _, childFile := range file.Children {
			forAllFilesInTree(childFile, op)
		}
	}
}

func setAllFileStatus(file *pb.File, status pb.Status) {
	file.Status = status
	if len(file.Children) > 0 {
		for _, childFile := range file.Children {
			setAllFileStatus(childFile, status)
		}
	}
}

// setAllFilePath prepend file path to file id
// TODO: will be not needed when libdrop fix
func setTransferAllFilePath(tr *pb.Transfer) {
	for _, file := range tr.Files {
		setAllFilePath(file, "")
	}
}

func setAllFilePath(file *pb.File, path string) {
	if path != "" {
		path += "/"
	}
	file.Id = path + file.Id
	if len(file.Children) > 0 {
		for _, childFile := range file.Children {
			setAllFilePath(childFile, file.Id)
		}
	}
}

// GetAllTransferFiles get all files in flat list
func GetAllTransferFiles(tr *pb.Transfer) (allFiles []*pb.File) {
	for _, file := range tr.Files {
		allFiles = append(allFiles, getFileFiles(file)...)
	}
	return
}

func getFileFiles(file *pb.File) (allFiles []*pb.File) {
	if len(file.Children) > 0 {
		for _, childFile := range file.Children {
			allFiles = append(allFiles, getFileFiles(childFile)...)
		}
		return
	}
	return []*pb.File{file}
}

// SetFileStatus finds file with fileID in files and updates its status
func SetFileStatus(files []*pb.File, fileID string, status pb.Status) error {
	fileFound := false
	for _, file := range files {
		if findAndSetFileStatus(file, fileID, status) {
			fileFound = true
			break
		}
	}

	if !fileFound {
		return fmt.Errorf("status %s reported for nonexistent file", status.String())
	}

	return nil
}

func findAndSetFileStatus(file *pb.File, fileID string, status pb.Status) bool {
	if file.Id == fileID {
		file.Status = status
		return true
	}
	if len(file.Children) > 0 { // dir, check childs
		fileFound := false
		for _, childFile := range file.Children {
			if findAndSetFileStatus(childFile, fileID, status) {
				fileFound = true
				break
			}
		}
		return fileFound
	}
	return false
}

// FindTransferFile find file in a tree
func FindTransferFile(tr *pb.Transfer, fileID string) *pb.File {
	if fileID != "" {
		for _, file := range tr.Files {
			if foundFile := findFile(file, fileID); foundFile != nil {
				return foundFile
			}
		}
	}
	return nil
}

func findFile(file *pb.File, fileID string) *pb.File {
	if file.Id == fileID {
		return file
	}
	if len(file.Children) > 0 { // dir, check childs
		for _, childFile := range file.Children {
			if foundFile := findFile(childFile, fileID); foundFile != nil {
				return foundFile
			}
		}
	}
	return nil
}

// CountTransferFiles count files in a tree
func CountTransferFiles(tr *pb.Transfer) (count uint64) {
	for _, file := range tr.Files {
		count += countFiles(file)
	}
	return
}

func countFiles(file *pb.File) (count uint64) {
	if len(file.Children) > 0 { // dir, check childs
		for _, childFile := range file.Children {
			count += countFiles(childFile)
		}
	} else {
		return 1
	}
	return count
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
//   - if at least one file status is SUCCESS and at leas one file status is erronous,
//     new transfer status is FINISHED_WITH_ERRORS
func GetNewTransferStatus(files []*pb.File, currentStatus pb.Status) pb.Status {
	allCanceled := true
	allFinished := true
	hasNoErrors := true
	for _, file := range files {
		if allCanceled && !checkAllFilesCanceled(file) {
			allCanceled = false
		}
		if allFinished && !checkAllFilesFinished(file) {
			allFinished = false
		}
		if hasNoErrors && checkFilesHasErrors(file) {
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

// getFileStatus find out status if children are
func getFileStatus(file *pb.File) (status pb.Status) {
	if len(file.Children) > 0 {
		if checkAllFilesCanceled(file) {
			return pb.Status_CANCELED
		}
		allFinished := checkAllFilesFinished(file)
		if allFinished && checkFilesHasErrors(file) {
			return pb.Status_FINISHED_WITH_ERRORS
		}
		if allFinished {
			return pb.Status_SUCCESS
		}
		if checkAllFilesRequested(file) {
			return pb.Status_REQUESTED
		}
		return pb.Status_ONGOING
	}
	return file.GetStatus()
}

// GetTransferFileStatus file status to human readable string
func GetTransferFileStatus(file *pb.File, in bool) (status string) {
	fileStatus := getFileStatus(file)
	if in {
		if status, ok := IncomingFileStatus[fileStatus]; ok {
			return status
		}
	} else {
		if status, ok := OutgoingFileStatus[fileStatus]; ok {
			return status
		}
	}
	if status, ok := FileStatus[fileStatus]; ok {
		return status
	}
	return "-"
}

func checkAllFilesRequested(file *pb.File) bool {
	if len(file.Children) > 0 { // dir, check childs
		for _, childFile := range file.Children {
			if !checkAllFilesRequested(childFile) { // one breaks it all
				return false
			}
		}
		return true
	}
	return file.Status == pb.Status_REQUESTED
}

func checkAllFilesCanceled(file *pb.File) bool {
	if len(file.Children) > 0 { // dir, check childs
		for _, childFile := range file.Children {
			if !checkAllFilesCanceled(childFile) { // one breaks it all
				return false
			}
		}
		return true
	}
	return file.Status == pb.Status_CANCELED
}

func checkAllFilesFinished(file *pb.File) bool {
	if len(file.Children) > 0 { // dir, check childs
		for _, childFile := range file.Children {
			if !checkAllFilesFinished(childFile) { // one breaks it all
				return false
			}
		}
		return true
	}
	return isFileCompleted(file.Status)
}

func checkFilesHasErrors(file *pb.File) bool {
	if len(file.Children) > 0 { // dir, check childs
		for _, childFile := range file.Children {
			if checkFilesHasErrors(childFile) { // one breaks it all
				return true
			}
		}
		return false
	}
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

// TransferProgessInfo info to report to the user
type TransferProgressInfo struct {
	TransferID  string
	Transferred uint32 // percent of transfered bytes
	Status      pb.Status
}
