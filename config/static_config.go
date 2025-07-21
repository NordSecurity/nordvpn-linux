package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// var staticConfigFilename = internal.DatFilesPathCommon + "install_static.dat"
var staticConfigFilename = filepath.Join(internal.DatFilesPathCommon, "install_static.dat")
var (
	ErrStaticValueAlreadySet    = errors.New("static value already configured")
	ErrStaticValueNotConfigured = errors.New("static value was not configured")
	ErrFailedToReadConfigFile   = errors.New("failed to read static config file")
)

type ConfigState int

const (
	NoFile ConfigState = iota
	FailedToInitialize
	Initialized
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
	state ConfigState
	cfg   StaticConfig
}

func tryInitStaticConfig(fs FilesystemHandle) (StaticConfig, ConfigState) {
	cfgFile, err := fs.ReadFile(staticConfigFilename)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to load static config:", err)
		if errors.Is(err, os.ErrNotExist) {
			return StaticConfig{}, NoFile
		}
		return StaticConfig{}, FailedToInitialize
	}

	var cfg StaticConfig
	err = json.Unmarshal(cfgFile, &cfg)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to unmarshal static config:", err)
		return cfg, FailedToInitialize
	}

	return cfg, Initialized
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
	if s.state == Initialized {
		return s.cfg, nil
	}

	cfg, state := tryInitStaticConfig(s.fs)
	if state == FailedToInitialize {
		return cfg, ErrFailedToReadConfigFile
	}
	s.cfg = cfg
	s.state = state

	return s.cfg, nil
}

func (s *FilesystemStaticConfigManager) GetRolloutGroup() (int, error) {
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
