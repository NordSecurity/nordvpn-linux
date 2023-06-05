package config

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	// InstallFilePath defines filename of install id file
	InstallFilePath = internal.DatFilesPath + "install.dat"
	// SettingsDataFilePath defines path to app configs file
	SettingsDataFilePath = internal.DatFilesPath + "settings.dat"
)

var errNoInstallFile = errors.New("install file doesn't exist")

// SaveFunc is used by Manager to save the config.
type SaveFunc func(Config) Config

// Manager is responsible for persisting and retrieving the config.
type Manager interface {
	// SaveWith updates parts of the config specified by the SaveFunc.
	SaveWith(SaveFunc) error
	// Load config into a given struct.
	Load(*Config) error
	// Reset config to default values.
	Reset() error
}

// Filesystem implements config persistence and retrieval from disk.
//
// Thread-safe.
type Filesystem struct {
	location string
	vault    string
	salt     string
	mu       sync.Mutex
}

// NewFilesystem is constructed from a given location and salt.
func NewFilesystem(location, vault, salt string) *Filesystem {
	return &Filesystem{
		location: location,
		vault:    vault,
		salt:     salt,
	}
}

// SaveWith modifications provided by fn.
//
// Thread-safe.
func (f *Filesystem) SaveWith(fn SaveFunc) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var c Config
	if err := f.load(&c); err != nil {
		return err
	}

	c = fn(c)
	return f.save(c)
}

func (f *Filesystem) save(c Config) error {
	pass, err := getPassphrase(f.vault, f.salt)
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
	return save(encrypted, f.location)
}

// Reset config values to defaults.
//
// Thread-safe.
func (f *Filesystem) Reset() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.save(*newConfig())
}

// Load encrypted config from the filesystem.
//
// Thread-safe.
func (f *Filesystem) Load(c *Config) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.load(c)
}

func (f *Filesystem) load(c *Config) error {
	if !internal.FileExists(f.location) {
		// reasigning value behind the pointer
		*c = *newConfig()
		return nil
	}

	pass, err := getPassphrase(f.vault, f.salt)
	if err != nil {
		return err
	}

	data, err := load(f.location)
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

	if c.FirewallMark == 0 {
		// without this users need to reset their configs
		c.FirewallMark = defaultFWMarkValue
	}

	if c.TokensData == nil {
		// in order not to crash when a user has an old config
		c.TokensData = map[int64]TokenData{}
	}

	if c.MachineID == [16]byte{} {
		c.MachineID = internal.MachineID()
	}
	return nil
}

// save data to given location on disk
func save(data []byte, location string) error {
	if internal.FileExists(location) {
		return os.WriteFile(location, data, internal.PermUserRW)
	}

	file, err := internal.FileCreate(location, internal.PermUserRW)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		// https://www.joeshaw.org/dont-defer-close-on-writable-files/
		// #nosec G104 -- errors.Join would be useful here
		file.Close()
		return err
	}
	return file.Close()
}

// load data from given location on disk
func load(location string) ([]byte, error) {
	if internal.FileExists(location) {
		// #nosec G304 -- no input comes from the user
		return os.ReadFile(location)
	}

	file, err := internal.FileCreate(location, internal.PermUserRW)
	if err != nil {
		return nil, err
	}
	// https://www.joeshaw.org/dont-defer-close-on-writable-files/
	// #nosec G307 -- no writes are made
	defer file.Close()

	return []byte{}, nil
}

// getPassphrase for accessing the data
func getPassphrase(vault string, salt string) (string, error) {
	key, err := loadKey(vault, salt)
	if err != nil {
		if errors.Is(err, errNoInstallFile) {
			err = newKey(vault, salt)
			if err != nil {
				return "", err
			}
			key, err = loadKey(vault, salt)
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
func newKey(vault string, salt string) error {
	cipher, err := internal.Encrypt(generateKey(), salt)
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	encoder := gob.NewEncoder(buffer)
	err = encoder.Encode(cipher)
	if err != nil {
		return err
	}

	err = internal.FileWrite(vault, buffer.Bytes(), internal.PermUserRW)
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
func loadKey(vault string, salt string) ([]byte, error) {
	if !internal.FileExists(vault) {
		return nil, errNoInstallFile
	}
	content, err := internal.FileRead(vault)
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
	plain, err := internal.Decrypt(cipher, salt)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
