package config

import (
	"io/fs"
	"testing"
)

type FilesystemMock struct {
	files    map[string][]byte
	WriteErr error
	ReadErr  error
}

func (fm *FilesystemMock) AddFile(name string, contents []byte) {
	fm.files[name] = contents
}

func (fm *FilesystemMock) FileExists(location string) bool {
	_, ok := fm.files[location]

	return ok
}

func (fm *FilesystemMock) ReadFile(location string) ([]byte, error) {
	if fm.ReadErr != nil {
		return nil, fm.ReadErr
	}
	return fm.files[location], nil
}

func (fm *FilesystemMock) WriteFile(location string, data []byte, mode fs.FileMode) error {
	if fm.WriteErr != nil {
		return fm.WriteErr
	}

	fm.files[location] = data
	return nil
}

func NewFilesystemMock(t *testing.T) FilesystemMock {
	t.Helper()

	return FilesystemMock{
		files: make(map[string][]byte),
	}
}
