package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

var staticConfigFilename = filepath.Join(internal.DatFilesPathCommon, "install_static.dat")
var (
	ErrStaticValueAlreadySet    = errors.New("static value already configured")
	ErrStaticValueNotConfigured = errors.New("static value was not configured")
	ErrFailedToReadConfigFile   = errors.New("failed to read static config file")
	ErrRolloutGroupOutOfBounds  = errors.New("rollout group out of bounds")
)

type configState int

const (
	staticConfigState_noFile configState = iota
	staticConfigState_failedToInitialize
	staticConfigState_initialized
)

const rolloutGroupUnsetValue = 0

type StaticConfig struct {
	RolloutGroup int `json:"rollout_group,omitempty"`
}

// StaticConfigManager stores values which remain constant thoroughout app's lifetime
type StaticConfigManager interface {
	GetRolloutGroup() (int, error)
	SetRolloutGroup(int) error
}

// FilesystemStaticConfigManager saves and reads values to a config file
type FilesystemStaticConfigManager struct {
	fs    FilesystemHandle
	state configState
	cfg   StaticConfig
	mu    sync.RWMutex
}

func tryInitStaticConfig(fs FilesystemHandle) (StaticConfig, configState) {
	cfgFile, err := fs.ReadFile(staticConfigFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// file not existing is normal on first start, dont log as error
			return StaticConfig{}, staticConfigState_noFile
		}
		log.Println(internal.ErrorPrefix, "failed to load static config:", err)
		return StaticConfig{}, staticConfigState_failedToInitialize
	}

	var cfg StaticConfig
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to unmarshal static config:", err)
		return cfg, staticConfigState_failedToInitialize
	}

	return cfg, staticConfigState_initialized
}

func NewFilesystemStaticConfigManager() *FilesystemStaticConfigManager {
	fs := StdFilesystemHandle{}
	cfg, state := tryInitStaticConfig(fs)

	return &FilesystemStaticConfigManager{
		fs:    fs,
		state: state,
		cfg:   cfg,
	}
}

func (s *FilesystemStaticConfigManager) getConfig() (StaticConfig, error) {
	if s.state == staticConfigState_initialized {
		return s.cfg, nil
	}

	if s.state == staticConfigState_failedToInitialize {
		return s.cfg, ErrFailedToReadConfigFile
	}

	if s.state == staticConfigState_noFile {
		cfg, state := tryInitStaticConfig(s.fs)
		s.cfg = cfg
		s.state = state

		if state == staticConfigState_failedToInitialize {
			return cfg, ErrFailedToReadConfigFile
		}
	}

	return s.cfg, nil
}

func (s *FilesystemStaticConfigManager) GetRolloutGroup() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := s.getConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to read config: %w", err)
	}

	if cfg.RolloutGroup == rolloutGroupUnsetValue {
		return 0, ErrStaticValueNotConfigured
	}

	return cfg.RolloutGroup, nil
}

func (s *FilesystemStaticConfigManager) SetRolloutGroup(rolloutGroup int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	const rolloutGroupMin = 1
	const rolloutGroupMax = 100

	if rolloutGroup < rolloutGroupMin || rolloutGroup > rolloutGroupMax {
		return ErrRolloutGroupOutOfBounds
	}

	cfg, err := s.getConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if cfg.RolloutGroup != rolloutGroupUnsetValue {
		return ErrStaticValueAlreadySet
	}

	cfg.RolloutGroup = rolloutGroup
	json, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize json: %w", err)
	}

	err = s.fs.WriteFile(staticConfigFilename, json, internal.PermUserRW)
	if err != nil {
		return fmt.Errorf("failed to save static config file: %w", err)
	}

	s.cfg = cfg
	s.state = staticConfigState_initialized
	return nil
}
