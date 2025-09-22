package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

var (
	ErrStaticValueAlreadySet    = errors.New("static value already configured")
	ErrStaticValueNotConfigured = errors.New("static value was not configured")
	ErrRolloutGroupOutOfBounds  = errors.New("rollout group out of bounds")
)

const (
	rolloutGroupUnsetValue = 0
	rolloutGroupMin        = 1
	rolloutGroupMax        = 100
)

type StaticConfig struct {
	RolloutGroup int `json:"rollout_group"`
}

// StaticConfigManager stores values which remain constant throughout app's lifetime
type StaticConfigManager interface {
	GetRolloutGroup() (int, error)
	SetRolloutGroup(int) error
}

// FilesystemStaticConfigManager saves and reads values to a config file
type FilesystemStaticConfigManager struct {
	fs FilesystemHandle
	mu sync.RWMutex
}

func NewFilesystemStaticConfigManager() *FilesystemStaticConfigManager {
	return &FilesystemStaticConfigManager{
		fs: StdFilesystemHandle{},
	}
}

// loadConfig reads the config from disk
func (s *FilesystemStaticConfigManager) loadConfig() (sc StaticConfig, err error) {
	defer func() {
		if err != nil {
			// in case read/load fails, try to write default/empty values
			if data, err := json.Marshal(StaticConfig{}); err == nil {
				s.fs.WriteFile(internal.StaticConfigFilename, data, internal.PermUserRW)
			}
		}
	}()

	var data []byte
	data, err = s.fs.ReadFile(internal.StaticConfigFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file doesn't exist yet, return empty config
			return StaticConfig{}, nil
		}
		return StaticConfig{}, fmt.Errorf("reading static config file: %w", err)
	}

	var cfg StaticConfig
	if err = json.Unmarshal(data, &cfg); err != nil {
		return StaticConfig{}, fmt.Errorf("unmarshaling static config: %w", err)
	}

	return cfg, nil
}

// saveConfig writes the config to disk
func (s *FilesystemStaticConfigManager) saveConfig(cfg StaticConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling static config: %w", err)
	}

	if err := s.fs.WriteFile(internal.StaticConfigFilename, data, internal.PermUserRW); err != nil {
		return fmt.Errorf("writing static config file: %w", err)
	}

	return nil
}

func (s *FilesystemStaticConfigManager) GetRolloutGroup() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.loadConfig()
	if err != nil {
		return 0, fmt.Errorf("getting rollout group: %w", err)
	}

	if cfg.RolloutGroup == rolloutGroupUnsetValue {
		return 0, ErrStaticValueNotConfigured
	}

	return cfg.RolloutGroup, nil
}

func (s *FilesystemStaticConfigManager) SetRolloutGroup(rolloutGroup int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rolloutGroup < rolloutGroupMin || rolloutGroup > rolloutGroupMax {
		return ErrRolloutGroupOutOfBounds
	}

	cfg, err := s.loadConfig()
	if err != nil {
		return fmt.Errorf("setting rollout group: %w", err)
	}

	if cfg.RolloutGroup != rolloutGroupUnsetValue {
		return ErrStaticValueAlreadySet
	}

	cfg.RolloutGroup = rolloutGroup
	if err = s.saveConfig(cfg); err != nil {
		return fmt.Errorf("saving rollout group: %w", err)
	}

	return nil
}
