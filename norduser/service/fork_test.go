package service

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"slices"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_parseNorduserPIDs(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		pids         []string
		expectedPIDs []int
	}{
		{
			name:         "empty list",
			pids:         []string{},
			expectedPIDs: []int{},
		},
		{
			name:         "non empty list",
			pids:         []string{" 35139", " 35153", " 35144"},
			expectedPIDs: []int{35139, 35153, 35144},
		},
		{
			name:         "list contains malformed entries",
			pids:         []string{" 35139", " aaaa", " 35144"},
			expectedPIDs: []int{35139, 35144},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pids := strings.Join(test.pids, "\n")
			result := parseNorduserPIDs(pids)
			assert.Equal(t, test.expectedPIDs, result)
		})
	}
}

func Test_findPIDOfUID(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		uidToPID    []string
		uid         int
		expectedPID int
	}{
		{
			name:        "empty list",
			uidToPID:    []string{},
			uid:         1001,
			expectedPID: -1,
		},
		{
			name:        "list with empty lines",
			uidToPID:    []string{"\n1001 35255\n"},
			uid:         1001,
			expectedPID: 35255,
		},
		{
			name:        "uid not present",
			uidToPID:    []string{" 1004 35139", " 1003 35153", " 1002 35144"},
			uid:         1001,
			expectedPID: -1,
		},
		{
			name:        "invalid pid",
			uidToPID:    []string{" 1001 aaaa", " 1003 35153", " 1002 35144"},
			uid:         1001,
			expectedPID: -1,
		},
		{
			name:        "pid found",
			uidToPID:    []string{" 1001 35255", " 1003 35153", " 1002 35144"},
			uid:         1001,
			expectedPID: 35255,
		},
		{
			name:        "different format",
			uidToPID:    []string{" 1001 35255", "10003 35153", "1002 35144"},
			uid:         10003,
			expectedPID: 35153,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pids := strings.Join(test.uidToPID, "\n")
			result := findPIDOfUID(pids, uint32(test.uid))
			assert.Equal(t, test.expectedPID, result)
		})
	}
}

type envConfiguratorMock struct{}

func (m *envConfiguratorMock) ConfigureEnv(uid, gid uint32) (*exec.Cmd, error) {
	// simulate a command that echoes a variable similar to `printenv`
	return exec.Command("echo", "XDG_CURRENT_DESKTOP=ubuntu:GNOME"), nil
}

type envConfiguratorMockWithConfiguratorError struct{}

func (m *envConfiguratorMockWithConfiguratorError) ConfigureEnv(uid, gid uint32) (*exec.Cmd, error) {
	return nil, fmt.Errorf("always error")
}

type envConfiguratorMockWithCmdError struct{}

func (m *envConfiguratorMockWithCmdError) ConfigureEnv(uid, gid uint32) (*exec.Cmd, error) {
	// simulate a command that always fails to execute
	return exec.Command("exit1", "1"), nil
}

func Test_MergeUserSessionEnv(t *testing.T) {
	mockConf := &envConfiguratorMock{}
	var currentEnv []string

	err := mergeUserSessionEnv(1000, 1000, &currentEnv, mockConf)
	assert.NoError(t, err, "mergeUserSessionEnv should not return an error")

	// validate the environment contains the expected variable
	found := slices.Contains(currentEnv, "XDG_CURRENT_DESKTOP=ubuntu:GNOME")
	assert.True(t, found, "Expected 'XDG_CURRENT_DESKTOP=ubuntu:GNOME' in environment")

	currentEnv = []string{}
	err = mergeUserSessionEnv(1000, 1000, &currentEnv, &envConfiguratorMockWithConfiguratorError{})
	assert.Error(t, err, "must return an error when invalid configurator is provided")

	currentEnv = []string{}
	err = mergeUserSessionEnv(1000, 1000, &currentEnv, &envConfiguratorMockWithCmdError{})
	assert.Error(t, err, "must return an error when invalid command is provided")

	err = mergeUserSessionEnv(1000, 1000, nil, &envConfiguratorMock{})
	assert.Error(t, err, "must return an error when invalid env slice is provided")
}

type gidProviderMock struct{}

func (m *gidProviderMock) GetNordvpnGid() (uint32, error) {
	return 1234, nil
}

type gidProviderMockErroneous struct{}

func (g gidProviderMockErroneous) GetNordvpnGid() (uint32, error) {
	return 0, fmt.Errorf("always fails")
}

func Test_SystemEnvConfigurator(t *testing.T) {
	conf := SystemEnvConfigurator{provider: &gidProviderMock{}}
	mockUid := uint32(1004)
	mockGid := uint32(1004)

	cmd, err := conf.ConfigureEnv(mockUid, mockGid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd.Path != "systemctl" && cmd.Args[1] != "--user" && cmd.Args[2] != "show-environment" {
		t.Fatalf("unexpected command configuration: %+v", cmd.Args)
	}

	if len(cmd.Env) == 0 {
		t.Fatalf("environment variable list is empty")
	}

	found_at := -1
	for it, val := range cmd.Env {
		if strings.Contains(val, "XDG_RUNTIME_DIR=") {
			found_at = it
			break
		}
	}

	if found_at == -1 {
		t.Fatalf("environment does not contain 'XDG_RUNTIME_DIR' variable: %+v", cmd.Env)
	}

	runtimeDir := strings.Split(cmd.Env[found_at], "=")[1]
	if runtimeDir != fmt.Sprintf("/run/user/%d", mockUid) {
		t.Fatalf("incorrect runtime user directory: %v", runtimeDir)
	}

	if cmd.SysProcAttr.Credential.Uid != mockUid {
		t.Fatalf("incorrect credentials (uid): %v, expected: %v",
			cmd.SysProcAttr.Credential.Uid, mockUid)
	}

	if cmd.SysProcAttr.Credential.Gid != mockGid {
		t.Fatalf("incorrect credentials (gid): %v, expected: %v",
			cmd.SysProcAttr.Credential.Gid, mockGid)
	}

	if len(cmd.SysProcAttr.Credential.Groups) == 0 {
		t.Fatalf("supplementary groups are empty")
	}

	found := slices.Contains(cmd.SysProcAttr.Credential.Groups, 1234)

	if !found {
		t.Fatalf("nordvpn group gid not found: %+v", cmd.SysProcAttr.Credential.Groups)
	}

	conf2 := SystemEnvConfigurator{provider: &gidProviderMockErroneous{}}
	_, err = conf2.ConfigureEnv(1000, 1000)
	if err == nil {
		t.Fatal("expected error")
	}
}
