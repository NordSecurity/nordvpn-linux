package fs

import (
	"fmt"
	"io/fs"
	"os"
	"testing"
)

type DirectoryInfo struct {
	Mode os.FileMode
	Uid  int
	Gid  int
}

type SystemFileHandleMock struct {
	files    map[string][]byte
	WriteErr error
	ReadErr  error
	Links    map[string]string
	// StatErrors maps file location to a potential stat error
	StatErrors   map[string]error
	RemoveErr    error
	RemoveAllErr error
	Directories  map[string]DirectoryInfo
	MkdirErr     error
	ChownErr     error
	ChmodErr     error
}

func (fm *SystemFileHandleMock) GetFile(location string) ([]byte, bool) {
	contents, ok := fm.files[location]
	return contents, ok
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
	return &MockFileInfo{name: location}, nil
}

func (fm *SystemFileHandleMock) SameFile(fi1 os.FileInfo, fi2 os.FileInfo) bool {
	if linkName, ok := fm.Links[fi1.Name()]; ok {
		return fi2.Name() == linkName
	}
	return false
}

func (fm *SystemFileHandleMock) WriteFile(location string, data []byte, mode fs.FileMode) error {
	if fm.WriteErr != nil {
		return fm.WriteErr
	}

	fm.files[location] = data
	return nil
}

func (fm *SystemFileHandleMock) Remove(location string) error {
	if fm.RemoveErr != nil {
		return fm.RemoveErr
	}

	delete(fm.files, location)

	return nil
}

func (fm *SystemFileHandleMock) RemoveAll(path string) error {
	if fm.RemoveAllErr != nil {
		return fm.RemoveAllErr
	}

	delete(fm.files, path)
	delete(fm.Directories, path)

	return nil
}

func (fm *SystemFileHandleMock) Mkdir(path string, perm os.FileMode) error {
	if fm.MkdirErr != nil {
		return fm.MkdirErr
	}

	fm.Directories[path] = DirectoryInfo{
		Mode: perm,
	}

	return nil
}

func (fm *SystemFileHandleMock) Chown(name string, uid int, gid int) error {
	dir, ok := fm.Directories[name]
	if !ok {
		return fmt.Errorf("no such directory")
	}

	if fm.ChownErr != nil {
		return fm.ChownErr
	}

	dir.Uid = uid
	dir.Gid = gid
	fm.Directories[name] = dir

	return nil
}

func (fm *SystemFileHandleMock) Chmod(name string, perm os.FileMode) error {
	dir, ok := fm.Directories[name]
	if !ok {
		return fmt.Errorf("no such directory")
	}

	if fm.ChmodErr != nil {
		return fm.ChmodErr
	}

	dir.Mode = perm
	fm.Directories[name] = dir

	return nil
}

func NewSystemFileHandleMock(t *testing.T) SystemFileHandleMock {
	t.Helper()

	return SystemFileHandleMock{
		files:       make(map[string][]byte),
		Links:       make(map[string]string),
		StatErrors:  make(map[string]error),
		Directories: make(map[string]DirectoryInfo),
	}
}
