// Package libdrop wraps libdrop fileshare implementation.
package libdrop

import (
	"fmt"
	"log"
	"net/netip"
	"sync"
	"time"

	norddrop "github.com/NordSecurity/libdrop-go/v7"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Fileshare is the main functional filesharing implementation using norddrop library.
// Thread safe.
type Fileshare struct {
	norddrop     *norddrop.NordDrop
	eventsDbPath string
	storagePath  string
	isProd       bool
	mutex        sync.Mutex
}

func logLevelToPrefix(level norddrop.LogLevel) string {
	switch level {
	case norddrop.LogLevelCritical, norddrop.LogLevelError:
		return internal.ErrorPrefix
	case norddrop.LogLevelWarning:
		return internal.WarningPrefix
	case norddrop.LogLevelDebug, norddrop.LogLevelTrace:
		return internal.DebugPrefix
	case norddrop.LogLevelInfo:
		return internal.InfoPrefix
	default:
		return internal.InfoPrefix
	}
}

type defaultKeyStore struct {
	pubkeyFunc func(string) []byte
	privKey    string
}

func (dks defaultKeyStore) OnPubkey(peer string) *[]byte {
	pubKey := dks.pubkeyFunc(peer)
	return &pubKey
}

func (dks defaultKeyStore) Privkey() []byte {
	return []byte(dks.privKey)
}

type defaultLogger struct {
	logLevel norddrop.LogLevel
}

func (dl defaultLogger) OnLog(level norddrop.LogLevel, msg string) {
	log.Println(logLevelToPrefix(level), "DROP("+norddrop.Version()+"): "+msg)
}

func (dl defaultLogger) Level() norddrop.LogLevel {
	return dl.logLevel
}

// New initializes norddrop library.
func New(
	eventFunc norddrop.EventCallback,
	eventsDbPath string,
	isProd bool,
	pubkeyFunc func(string) []byte,
	privKey string,
	storagePath string,
) (*Fileshare, error) {
	keyStore := defaultKeyStore{
		pubkeyFunc: pubkeyFunc,
		privKey:    privKey,
	}
	logLevel := norddrop.LogLevelTrace
	if isProd {
		logLevel = norddrop.LogLevelError
	}

	logger := defaultLogger{logLevel}

	norddrop, err := norddrop.NewNordDrop(eventFunc, keyStore, logger)
	if err != nil {
		return nil, fmt.Errorf("creating norddrop instance: %w", err)
	}

	return &Fileshare{
		norddrop:     norddrop,
		eventsDbPath: eventsDbPath,
		storagePath:  storagePath,
		isProd:       isProd,
	}, nil
}

// Enable executes Start in norddrop library. Has to be called before using other Fileshare methods.
func (f *Fileshare) Enable(listenAddr netip.Addr) (err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	log.Println(internal.InfoPrefix, "libdrop version:", norddrop.Version())

	if err = f.start(listenAddr, f.eventsDbPath, f.isProd, f.storagePath); err != nil {
		return fmt.Errorf("starting drop: %w", err)
	}

	return nil
}

func (f *Fileshare) start(
	listenAddr netip.Addr,
	eventsDbPath string,
	isProd bool,
	storagePath string,
) error {
	config := norddrop.Config{
		DirDepthLimit:     fileshare.DirDepthLimit,
		TransferFileLimit: fileshare.TransferFileLimit,
		MooseEventPath:    eventsDbPath,
		MooseProd:         isProd,
		StoragePath:       storagePath,
	}

	return f.norddrop.Start(listenAddr.String(), config)
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
	return f.norddrop.Stop()
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

	transferDescriptors := make([]norddrop.TransferDescriptor, len(paths))

	for i, path := range paths {
		transferDescriptors[i] = norddrop.TransferDescriptorPath{Path: path}
	}

	transfer, err := f.norddrop.NewTransfer(peer.String(), transferDescriptors)
	if err != nil {
		return "", fmt.Errorf("transfer wasn't created")
	}

	return transfer, nil
}

// Accept starts downloading provided files into dstPath.
// dstPath must be absolute.
func (f *Fileshare) Accept(transferID, dstPath string, fileID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if err := f.norddrop.DownloadFile(transferID, fileID, dstPath); err != nil {
		return err
	}

	return nil
}

// Finalize file transfer.
func (f *Fileshare) Finalize(transferID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.norddrop.FinalizeTransfer(transferID)
}

// CancelFile id in a transfer
func (f *Fileshare) CancelFile(transferID string, fileID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if err := f.norddrop.RejectFile(transferID, fileID); err != nil {
		return err
	}

	return nil
}

// GetTransfersSince provided time from fileshare implementation storage
func (f *Fileshare) GetTransfersSince(t time.Time) ([]norddrop.TransferInfo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	since := t.Unix()
	transfers, err := f.norddrop.TransfersSince(since)
	if err != nil {
		return nil, fmt.Errorf("getting transfers since %d: %w", since, err)
	}

	return transfers, nil
}

// PurgeTransfersUntil provided time from fileshare implementation storage
func (f *Fileshare) PurgeTransfersUntil(until time.Time) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	// TODO: In the calculation below: `until.Unix() * 100` it should be
	// multiplied by 1000 to get number of milliseconds. The issue is that there
	// is a bug on the libdrop side here: https://github.com/NordSecurity/libdrop/blob/v7.0.0/norddrop/src/uni.rs#L100
	// It converts milliseconds to seconds by dividing by 100 instead of 1000
	// resulting in incorrect dates in the year ~2515 and purgin of all transfers.
	// This will be fixed with migration to v8.0.0 of libdrop.
	return f.norddrop.PurgeTransfersUntil(until.Unix() * 100)
}
