package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	// InstallFilePath defines filename of install id file
	InstallFilePath = filepath.Join(internal.DatFilesPathCommon, "install.dat")
	// SettingsDataFilePath defines path to app configs file
	SettingsDataFilePath = filepath.Join(internal.DatFilesPath, "settings.dat")
)

var errNoInstallFile = errors.New("install file doesn't exist")

// SaveFunc is used by Manager to save the config.
type SaveFunc func(Config) Config

func getCaller() string {
	// we need to skip two frames, one for getCaller, and one for caller of getCaller
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}

// Manager is responsible for persisting and retrieving the config.
type Manager interface {
	// SaveWith updates parts of the config specified by the SaveFunc.
	SaveWith(SaveFunc) error
	// Load config into a given struct.
	Load(*Config) error
	// Reset config to default values.
	Reset(preserveLoginData bool) error
}

type FilesystemHandle interface {
	FileExists(string) bool
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, fs.FileMode) error
}

type StdFilesystemHandle struct{}

func (StdFilesystemHandle) FileExists(location string) bool {
	return internal.FileExists(location)
}

func (StdFilesystemHandle) ReadFile(location string) ([]byte, error) {
	// #nosec G304 -- no input comes from the user
	return os.ReadFile(location)
}

func (StdFilesystemHandle) WriteFile(location string, data []byte, mode fs.FileMode) error {
	if err := internal.EnsureDir(location); err != nil {
		return err
	}
	return os.WriteFile(location, data, mode)
}

type DataConfigChange struct {
	Config *Config
	// Caller contains file and line number of the call to update the config
	Caller string
}

type ConfigPublisher interface {
	Publish(DataConfigChange)
}

// FilesystemConfigManager implements config persistence and retrieval from disk.
//
// Thread-safe.
type FilesystemConfigManager struct {
	location        string
	vault           string
	salt            string
	machineIDGetter MachineIDGetter
	fsHandle        FilesystemHandle
	NewInstallation bool
	configPublisher ConfigPublisher
	mu              sync.Mutex
}

// NewFilesystemConfigManager is constructed from a given location and salt.
func NewFilesystemConfigManager(location, vault, salt string,
	machineIDGetter MachineIDGetter,
	fsHandle FilesystemHandle,
	configPublisher ConfigPublisher,
) *FilesystemConfigManager {
	return &FilesystemConfigManager{
		location:        location,
		vault:           vault,
		salt:            salt,
		machineIDGetter: machineIDGetter,
		fsHandle:        fsHandle,
		configPublisher: configPublisher,
	}
}

// SaveWith modifications provided by fn.
//
// Thread-safe.
func (f *FilesystemConfigManager) SaveWith(fn SaveFunc) error {
	// We want to publish the setting changes after the config change mutex is unlocked. Otherwise it could cause a
	// deadlock when conifg change subscriber tries to read the config with the same manager when the change is
	// published. The assumption here is that publisher is protected with it's own lock.
	caller := getCaller()
	var c Config
	var err error
	defer func() {
		if err == nil && f.configPublisher != nil {
			f.configPublisher.Publish(DataConfigChange{
				Config: &c,
				Caller: caller,
			})
		}
	}()

	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.load(&c); err != nil {
		return err
	}

	c = fn(c)
	err = f.save(c)

	return err
}

func (f *FilesystemConfigManager) save(c Config) error {
	pass, err := f.getPassphrase()
	if err != nil {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	encrypted, err := internal.Encrypt(data, pass)
	if err != nil {
		return err
	}

	return f.fsHandle.WriteFile(f.location, encrypted, internal.PermUserRW)
}

// Reset config values to defaults.
//
// Thread-safe.
func (f *FilesystemConfigManager) Reset(preserveLoginData bool) (retErr error) {
	// We want to publish the setting changes after the config change mutex is unlocked. Otherwise it could cause a
	// deadlock when conifg change subscriber tries to read the config with the same manager when the change is
	// published. The assumption here is that publisher is protected with it's own lock.
	caller := getCaller()
	var c Config
	defer func() {
		if retErr == nil && f.configPublisher != nil {
			f.configPublisher.Publish(DataConfigChange{
				Config: &c,
				Caller: caller,
			})
		}
	}()

	if preserveLoginData {
		var cfg Config
		retErr = f.load(&cfg)
		if retErr != nil {
			return fmt.Errorf("loading old config: %w", retErr)
		}
		c = *newConfigWithLoginData(f.machineIDGetter, cfg)
		retErr = f.save(c)
		return
	}
	c = *newConfig(f.machineIDGetter)
	retErr = f.save(c)
	return
}

// Load encrypted config from the filesystem.
//
// Thread-safe.
func (f *FilesystemConfigManager) Load(c *Config) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.load(c)
}

func (f *FilesystemConfigManager) load(c *Config) error {
	// always init with default settings and override later with the values from the file
	*c = *newConfig(f.machineIDGetter)

	if !f.fsHandle.FileExists(f.location) {
		f.NewInstallation = true
		return f.save(*c)
	}

	pass, err := f.getPassphrase()
	if err != nil {
		return err
	}

	var data []byte

	// #nosec G304 -- no input comes from the user
	data, err = f.fsHandle.ReadFile(f.location)
	if err != nil {
		return err
	}

	decrypted, err := internal.Decrypt(data, pass)
	if err != nil {
		return err
	}

	// this overrides default values
	if err := json.Unmarshal(decrypted, c); err != nil {
		return err
	}

	return nil
}

// getPassphrase for accessing the data
func (f *FilesystemConfigManager) getPassphrase() (string, error) {
	key, err := f.loadKey()
	if err != nil {
		if errors.Is(err, errNoInstallFile) {
			err = f.newKey()
			if err != nil {
				return "", err
			}
			key, err = f.loadKey()
			if err != nil {
				return "", err
			}
			return string(key), nil
		}
		return "", err
	}
	return string(key), nil
}

// newKey used for decryption
func (f *FilesystemConfigManager) newKey() error {
	cipher, err := internal.Encrypt(generateKey(), f.salt)
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(cipher)
	if err != nil {
		return err
	}

	err = f.fsHandle.WriteFile(f.vault, buffer.Bytes(), internal.PermUserRW)
	if err != nil {
		return err
	}
	return nil
}

func generateKey() []byte {
	min, max := 33, 126
	key := make([]byte, 0, 32)
	source := rand.NewSource(time.Now().UnixNano())
	// #nosec G404 -- config encryption will go away after OSS
	r := rand.New(source)
	for i := 0; i < 32; i++ {
		character := r.Intn(max-min) + min
		key = append(key, byte(character))
	}
	return key
}

// loadKey for decryption from disk
func (f *FilesystemConfigManager) loadKey() ([]byte, error) {
	if !f.fsHandle.FileExists(f.vault) {
		return nil, errNoInstallFile
	}
	content, err := f.fsHandle.ReadFile(f.vault)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, errNoInstallFile
	}
	decoder := gob.NewDecoder(bytes.NewReader(content))
	var cipher []byte
	err = decoder.Decode(&cipher)
	if err != nil {
		return nil, err
	}
	plain, err := internal.Decrypt(cipher, f.salt)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
