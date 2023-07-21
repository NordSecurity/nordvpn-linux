package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"

	mapset "github.com/deckarep/golang-set"
)

// Manager is responsible for loading and saving configurations
type Manager interface {
	Save(config Config) error
	Load() (Config, error)
}

// EncryptedManager is an implementation of configuration manager by using encrypted files
type EncryptedManager struct {
	filePath string
	uid      int
	gid      int
	salt     string
}

// NewEncryptedManager is a default constructor for EncryptedManager
func NewEncryptedManager(filePath string, uid int, gid int, salt string) EncryptedManager {
	if uid == 0 {
		uid = os.Getuid()
	}
	if gid == 0 {
		gid = os.Getgid()
	}
	return EncryptedManager{
		filePath: filePath,
		uid:      uid,
		gid:      gid,
		salt:     salt,
	}
}

// Load app configuration from file
func (m EncryptedManager) Load() (Config, error) {
	file, err := retrieveFile(m.filePath, m.uid, m.gid, false, m.salt)
	if err != nil {
		return Config{}, fmt.Errorf("retrieving config file: %w", err)
	}
	// https://www.joeshaw.org/dont-defer-close-on-writable-files/
	// #nosec G307 -- file is opened for only reading
	defer file.Close()

	encryptedData, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("reading config file content: %w", err)
	}

	config, err := parseConfig(encryptedData, m.uid, m.salt)
	if err != nil {
		return Config{}, fmt.Errorf("parsing config file: %w", err)
	}

	config.setDefaultsIfEmpty()
	return config, nil
}

// parseConfig decrypts the provided content and parses it to config structure
func parseConfig(encData []byte, uid int, salt string) (Config, error) {
	plainData, err := internal.Decrypt(encData, getPassphrase(uid, salt))
	if err != nil {
		return Config{}, fmt.Errorf("decrypting configuration: %w", err)
	}

	var config Config
	err = json.Unmarshal(plainData, &config)
	if err != nil {
		return Config{}, fmt.Errorf("parsing configuration from JSON: %w", err)
	}

	return config, nil
}

// retrieveFile returns existing config file if it exists. If it does not exist, it creates a new one and
// returns it
func retrieveFile(filePath string, uid int, gid int, save bool, salt string) (*os.File, error) {
	var configFile *os.File
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		configFile, err = createFile(filePath, uid, gid, salt)
		if err != nil {
			return nil, fmt.Errorf("creating new configuration file: %w", err)
		}
	} else {
		flag := os.O_RDWR
		if save {
			flag = flag | os.O_TRUNC
		}
		// #nosec G304 -- no input comes from the user
		configFile, err = os.OpenFile(filePath, flag, internal.PermUserRW)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
	}
	return configFile, nil
}

// createFile creates a file to filePath
func createFile(filePath string, uid int, gid int, salt string) (*os.File, error) {
	file, err := internal.FileCreateForUser(filePath, internal.PermUserRW, uid, gid)
	if err != nil {
		return nil, fmt.Errorf("creating a file: %w", err)
	}
	jsonData, err := json.Marshal(NewConfig())
	if err != nil {
		return nil, fmt.Errorf("marshaling configuration to a JSON structure: %w", err)
	}
	bytes, err := internal.Encrypt(jsonData, getPassphrase(uid, salt))
	if err != nil {
		return nil, fmt.Errorf("encrypting configuration file: %w", err)
	}
	_, err = file.Write(bytes)
	if err != nil {
		return nil, fmt.Errorf("writing file content: %w", err)
	}
	// return pointer to the beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("returning reader pointer to beginning: %w", err)
	}
	return file, nil
}

// Save specified configuration to a file in filePath
func (m EncryptedManager) Save(c Config) error {
	configFile, err := retrieveFile(m.filePath, m.uid, m.gid, true, m.salt)
	if err != nil {
		return fmt.Errorf("retrieving configuration file: %w", err)
	}
	jsonData, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling configuration to JSON: %w", err)
	}
	bytes, err := internal.Encrypt(jsonData, getPassphrase(m.uid, m.salt))
	if err != nil {
		return fmt.Errorf("encrypting configuration: %w", err)
	}
	_, err = configFile.Write(bytes)
	if err != nil {
		return fmt.Errorf("writing to configuration file: %w", err)
	}
	err = configFile.Close()
	if err != nil {
		return fmt.Errorf("closing configuration file: %w", err)
	}
	return nil
}

// setDefaultsIfEmpty sets default values
func (c *Config) setDefaultsIfEmpty() *Config {
	if c.Allowlist.Subnets == nil {
		c.Allowlist.Subnets = mapset.NewSet()
	}
	if c.Allowlist.Ports.UDP == nil {
		c.Allowlist.Ports.UDP = mapset.NewSet()
	}
	if c.Allowlist.Ports.TCP == nil {
		c.Allowlist.Ports.TCP = mapset.NewSet()
	}
	if c.Technology == config.Technology_UNKNOWN_TECHNOLOGY {
		c.Technology = config.Technology_NORDLYNX
	}
	return c
}

func getPassphrase(uid int, salt string) string {
	return salt + strconv.Itoa(uid)
}
