package sysinfo

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_systemdTargetDetector_Get(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		target   string
		expected SystemDeviceType
	}{
		{"Valid graphical target", "graphical.target", SystemDeviceTypeDesktop},
		{"Valid server target", "multi-user.target", SystemDeviceTypeServer},
		{"Invalid random target", "memory.target", SystemDeviceTypeUnknown},
		{"Unknown abracadabra target", "abracadabra.target", SystemDeviceTypeUnknown},
		{"Valid server target with spaces", " multi-user.target ", SystemDeviceTypeServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &systemdTargetDetector{
				detectTarget: func() (string, error) {
					return tt.target, nil
				},
			}
			devType, err := d.Get()
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, devType)
		})
	}

	d := &systemdTargetDetector{
		detectTarget: func() (string, error) {
			return "", fmt.Errorf("error occurred")
		},
	}
	_, err := d.Get()
	assert.NotNil(t, err)
}

func Test_graphicalEnvDetector_Get(t *testing.T) {
	category.Set(t, category.Unit)

	mockStat := func(paths map[string]bool) fileInfoFunc {
		return func(name string) (os.FileInfo, error) {
			if paths[name] {
				return &fakeDirInfo{}, nil
			}
			return nil, os.ErrNotExist
		}
	}

	tests := []struct {
		name        string
		envValue    string
		detectErr   error
		statPaths   map[string]bool
		expectType  SystemDeviceType
		expectError bool
	}{
		{
			name:       "Env set to known DE",
			envValue:   "gnome",
			expectType: SystemDeviceTypeDesktop,
		},
		{
			name:       "Env unset with GUI path present",
			envValue:   EnvValueUnset,
			statPaths:  map[string]bool{"/etc/X11": true},
			expectType: SystemDeviceTypeDesktop,
		},
		{
			name:       "Env unset with no GUI paths",
			envValue:   EnvValueUnset,
			statPaths:  map[string]bool{},
			expectType: SystemDeviceTypeUnknown,
		},
		{
			name:        "Detect error",
			detectErr:   fmt.Errorf("fail"),
			expectError: true,
			expectType:  SystemDeviceTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := graphicalEnvDetector{
				detectEnv: func() (string, error) {
					return tt.envValue, tt.detectErr
				},
				statPath: mockStat(tt.statPaths),
			}

			devType, err := d.Get()
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectType, devType)
			}
		})
	}
}

type fakeDirInfo struct{}

func (f *fakeDirInfo) Name() string       { return "mockdir" }
func (f *fakeDirInfo) Size() int64        { return 0 }
func (f *fakeDirInfo) Mode() os.FileMode  { return os.ModeDir }
func (f *fakeDirInfo) ModTime() time.Time { return time.Time{} }
func (f *fakeDirInfo) IsDir() bool        { return true }
func (f *fakeDirInfo) Sys() any           { return nil }

func Test_xdgSessionDetector_Get(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		detectValue string
		detectError error
		expected    SystemDeviceType
	}{
		{"X11 session type", "x11", nil, SystemDeviceTypeDesktop},
		{"Wayland session type", "wayland", nil, SystemDeviceTypeDesktop},
		{"TTY session type", "tty", nil, SystemDeviceTypeServer},
		{"Unknown session type", "black-mesa", nil, SystemDeviceTypeUnknown},
		{"Empty session type", "", nil, SystemDeviceTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := xdgSessionDetector{
				detectSession: func() (string, error) {
					return tt.detectValue, tt.detectError
				},
			}

			devType, err := d.Get()
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, devType)
		})
	}

	d := xdgSessionDetector{
		detectSession: func() (string, error) {
			return "", fmt.Errorf("simulate error")
		},
	}

	_, err := d.Get()
	assert.NotNil(t, err, "must fail with always failing detector")
}
