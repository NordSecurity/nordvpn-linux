package fs

import (
	"io/fs"
	"os"
	"syscall"
	"time"
)

type MockDirEntry struct {
	DirName string
}

func (m *MockDirEntry) Name() string {
	return m.DirName
}

func (m *MockDirEntry) IsDir() bool {
	return true
}

func (m *MockDirEntry) Type() fs.FileMode {
	return os.ModeSymlink
}

func (m *MockDirEntry) Info() (fs.FileInfo, error) {
	return &MockFileInfo{
		name:    m.DirName,
		size:    1024,
		mode:    os.FileMode(0644),
		modTime: time.Now(),
		sys:     &syscall.Stat_t{},
	}, nil
}
