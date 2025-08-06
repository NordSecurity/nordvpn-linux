package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	ErrSymlinkDetected  = errors.New("symlink detected in path")
	ErrHardlinkDetected = errors.New("hardlink detected")
	ErrPathTraversal    = errors.New("path traversal attempt detected")
	ErrInvalidPath      = errors.New("path outside allowed directories")
	ErrSuspiciousPath   = errors.New("suspicious path pattern detected")
	ErrFileTooLarge     = errors.New("file size exceeds maximum allowed")
	ErrNotRegularFile   = errors.New("not a regular file")
)

// SecureFileRead performs safe file reading with link attack prevention
func SecureFileRead(path string) ([]byte, error) {
	cleanPath := filepath.Clean(path)

	if err := CheckPathForSymlinks(cleanPath); err != nil {
		return nil, fmt.Errorf("symlink check failed: %w", err)
	}

	// check file size first
	info, err := os.Lstat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("stat failed: %w", err)
	}
	if info.Size() > MaxBytesLimit {
		return nil, fmt.Errorf("%w: %d bytes", ErrFileTooLarge, info.Size())
	}

	fd, err := syscall.Open(cleanPath, syscall.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		if !errors.Is(err, syscall.EINVAL) {
			return nil, fmt.Errorf("open failed: %w", err)
		}
		// fallback to regular open if O_NOFOLLOW is not supported
		// but verify it's not a symlink first
		if err := VerifyNotLink(cleanPath); err != nil {
			return nil, err
		}
		// use regular file read
		return FileRead(cleanPath)
	}
	defer syscall.Close(fd)

	// get file info from already open file descriptor (second check)
	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		return nil, fmt.Errorf("fstat failed: %w", err)
	}

	// validate file size again
	if stat.Size > MaxBytesLimit {
		return nil, fmt.Errorf("%w: %d bytes", ErrFileTooLarge, stat.Size)
	}

	// check if regular file (not tty or device)
	if stat.Mode&syscall.S_IFMT != syscall.S_IFREG {
		return nil, ErrNotRegularFile
	}

	// check hardlinks
	if stat.Nlink > 1 {
		return nil, fmt.Errorf("%w: %d links", ErrHardlinkDetected, stat.Nlink)
	}

	// read file
	return io.ReadAll(io.LimitReader(os.NewFile(uintptr(fd), cleanPath), MaxBytesLimit))
}

// CheckPathForSymlinks checks if any component of the path is a symlink
func CheckPathForSymlinks(path string) error {
	var pathComponents []string
	var checkPath string

	// handle absolute and relative paths
	cleanPath := filepath.Clean(path)

	if filepath.IsAbs(cleanPath) {
		pathComponents = strings.Split(cleanPath, string(os.PathSeparator))
		checkPath = "/"
	} else {
		pathComponents = strings.Split(cleanPath, string(os.PathSeparator))
		checkPath = ""
	}

	// check path components in a loop
	for _, component := range pathComponents {
		if component == "" {
			continue
		}

		if checkPath == "" {
			checkPath = component
		} else {
			checkPath = filepath.Join(checkPath, component)
		}

		info, err := os.Lstat(checkPath)
		if err != nil {
			if os.IsNotExist(err) {
				// intermediate directory and all what is beneath
				// or last component which is file doesnt exist yet
				return nil
			}
			return fmt.Errorf("stat failed for %s: %w", checkPath, err)
		}
		// check if symlink (dir cannot be hard linked)
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w at: %s", ErrSymlinkDetected, checkPath)
		}
	}

	return nil
}

// VerifyNotLink verifies that a file is not a symlink or hardlink
func VerifyNotLink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("lstat failed: %w", err)
	}

	// check if symlink
	if info.Mode()&os.ModeSymlink != 0 {
		return ErrSymlinkDetected
	}

	// check for hardlinks
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		if stat.Nlink > 1 {
			return fmt.Errorf("%w: %d links", ErrHardlinkDetected, stat.Nlink)
		}
	}

	return nil
}

// SecureFileWrite performs safe file writing with symlink and hardlink attack prevention
func SecureFileWrite(path string, contents []byte, permissions os.FileMode) error {
	// clean and validate the path
	cleanPath := filepath.Clean(path)

	// check for symlinks in the entire path
	if err := CheckPathForSymlinks(cleanPath); err != nil {
		return fmt.Errorf("symlink check failed: %w", err)
	}

	// ensure directory exists
	dir := filepath.Dir(cleanPath)
	if err := EnsureDir(cleanPath); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// check directory is not a symlink
	if err := CheckPathForSymlinks(dir); err != nil {
		return fmt.Errorf("directory symlink check failed: %w", err)
	}

	// create temporary file in the same directory
	tmpfile, err := os.CreateTemp(dir, ".tmp-config-")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpname := tmpfile.Name()

	// ensure cleanup
	defer func() {
		if tmpfile != nil {
			tmpfile.Close()
			os.Remove(tmpname)
		}
	}()

	// verify temp file is not a link
	if err := VerifyNotLink(tmpname); err != nil {
		return fmt.Errorf("temp file compromised: %w", err)
	}

	// write data to temp file
	if _, err := tmpfile.Write(contents); err != nil {
		return fmt.Errorf("writing to temp file: %w", err)
	}

	// set permissions
	if err := tmpfile.Chmod(permissions); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// sync to disk
	if err := tmpfile.Sync(); err != nil {
		return fmt.Errorf("syncing file: %w", err)
	}

	// close temp file
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	tmpfile = nil

	// remove existing file if it exists (prevents hardlink issues)
	if FileExists(cleanPath) {
		if err := os.Remove(cleanPath); err != nil {
			return fmt.Errorf("removing existing file: %w", err)
		}
	}

	// atomic rename
	if err := os.Rename(tmpname, cleanPath); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	// verify the final file
	if err := VerifyNotLink(cleanPath); err != nil {
		_ = os.Remove(cleanPath)
		return fmt.Errorf("final file verification failed: %w", err)
	}

	// set permissions again (rename might change them)
	if err := os.Chmod(cleanPath, permissions); err != nil {
		return fmt.Errorf("setting final permissions: %w", err)
	}

	return nil
}
