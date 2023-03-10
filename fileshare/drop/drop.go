// Package drop wraps libdrop fileshare implementation.
package drop

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"sync"

	norddropgo "github.com/NordSecurity/libdrop/norddrop/ffi/bindings/linux/go"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Fileshare is the main functional filesharing implementation using norddrop library.
// Thread safe.
type Fileshare struct {
	norddrop     norddropgo.Norddrop
	eventsDbPath string
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
		"DROP: "+message,
	)
}

// New initializes norddrop library.
func New(eventFunc func(string), eventsDbPath string, isProd bool) *Fileshare {
	logLevel := norddropgo.NORDDROPLOGTRACE
	if isProd {
		logLevel = norddropgo.NORDDROPLOGERROR
	}
	return &Fileshare{
		norddrop:     norddropgo.NewNorddrop(eventFunc, logLevel, logCB),
		eventsDbPath: eventsDbPath,
		isProd:       isProd,
	}
}

// Enable executes Start in norddrop library. Has to be called before using other Fileshare methods.
func (f *Fileshare) Enable(listenAddr netip.Addr) (err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if err = f.start(listenAddr, f.eventsDbPath, f.isProd); err != nil {
		return fmt.Errorf("starting drop: %w", err)
	}

	return nil
}

type libdropStartConfig struct {
	DirDepthLimit          uint64 `json:"dir_depth_limit"`
	TransferFileLimit      uint64 `json:"transfer_file_limit"`
	ReqConnectionTimeoutMs uint64 `json:"req_connection_timeout_ms"`
	TransferIdleLifetimeMs uint64 `json:"transfer_idle_lifetime_ms"`
	MooseEventPath         string `json:"moose_event_path"`
	IsProd                 bool   `json:"moose_prod"`
}

func (f *Fileshare) start(listenAddr netip.Addr, eventsDbPath string, isProd bool) error {
	configJSON, err := json.Marshal(libdropStartConfig{
		DirDepthLimit:          fileshare.DirDepthLimit,
		TransferFileLimit:      fileshare.TransferFileLimit,
		ReqConnectionTimeoutMs: fileshare.ReqConnectionTimeoutMs,
		TransferIdleLifetimeMs: fileshare.TransferIdleLifetimeMs,
		MooseEventPath:         eventsDbPath,
		IsProd:                 isProd,
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

	res := f.norddrop.CancelFile(transferID, fileID)
	if err := toError(res); err != nil {
		return err
	}

	return nil
}
