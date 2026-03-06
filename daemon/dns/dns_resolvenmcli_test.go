package dns

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	"github.com/stretchr/testify/assert"
)

func generateConfig(t *testing.T, servers ...string) string {
	t.Helper()

	configTemplate := `[global-dns-domain-*]

servers=%s`

	return fmt.Sprintf(configTemplate, strings.Join(servers, ","))
}

func Test_NMCliSetUnset(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name                      string
		dnsServers                []string
		otherConfigFiles          []string
		shouldSetFail             bool
		shouldUnsetFail           bool
		expectedFileContents      string
		writeErr                  error
		removeErr                 error
		nmcliSetConfigReloadErr   error
		nmcliUnsetConfigReloadErr error
		getConfigFilesErr         error
	}{
		{
			name:                 "success",
			dnsServers:           []string{"1.1.1.1"},
			expectedFileContents: generateConfig(t, "1.1.1.1"),
			shouldSetFail:        false,
			shouldUnsetFail:      false,
		},
		{
			name:                 "success multiple servers",
			dnsServers:           []string{"1.1.1.1", "8.8.8.8"},
			expectedFileContents: generateConfig(t, "1.1.1.1", "8.8.8.8"),
			shouldSetFail:        false,
			shouldUnsetFail:      false,
		},
		{
			name:                 "other config files in directory are of lower priority",
			dnsServers:           []string{"1.1.1.1"},
			otherConfigFiles:     []string{"z-other-file.conf", "99-nordvpn-dns.conf"},
			expectedFileContents: generateConfig(t, "1.1.1.1"),
			shouldSetFail:        false,
			shouldUnsetFail:      false,
		},
		{
			name:            "writing to file fails",
			dnsServers:      []string{"1.1.1.1"},
			shouldSetFail:   true,
			shouldUnsetFail: true,
			writeErr:        errors.New("file write failed"),
			removeErr:       errors.New("config file doesn't exist"),
		},
		{
			name:                    "nmcli config reload fails",
			dnsServers:              []string{"1.1.1.1"},
			shouldSetFail:           true,
			shouldUnsetFail:         true,
			nmcliSetConfigReloadErr: fmt.Errorf("failed to reload config"),
			removeErr:               errors.New("config file doesn't exist"),
		},
		{
			name:                 "removing config file fails",
			dnsServers:           []string{"1.1.1.1"},
			expectedFileContents: generateConfig(t, "1.1.1.1"),
			shouldSetFail:        false,
			shouldUnsetFail:      true,
			removeErr:            errors.New("failed to remove the file"),
		},
		{
			name:                      "nmcli config reload fails on unset",
			dnsServers:                []string{"1.1.1.1"},
			expectedFileContents:      generateConfig(t, "1.1.1.1"),
			shouldSetFail:             false,
			shouldUnsetFail:           true,
			nmcliUnsetConfigReloadErr: errors.New("failed to reload config file"),
		},
		{
			name:                 "other config files have higher priority",
			dnsServers:           []string{"1.1.1.1"},
			expectedFileContents: generateConfig(t, "1.1.1.1"),
			otherConfigFiles:     []string{"~-other-conf-file.conf", "99-other-conf-file.conf"},
			shouldSetFail:        true,
			shouldUnsetFail:      true,
			removeErr:            errors.New("config file doesn't exist"),
		},
		{
			name:                 "reading other files in the directory fails",
			dnsServers:           []string{"1.1.1.1"},
			expectedFileContents: generateConfig(t, "1.1.1.1"),
			getConfigFilesErr:    fmt.Errorf("failed to readdirnames"),
			shouldSetFail:        true,
			shouldUnsetFail:      true,
			removeErr:            errors.New("config file doesn't exist"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockFs := fs.NewSystemFileHandleMock(t)
			mockFs.WriteErr = test.writeErr
			mockFs.RemoveErr = test.removeErr

			getConfigFilesFunc := func() ([]string, error) {
				if test.getConfigFilesErr != nil {
					return []string{}, test.getConfigFilesErr
				}
				return test.otherConfigFiles, nil
			}

			nmcliFunc := func(...string) ([]byte, error) {
				return []byte{}, test.nmcliSetConfigReloadErr
			}

			setter := NMCli{
				runNMCliCommandFunc: nmcliFunc,
				getConfigFilesFunc:  getConfigFilesFunc,
				filesystemHandle:    &mockFs}
			err := setter.Set("", test.dnsServers)

			if test.shouldSetFail {
				assert.Error(t, err, "Expected error to be returned by Set but it was not returned.")
			} else {
				assert.Nil(t, err, "Unexpected error returned by Set.")
				file, ok := mockFs.GetFile(networkManagerConfigFilePath)
				assert.True(t, ok, "Config file was not created after running Set.")
				assert.Equal(t, test.expectedFileContents, string(file), "Unexpected contents of the config file.")
			}

			nmcliFunc = func(...string) ([]byte, error) {
				return []byte{}, test.nmcliUnsetConfigReloadErr
			}
			setter.runNMCliCommandFunc = nmcliFunc

			err = setter.Unset("")
			if test.shouldUnsetFail {
				assert.Error(t, err, "Expected error to be returned by Unset but it was not returned.")
			} else {
				assert.Nil(t, err, "Unexpected error returned by Unset.")
				_, ok := mockFs.GetFile(networkManagerConfigFilePath)
				assert.True(t, ok, "File was not removed after Unset.")
			}
		})
	}
}
