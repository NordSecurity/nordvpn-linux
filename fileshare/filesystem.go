package fileshare

import (
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// Filesystem defines file operations used by fileshare
type Filesystem interface {
	fs.StatFS
	fs.ReadDirFS
	Statfs(path string) (unix.Statfs_t, error)
	Lstat(path string) (fs.FileInfo, error)
}

// StdFilesystem is a wrapper for golang std filesystem implementation
type StdFilesystem struct {
	basepath string
}

// NewStdFilesystem creates an StdFilesystem instance, basepath is the path prepended to all path arguments
func NewStdFilesystem(basepath string) StdFilesystem {
	return StdFilesystem{
		basepath: basepath,
	}
}

// Open a file
func (stdFs StdFilesystem) Open(name string) (fs.File, error) {
	cleanPath := filepath.Clean(filepath.Join(stdFs.basepath, name))
	return os.Open(cleanPath)
}

// Stat a path
func (stdFs StdFilesystem) Stat(path string) (fs.FileInfo, error) {
	cleanPath := filepath.Clean(filepath.Join(stdFs.basepath, path))
	return os.Stat(cleanPath)
}

// ReadDir returns DirEntry for all of the files and directories in path
func (stdFs StdFilesystem) ReadDir(path string) ([]fs.DirEntry, error) {
	cleanPath := filepath.Clean(filepath.Join(stdFs.basepath, path))
	return os.ReadDir(cleanPath)
}

// Statfs returns info about filesystem
func (stdFs StdFilesystem) Statfs(path string) (unix.Statfs_t, error) {
	cleanPath := filepath.Clean(filepath.Join(stdFs.basepath, path))
	var statfs unix.Statfs_t
	err := unix.Statfs(cleanPath, &statfs)
	return statfs, err
}

// Lstat a path
func (stdFs StdFilesystem) Lstat(path string) (fs.FileInfo, error) {
	cleanPath := filepath.Clean(filepath.Join(stdFs.basepath, path))
	return os.Lstat(cleanPath)
}
