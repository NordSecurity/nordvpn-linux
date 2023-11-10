// Package libdrop wraps libdrop fileshare implementation.
package libdrop

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"sync"
	"time"

	norddropgo "github.com/NordSecurity/libdrop/norddrop/ffi/bindings/linux/go"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	DirDepthLimit     = 5
	TransferFileLimit = 1000
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
		DirDepthLimit:     DirDepthLimit,
		TransferFileLimit: TransferFileLimit,
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

// Transfer as represented in libdrop storage
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

func (f *Fileshare) GetTransfersSince(t time.Time) ([]Transfer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	transfers := []Transfer{}
	rawTransfers := f.norddrop.GetTransfersSince(t.Unix())
	err := json.Unmarshal([]byte(rawTransfers), &transfers)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling libdrop transfers JSON: %w", err)
	}

	return transfers, nil
}
