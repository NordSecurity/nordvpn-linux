// Package drop wraps libdrop fileshare implementation.
package drop

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"path/filepath"
	"sync"
	"time"

	norddropgo "github.com/NordSecurity/libdrop/norddrop/ffi/bindings/linux/go"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Fileshare is the main functional filesharing implementation using norddrop library.
// Thread safe.
type Fileshare struct {
	norddrop     norddropgo.Norddrop
	eventsDbPath string
	appVersion   string
	storagePath  string
	isProd       bool
	mutex        sync.Mutex
}

func logLevelToPrefix(level norddropgo.Enum_SS_norddrop_log_level) string {
	switch level {
	case norddropgo.NORDDROPLOGCRITICAL, norddropgo.NORDDROPLOGERROR:
		return internal.ErrorPrefix
	case norddropgo.NORDDROPLOGWARNING:
		return internal.WarningPrefix
	case norddropgo.NORDDROPLOGDEBUG, norddropgo.NORDDROPLOGTRACE:
		return internal.DebugPrefix
	default:
		return internal.InfoPrefix
	}
}

func logCB(level int, message string) {
	log.Println(
		logLevelToPrefix(norddropgo.Enum_SS_norddrop_log_level(level)),
		"DROP("+norddropgo.NorddropVersion()+"): "+message,
	)
}

// New initializes norddrop library.
func New(
	eventFunc func(string),
	eventsDbPath string,
	appVersion string,
	isProd bool,
	pubkeyFunc func(string) []byte,
	privKey string,
	storagePath string,
) *Fileshare {
	logLevel := norddropgo.NORDDROPLOGTRACE
	if isProd {
		logLevel = norddropgo.NORDDROPLOGERROR
	}
	return &Fileshare{
		norddrop:     norddropgo.NewNorddrop(eventFunc, logLevel, logCB, pubkeyFunc, privKey),
		eventsDbPath: eventsDbPath,
		appVersion:   appVersion,
		storagePath:  storagePath,
		isProd:       isProd,
	}
}

// Enable executes Start in norddrop library. Has to be called before using other Fileshare methods.
func (f *Fileshare) Enable(listenAddr netip.Addr) (err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	log.Println(internal.InfoPrefix, "libdrop version:", norddropgo.NorddropVersion())

	if err = f.start(listenAddr, f.eventsDbPath, f.appVersion, f.isProd, f.storagePath); err != nil {
		return fmt.Errorf("starting drop: %w", err)
	}

	return nil
}

type libdropStartConfig struct {
	DirDepthLimit     uint64 `json:"dir_depth_limit"`
	TransferFileLimit uint64 `json:"transfer_file_limit"`
	MooseEventPath    string `json:"moose_event_path"`
	LinuxAppVersion   string `json:"moose_app_version"`
	IsProd            bool   `json:"moose_prod"`
	StoragePath       string `json:"storage_path"`
}

func (f *Fileshare) start(
	listenAddr netip.Addr,
	eventsDbPath string,
	appVersion string,
	isProd bool,
	storagePath string,
) error {
	configJSON, err := json.Marshal(libdropStartConfig{
		DirDepthLimit:     fileshare.DirDepthLimit,
		TransferFileLimit: fileshare.TransferFileLimit,
		MooseEventPath:    eventsDbPath,
		LinuxAppVersion:   appVersion,
		IsProd:            isProd,
		StoragePath:       storagePath,
	})
	if err != nil {
		return fmt.Errorf("marshalling libdrop config: %w", err)
	}
	return toError(f.norddrop.Start(listenAddr.String(), string(configJSON)))
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
	return toError(f.norddrop.Stop())
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

	type transferDescriptor struct {
		Path string `json:"path"`
	}

	transferDescriptors := []transferDescriptor{}
	for _, path := range paths {
		transferDescriptors = append(transferDescriptors, transferDescriptor{Path: path})
	}

	json, err := json.Marshal(transferDescriptors)
	if err != nil {
		return "", err
	}

	transfer := f.norddrop.NewTransfer(peer.String(), string(json))
	if transfer == "" {
		return "", fmt.Errorf("transfer wasn't created")
	}
	return transfer, nil
}

// Accept starts downloading provided files into dstPath.
// dstPath must be absolute.
func (f *Fileshare) Accept(transferID, dstPath string, fileID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	res := f.norddrop.Download(transferID, fileID, dstPath)
	if err := toError(res); err != nil {
		return err
	}

	return nil
}

// Cancel file transfer.
func (f *Fileshare) Cancel(transferID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	res := f.norddrop.CancelTransfer(transferID)
	return toError(res)
}

// CancelFile id in a transfer
func (f *Fileshare) CancelFile(transferID string, fileID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	res := f.norddrop.RejectFile(transferID, fileID)
	if err := toError(res); err != nil {
		return err
	}

	return nil
}

type Transfer struct {
	ID        string          `json:"id"`
	Peer      string          `json:"peer_id"`
	CreatedAt int64           `json:"created_at"`
	States    []TransferState `json:"states"`
	Direction string          `json:"type"`
	Files     []File          `json:"paths"`
}

type TransferState struct {
	CreatedAt  uint64 `json:"created_at"`
	State      string `json:"state"`
	ByPeer     bool   `json:"by_peer"`
	StatusCode int    `json:"status_code"`
}

type File struct {
	ID           string      `json:"file_id"`
	TransferID   string      `json:"transfer_id"`
	BasePath     string      `json:"base_path"`
	RelativePath string      `json:"relative_path"`
	TotalSize    uint64      `json:"bytes"`
	CreatedAt    uint64      `json:"created_at"`
	States       []FileState `json:"states"`
}

type FileState struct {
	CreatedAt     uint64 `json:"created_at"`
	State         string `json:"state"`
	BytesSent     uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
	BasePath      string `json:"base_dir"`
	FinalPath     string `json:"final_path"`
	StatusCode    int    `json:"status_code"`
}

func (f *Fileshare) Load() (map[string]*pb.Transfer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	libdropTransfers := []Transfer{}
	transfers := map[string]*pb.Transfer{}
	rawTransfers := f.norddrop.GetTransfersSince(0)
	err := json.Unmarshal([]byte(rawTransfers), &libdropTransfers)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling libdrop transfers JSON: %w", err)
	}

	for _, t := range libdropTransfers {
		transfer := libdropTransferToInternalTransfer(t)
		transfers[transfer.Id] = transfer
	}

	return transfers, nil
}

func libdropTransferToInternalTransfer(in Transfer) *pb.Transfer {
	out := &pb.Transfer{}

	out.Id = in.ID
	out.Peer = in.Peer
	out.Created = timestamppb.New(time.UnixMilli(in.CreatedAt))

	switch in.Direction {
	case "outgoing":
		out.Direction = pb.Direction_OUTGOING
	case "incoming":
		out.Direction = pb.Direction_OUTGOING
	default:
		log.Printf("%s unknown direction found when parsing libdrop transfers: %s",
			internal.WarningPrefix, in.Direction)
		out.Direction = pb.Direction_UNKNOWN_DIRECTION
	}

	for _, file := range in.Files {
		outFile := libdropFileToInternalFile(file)
		out.Files = append(out.Files, outFile)
		out.TotalSize += outFile.Size
		out.TotalTransferred += outFile.Transferred

		// Determine transfer path.

		var fileBasePath string
		// Base path is in different place for incoming and outgoing files
		if file.BasePath != "" {
			fileBasePath = file.BasePath
		} else if len(file.States) != 0 && file.States[0].BasePath != "" {
			fileBasePath = file.States[0].BasePath
		}

		// If relative path doesn't contain a directory - means user specified a single file
		// We show the full path in that case for outgoing transfers
		if out.Direction == pb.Direction_OUTGOING && filepath.Base(file.RelativePath) == file.RelativePath {
			fileBasePath = outFile.FullPath
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
		case "canceled":
			if lastState.ByPeer {
				out.Status = pb.Status_CANCELED_BY_PEER
			} else {
				out.Status = pb.Status_CANCELED
			}
		case "failed":
			out.Status = pb.Status(lastState.StatusCode)
		}
	} else {
		out.Status = getTransferStatus(out.Files)
	}

	return out
}

func libdropFileToInternalFile(in File) *pb.File {
	out := &pb.File{
		Id:       in.ID,
		Path:     in.RelativePath,
		FullPath: filepath.Join(in.BasePath, in.RelativePath),
		Size:     in.TotalSize,
		// This only shows the amount from the last status change
		// The only up to date source for data transferred is events
		// To show the correct value we track ongoing transfers (TransferProgress) in event manager,
		// and event manager overwrites values of ongoing transfers before passing them further
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
			out.Status = pb.Status_SUCCESS
		case "failed":
			out.Status = pb.Status(lastState.StatusCode)
		case "paused":
			out.Status = pb.Status_PAUSED
		case "pending":
			out.Status = pb.Status_REQUESTED
		case "reject":
			out.Status = pb.Status_CANCELED
		case "started":
			out.Status = pb.Status_ONGOING
		default:
			log.Printf("%s unknown status found when parsing libdrop transfers: %s",
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
	hasStarted := true
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
		if hasStarted && file.Status != pb.Status_REQUESTED {
			hasStarted = false
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
		file.Status != pb.Status_ONGOING
}

func isFileCompleted(fileStatus pb.Status) bool {
	return fileStatus != pb.Status_REQUESTED &&
		fileStatus != pb.Status_ONGOING
}
