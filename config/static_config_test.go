package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"gotest.tools/v3/assert"
)

func TestTryInitConfigState(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name          string
		expectedState configState
		filename      string
		fileContents  []byte
		readErr       error
	}{
		{
			name:          "init success",
			expectedState: initialized,
			filename:      staticConfigFilename,
			fileContents:  []byte("{}"),
		},
		{
			name:          "no file",
			expectedState: noFile,
			readErr:       os.ErrNotExist,
		},
		{
			name:          "invalid json",
			expectedState: failedToInitialize,
			filename:      staticConfigFilename,
			fileContents:  []byte("{"),
		},
		{
			name:          "failed to read file",
			expectedState: failedToInitialize,
			readErr:       os.ErrPermission,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := config.NewFilesystemMock(t)
			fsMock.ReadErr = test.readErr
			fsMock.AddFile(test.filename, test.fileContents)
			_, state := tryInitStaticConfig(&fsMock)
			assert.Equal(t, test.expectedState, state, "Unexpected state returned after config initialization")
		})
	}
}

func TestGetRolloutGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		currentConfigState   configState
		currentConfig        StaticConfig
		readErr              error
		expectedRolloutGroup int
		expectedErr          error
	}{
		{
			name:               "config initialized, rollout group is configured",
			currentConfigState: initialized,
			currentConfig: StaticConfig{
				RolloutGroup: 1,
			},
			expectedRolloutGroup: 1,
			expectedErr:          nil,
		},
		{
			name:               "config initialized, rollout group is not configured",
			currentConfigState: initialized,
			currentConfig: StaticConfig{
				RolloutGroup: 0,
			},
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "config not initialized",
			currentConfigState:   noFile,
			currentConfig:        StaticConfig{},
			readErr:              os.ErrNotExist,
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "failed to read config file",
			currentConfigState:   noFile,
			currentConfig:        StaticConfig{},
			readErr:              os.ErrPermission,
			expectedRolloutGroup: 0,
			expectedErr:          ErrFailedToReadConfigFile,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := config.NewFilesystemMock(t)
			fsMock.ReadErr = test.readErr
			manager := FilesystemStaticConfigManager{
				fs:    &fsMock,
				state: test.currentConfigState,
				cfg:   test.currentConfig,
			}

			rolloutGroup, err := manager.GetRolloutGroup()
			assert.Equal(t, test.expectedRolloutGroup, rolloutGroup, "Unexpected rollout group value.")
			assert.ErrorIs(t, test.expectedErr, err)
		})
	}
}

func TestSetRolloutGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		currentConfig        StaticConfig
		currentConfigState   configState
		targetRolloutGroup   int
		readErr              error
		writeErr             error
		expectedErr          error
		expectedRolloutGroup int
	}{
		{
			name: "config initialized, rollout group not initialized",
			currentConfig: StaticConfig{
				RolloutGroup: 0,
			},
			currentConfigState:   initialized,
			targetRolloutGroup:   1,
			expectedErr:          nil,
			expectedRolloutGroup: 1,
		},
		{
			name: "config initialized, rollout group is initialized",
			currentConfig: StaticConfig{
				RolloutGroup: 20,
			},
			currentConfigState:   initialized,
			targetRolloutGroup:   1,
			expectedErr:          ErrStaticValueAlreadySet,
			expectedRolloutGroup: 20,
		},
		{
			name:                 "config not initialized",
			currentConfig:        StaticConfig{},
			currentConfigState:   noFile,
			targetRolloutGroup:   20,
			expectedErr:          nil,
			expectedRolloutGroup: 20,
		},
		{
			name:                 "config not initialized",
			currentConfig:        StaticConfig{},
			currentConfigState:   noFile,
			targetRolloutGroup:   20,
			expectedErr:          nil,
			expectedRolloutGroup: 20,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfgJson, _ := json.Marshal(test.currentConfig)

			fsMock := config.NewFilesystemMock(t)
			fsMock.AddFile(staticConfigFilename, cfgJson)
			fsMock.ReadErr = test.readErr
			fsMock.WriteErr = test.writeErr

			manager := FilesystemStaticConfigManager{
				fs:    &fsMock,
				state: test.currentConfigState,
				cfg:   test.currentConfig,
			}

			err := manager.SetRolloutGroup(test.targetRolloutGroup)
			assert.ErrorIs(t, err, test.expectedErr, "Unexpected error returned after setting the rollout group.")
			assert.Equal(t, test.expectedRolloutGroup, manager.cfg.RolloutGroup,
				"Invalid rollout group saved in config.")

			cfgJson, _ = fsMock.ReadFile(staticConfigFilename)
			var cfg StaticConfig
			json.Unmarshal(cfgJson, &cfg)

			assert.Equal(t, test.expectedRolloutGroup, cfg.RolloutGroup,
				"Rollout group was not saved in the config file.")
		})
	}
}
