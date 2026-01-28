package fs

import (
	"io/fs"
	"time"
)

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
