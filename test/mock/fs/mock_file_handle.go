package fs

import (
	"io/fs"
	"os"
	"testing"
)

type SystemFileHandleMock struct {
	files    map[string][]byte
	WriteErr error
	ReadErr  error
	// IsSameFile determines the result of SameFile calls.
	IsSameFile bool
	// StatErrors maps file location to a potential stat error
	StatErrors map[string]error
}

func (fm *SystemFileHandleMock) AddFile(name string, contents []byte) {
	fm.files[name] = contents
}

func (fm *SystemFileHandleMock) FileExists(location string) bool {
	_, ok := fm.files[location]

	return ok
}

func (fm *SystemFileHandleMock) ReadFile(location string) ([]byte, error) {
	if fm.ReadErr != nil {
		return nil, fm.ReadErr
	}
	return fm.files[location], nil
}

func (fm *SystemFileHandleMock) Stat(location string) (os.FileInfo, error) {
	if statErr, ok := fm.StatErrors[location]; ok {
		return &MockFileInfo{}, statErr
	}
	return &MockFileInfo{}, nil
}

func (fm *SystemFileHandleMock) SameFile(fi1 os.FileInfo, fi2 os.FileInfo) bool {
	return fm.IsSameFile
}

func (fm *SystemFileHandleMock) WriteFile(location string, data []byte, mode fs.FileMode) error {
	if fm.WriteErr != nil {
		return fm.WriteErr
	}

	fm.files[location] = data
	return nil
}

func NewSystemFileHandleMock(t *testing.T) SystemFileHandleMock {
	t.Helper()

	return SystemFileHandleMock{
		files:      make(map[string][]byte),
		StatErrors: make(map[string]error),
	}
}
