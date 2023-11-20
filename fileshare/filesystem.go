package fileshare

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
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

// GetDefaultDownloadDirectory returns users Downloads directory or an error if it doesn't exist
func GetDefaultDownloadDirectory() (string, error) {
	username, err := user.Current()
	log.Println(username.Name)

	if err != nil {
		return "", fmt.Errorf("failed to obtain username: %s", err.Error())
	}

	path := filepath.Join(username.HomeDir, "Downloads")
	if _, err = os.Stat(path); err != nil {
		return "", fmt.Errorf("user downloads directory not found: %s", err.Error())
	}

	return path, nil
}
