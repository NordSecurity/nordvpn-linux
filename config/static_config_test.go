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
			expectedState: staticConfigState_initialized,
			filename:      staticConfigFilename,
			fileContents:  []byte("{}"),
		},
		{
			name:          "no file",
			expectedState: staticConfigState_noFile,
			readErr:       os.ErrNotExist,
		},
		{
			name:          "invalid json",
			expectedState: staticConfigState_failedToInitialize,
			filename:      staticConfigFilename,
			fileContents:  []byte("{"),
		},
		{
			name:          "failed to read file",
			expectedState: staticConfigState_failedToInitialize,
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
			currentConfigState: staticConfigState_initialized,
			currentConfig: StaticConfig{
				RolloutGroup: 1,
			},
			expectedRolloutGroup: 1,
			expectedErr:          nil,
		},
		{
			name:               "config initialized, rollout group is not configured",
			currentConfigState: staticConfigState_initialized,
			currentConfig: StaticConfig{
				RolloutGroup: 0,
			},
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "config not initialized",
			currentConfigState:   staticConfigState_noFile,
			currentConfig:        StaticConfig{},
			readErr:              os.ErrNotExist,
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "failed to read config file",
			currentConfigState:   staticConfigState_noFile,
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
			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			}
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
			currentConfigState:   staticConfigState_initialized,
			targetRolloutGroup:   1,
			expectedErr:          nil,
			expectedRolloutGroup: 1,
		},
		{
			name: "config initialized, rollout group is initialized",
			currentConfig: StaticConfig{
				RolloutGroup: 20,
			},
			currentConfigState:   staticConfigState_initialized,
			targetRolloutGroup:   1,
			expectedErr:          ErrStaticValueAlreadySet,
			expectedRolloutGroup: 20,
		},
		{
			name:                 "config not initialized",
			currentConfig:        StaticConfig{},
			currentConfigState:   staticConfigState_noFile,
			targetRolloutGroup:   20,
			expectedErr:          nil,
			expectedRolloutGroup: 20,
		},
		{
			name:                 "config not initialized",
			currentConfig:        StaticConfig{},
			currentConfigState:   staticConfigState_noFile,
			targetRolloutGroup:   20,
			expectedErr:          nil,
			expectedRolloutGroup: 20,
		},
		{
			name: "rollout group is out of bounds upper",
			currentConfig: StaticConfig{
				RolloutGroup: 0,
			},
			currentConfigState:   staticConfigState_initialized,
			targetRolloutGroup:   101,
			expectedErr:          ErrRolloutGroupOutOfBounds,
			expectedRolloutGroup: 0,
		},
		{
			name: "rollout group is out of bounds lower",
			currentConfig: StaticConfig{
				RolloutGroup: 0,
			},
			currentConfigState:   staticConfigState_initialized,
			targetRolloutGroup:   0,
			expectedErr:          ErrRolloutGroupOutOfBounds,
			expectedRolloutGroup: 0,
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

// countingFilesystem wraps a FilesystemHandle to count method calls
type countingFilesystem struct {
	FilesystemHandle
	readCallCount int
}

func (c *countingFilesystem) ReadFile(location string) ([]byte, error) {
	c.readCallCount++
	return c.FilesystemHandle.ReadFile(location)
}

func TestGetConfigInitializesOnlyOnce(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name               string
		initialState       configState
		initialConfig      StaticConfig
		readCallsExpected  int
		expectedFinalState configState
		readErr            error
	}{
		{
			name:               "already initialized - should not read file",
			initialState:       staticConfigState_initialized,
			initialConfig:      StaticConfig{RolloutGroup: 42},
			readCallsExpected:  0,
			expectedFinalState: staticConfigState_initialized,
		},
		{
			name:               "failed state - should not retry",
			initialState:       staticConfigState_failedToInitialize,
			initialConfig:      StaticConfig{},
			readCallsExpected:  0,
			expectedFinalState: staticConfigState_failedToInitialize,
		},
		{
			name:               "no file state - should try once and succeed",
			initialState:       staticConfigState_noFile,
			initialConfig:      StaticConfig{},
			readCallsExpected:  1,
			expectedFinalState: staticConfigState_initialized,
		},
		{
			name:               "no file state - should try once and fail",
			initialState:       staticConfigState_noFile,
			initialConfig:      StaticConfig{},
			readCallsExpected:  1,
			expectedFinalState: staticConfigState_failedToInitialize,
			readErr:            os.ErrPermission,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := config.NewFilesystemMock(t)
			fsMock.ReadErr = test.readErr
			if test.readErr == nil {
				fsMock.AddFile(staticConfigFilename, []byte(`{"rollout_group": 99}`))
			}

			countingFS := &countingFilesystem{
				FilesystemHandle: &fsMock,
			}

			manager := &FilesystemStaticConfigManager{
				fs:    countingFS,
				state: test.initialState,
				cfg:   test.initialConfig,
			}

			// Call getConfig multiple times
			for i := 0; i < 3; i++ {
				cfg, err := manager.getConfig()

				if test.readErr != nil && test.initialState == staticConfigState_noFile {
					assert.ErrorIs(t, err, ErrFailedToReadConfigFile)
				} else if test.initialState == staticConfigState_failedToInitialize {
					assert.ErrorIs(t, err, ErrFailedToReadConfigFile)
				} else {
					assert.NilError(t, err)
				}

				// Verify config is returned correctly
				if test.initialState == staticConfigState_initialized {
					assert.Equal(t, test.initialConfig.RolloutGroup, cfg.RolloutGroup)
				} else if test.readErr == nil && test.initialState == staticConfigState_noFile {
					assert.Equal(t, 99, cfg.RolloutGroup)
				}
			}

			// Verify read was called expected number of times
			assert.Equal(t, test.readCallsExpected, countingFS.readCallCount,
				"Expected %d read calls but got %d", test.readCallsExpected, countingFS.readCallCount)

			// Verify final state
			assert.Equal(t, test.expectedFinalState, manager.state,
				"Expected final state %v but got %v", test.expectedFinalState, manager.state)
		})
	}
}
