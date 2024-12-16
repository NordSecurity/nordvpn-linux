package meshnet

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestGiveProcessPID(t *testing.T) {
	category.Set(t, category.Unit)

	execPath := "/usr/lib/nordvpn/nordfileshare"
	pid := PID(12)
	tests := []struct {
		name             string
		expectedExecPath string
		readdir          readdirFunc
		readfile         readfileFunc
		expectedPID      *PID
	}{
		{
			name:             "returns nil for empty process list",
			expectedExecPath: "/not/important",
			readdir:          func(string) ([]os.DirEntry, error) { return []os.DirEntry{}, nil },
			readfile:         defaultReadfile,
			expectedPID:      nil,
		},
		{
			name:             "returns correct PID with exactly one process on the list",
			expectedExecPath: execPath,
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "12",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte(execPath), nil },
			expectedPID: &pid,
		},
		{
			name:             "returns correct PID for running snap process",
			expectedExecPath: "/snap/nordvpn/x1/usr/lib/nordvpn/nordfileshare",
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "12",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte("/snap/nordvpn/x1/usr/lib/nordvpn/nordfileshare"), nil },
			expectedPID: &pid,
		},
		{
			name:             "returns true with same path but in non-canonical form",
			expectedExecPath: "/../../..//usr//lib/nordvpn/some/dirs/../..//nordfileshare",
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "12",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte(execPath), nil },
			expectedPID: &pid,
		},
		{
			name:             "returns correct PID with multiple processes and one is valid",
			expectedExecPath: execPath,
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "wrong-dir-name",
					},
					&MockDirEntry{
						name: "12",
					},
					&MockDirEntry{
						name: "42",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte(execPath), nil },
			expectedPID: &pid,
		},
		{
			name:             "returns nil with multiple processes and none is valid",
			expectedExecPath: execPath,
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "wrong-dir-name",
					},
					&MockDirEntry{
						name: "another-wrong",
					},
					&MockDirEntry{
						name: "42",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte("/different/path"), nil },
			expectedPID: nil,
		},
		{
			name:             "returns nil when directory is not a pid",
			expectedExecPath: "/not/important",
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "not-a-pid",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte("/also/not/important"), nil },
			expectedPID: nil,
		},
		{
			name:             "returns nil when unable to read cmdline file",
			expectedExecPath: "/not/important",
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "42",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return nil, errors.New("test error") },
			expectedPID: nil,
		},
		{
			name:             "returns nil when cmdline is empty",
			expectedExecPath: "/not/important",
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "42",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte(""), nil },
			expectedPID: nil,
		},
		{
			name:             "returns nil when no process has expected executable path",
			expectedExecPath: execPath,
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{
					&MockDirEntry{
						name: "not-a-pid",
					},
				}, nil
			},
			readfile:    func(string) ([]byte, error) { return []byte("/some/other/path"), nil },
			expectedPID: nil,
		},
		{
			name:             "returns nil when  readdir fails to read directories",
			expectedExecPath: execPath,
			readdir: func(string) ([]os.DirEntry, error) {
				return []os.DirEntry{}, errors.New("test error")
			},
			readfile:    func(string) ([]byte, error) { return []byte(""), nil },
			expectedPID: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			procChecker := DefaultProcChecker{
				readfile: test.readfile,
				readdir:  test.readdir,
			}

			PID := procChecker.GiveProcessPID(test.expectedExecPath)

			assert.Equal(t, test.expectedPID, PID)
		})
	}
}

type MockDirEntry struct {
	name string
}

func (m *MockDirEntry) Name() string {
	return m.name
}

func (m *MockDirEntry) IsDir() bool {
	return true
}

func (m *MockDirEntry) Type() fs.FileMode {
	return os.ModeSymlink
}

func (m *MockDirEntry) Info() (fs.FileInfo, error) {
	return &MockFileInfo{
		name:    m.name,
		size:    1024,
		mode:    os.FileMode(0644),
		modTime: time.Now(),
		sys:     &syscall.Stat_t{},
	}, nil
}

type MockFileInfo struct {
	modTime time.Time
	sys     any
	name    string
	size    int64
	mode    fs.FileMode
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m *MockFileInfo) ModTime() time.Time { return m.modTime }
func (m *MockFileInfo) IsDir() bool        { return m.mode.IsDir() }
func (m *MockFileInfo) Sys() any           { return m.sys }
