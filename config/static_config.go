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
	noFile configState = iota
	failedToInitialize
	initialized
)

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
	mu    sync.Mutex
}

func tryInitStaticConfig(fs FilesystemHandle) (StaticConfig, configState) {
	cfgFile, err := fs.ReadFile(staticConfigFilename)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to load static config:", err)
		if errors.Is(err, os.ErrNotExist) {
			return StaticConfig{}, noFile
		}
		return StaticConfig{}, failedToInitialize
	}

	var cfg StaticConfig
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to unmarshal static config:", err)
		return cfg, failedToInitialize
	}

	return cfg, initialized
}

func NewFilesystemStaticConfigManager() FilesystemStaticConfigManager {
	fs := StdFilesystemHandle{}
	cfg, state := tryInitStaticConfig(fs)

	return FilesystemStaticConfigManager{
		fs:    fs,
		state: state,
		cfg:   cfg,
	}
}

func (s *FilesystemStaticConfigManager) getConfig() (StaticConfig, error) {
	if s.state == initialized {
		return s.cfg, nil
	}

	cfg, state := tryInitStaticConfig(s.fs)
	if state == failedToInitialize {
		return cfg, ErrFailedToReadConfigFile
	}
	s.cfg = cfg
	s.state = state

	return s.cfg, nil
}

func (s *FilesystemStaticConfigManager) GetRolloutGroup() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.getConfig()
	if err != nil {
		return 0, err
	}

	if cfg.RolloutGroup == 0 {
		return 0, ErrStaticValueNotConfigured
	}

	return cfg.RolloutGroup, nil
}

func (s *FilesystemStaticConfigManager) SetRolloutGroup(rolloutGroup int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rolloutGroup < 1 || rolloutGroup > 100 {
		return ErrRolloutGroupOutOfBounds
	}

	cfg, err := s.getConfig()
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if cfg.RolloutGroup != 0 {
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
	return nil
}
