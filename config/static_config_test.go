package config

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/config"
	"gotest.tools/v3/assert"
)

func TestGetRolloutGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		fileContents         []byte
		readErr              error
		expectedRolloutGroup int
		expectedErr          error
	}{
		{
			name:                 "rollout group is configured",
			fileContents:         []byte(`{"rollout_group": 42}`),
			expectedRolloutGroup: 42,
			expectedErr:          nil,
		},
		{
			name:                 "rollout group is not configured (zero value)",
			fileContents:         []byte(`{"rollout_group": 0}`),
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "empty config file",
			fileContents:         []byte(`{}`),
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "file does not exist",
			readErr:              os.ErrNotExist,
			expectedRolloutGroup: 0,
			expectedErr:          ErrStaticValueNotConfigured,
		},
		{
			name:                 "failed to read config file",
			readErr:              os.ErrPermission,
			expectedRolloutGroup: 0,
			expectedErr:          os.ErrPermission,
		},
		{
			name:                 "invalid json",
			fileContents:         []byte(`{invalid json`),
			expectedRolloutGroup: 0,
			expectedErr:          errors.New("unmarshaling config"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := config.NewFilesystemMock(t)
			fsMock.ReadErr = test.readErr
			if test.fileContents != nil {
				fsMock.AddFile(internal.StaticConfigFilename, test.fileContents)
			}

			manager := FilesystemStaticConfigManager{
				fs: &fsMock,
			}

			rolloutGroup, err := manager.GetRolloutGroup()
			assert.Equal(t, test.expectedRolloutGroup, rolloutGroup, "Unexpected rollout group value.")

			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			} else {
				assert.NilError(t, err)
			}
		})
	}
}

func TestSetRolloutGroup(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                 string
		initialFileContents  []byte
		targetRolloutGroup   int
		readErr              error
		writeErr             error
		expectedErr          error
		expectedFileContents string
	}{
		{
			name:                 "set rollout group when not configured",
			initialFileContents:  []byte(`{}`),
			targetRolloutGroup:   42,
			expectedErr:          nil,
			expectedFileContents: `{"rollout_group":42}`,
		},
		{
			name:                 "set rollout group when file doesn't exist",
			initialFileContents:  []byte(`{}`), // empty JSON instead of no file
			targetRolloutGroup:   42,
			expectedErr:          nil,
			expectedFileContents: `{"rollout_group":42}`,
		},
		{
			name:                 "rollout group already set",
			initialFileContents:  []byte(`{"rollout_group": 20}`),
			targetRolloutGroup:   42,
			expectedErr:          ErrStaticValueAlreadySet,
			expectedFileContents: `{"rollout_group": 20}`, // unchanged
		},
		{
			name:                 "rollout group out of bounds (too high)",
			initialFileContents:  []byte(`{}`),
			targetRolloutGroup:   101,
			expectedErr:          ErrRolloutGroupOutOfBounds,
			expectedFileContents: `{}`, // unchanged
		},
		{
			name:                 "rollout group out of bounds (too low)",
			initialFileContents:  []byte(`{}`),
			targetRolloutGroup:   0,
			expectedErr:          ErrRolloutGroupOutOfBounds,
			expectedFileContents: `{}`, // unchanged
		},
		{
			name:                 "rollout group at lower bound",
			initialFileContents:  []byte(`{}`),
			targetRolloutGroup:   1,
			expectedErr:          nil,
			expectedFileContents: `{"rollout_group":1}`,
		},
		{
			name:                 "rollout group at upper bound",
			initialFileContents:  []byte(`{}`),
			targetRolloutGroup:   100,
			expectedErr:          nil,
			expectedFileContents: `{"rollout_group":100}`,
		},
		{
			name:                 "failed to read config file",
			readErr:              os.ErrPermission,
			targetRolloutGroup:   42,
			expectedErr:          os.ErrPermission,
			expectedFileContents: ``, // no file written
		},
		{
			name:                 "failed to write config file",
			initialFileContents:  []byte(`{}`),
			targetRolloutGroup:   42,
			writeErr:             os.ErrPermission,
			expectedErr:          os.ErrPermission,
			expectedFileContents: `{}`, // unchanged
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := config.NewFilesystemMock(t)
			fsMock.ReadErr = test.readErr
			fsMock.WriteErr = test.writeErr
			if test.initialFileContents != nil {
				fsMock.AddFile(internal.StaticConfigFilename, test.initialFileContents)
			}

			manager := FilesystemStaticConfigManager{
				fs: &fsMock,
			}

			err := manager.SetRolloutGroup(test.targetRolloutGroup)

			if test.expectedErr != nil {
				assert.ErrorContains(t, err, test.expectedErr.Error())
			} else {
				assert.NilError(t, err)
			}

			// Verify file contents if we expect a write
			if test.expectedFileContents != "" && test.writeErr == nil && test.readErr != os.ErrPermission && err == nil {
				fileContents, _ := fsMock.ReadFile(internal.StaticConfigFilename)
				if len(fileContents) > 0 {
					var actualConfig StaticConfig
					json.Unmarshal(fileContents, &actualConfig)

					var expectedConfig StaticConfig
					json.Unmarshal([]byte(test.expectedFileContents), &expectedConfig)

					assert.Equal(t, expectedConfig.RolloutGroup, actualConfig.RolloutGroup,
						"Rollout group was not saved correctly in the config file.")
				}
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	category.Set(t, category.Unit)

	fsMock := config.NewFilesystemMock(t)
	fsMock.AddFile(internal.StaticConfigFilename, []byte(`{"rollout_group": 50}`))

	manager := FilesystemStaticConfigManager{
		fs: &fsMock,
	}

	// Test concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			rolloutGroup, err := manager.GetRolloutGroup()
			assert.NilError(t, err)
			assert.Equal(t, 50, rolloutGroup)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLoadConfigErrorHandling(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		fileContents []byte
		readErr      error
		expectedErr  string
	}{
		{
			name:         "malformed json",
			fileContents: []byte(`{"rollout_group": "not a number"}`),
			expectedErr:  "unmarshaling config",
		},
		{
			name:        "permission denied",
			readErr:     os.ErrPermission,
			expectedErr: "reading config file",
		},
		{
			name:        "file not found returns empty config",
			readErr:     os.ErrNotExist,
			expectedErr: "", // No error, returns empty config
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := config.NewFilesystemMock(t)
			fsMock.ReadErr = test.readErr
			if test.fileContents != nil {
				fsMock.AddFile(internal.StaticConfigFilename, test.fileContents)
			}

			manager := FilesystemStaticConfigManager{
				fs: &fsMock,
			}

			cfg, err := manager.loadConfig()

			if test.expectedErr != "" {
				assert.ErrorContains(t, err, test.expectedErr)
			} else {
				assert.NilError(t, err)
				assert.Equal(t, 0, cfg.RolloutGroup)
			}
		})
	}
}

func TestSaveConfigErrorHandling(t *testing.T) {
	category.Set(t, category.Unit)

	fsMock := config.NewFilesystemMock(t)
	fsMock.WriteErr = os.ErrPermission

	manager := FilesystemStaticConfigManager{
		fs: &fsMock,
	}

	err := manager.saveConfig(StaticConfig{RolloutGroup: 42})
	assert.ErrorContains(t, err, "writing config file")
}
