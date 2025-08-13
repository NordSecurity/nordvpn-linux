package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureFileRead_PreventSymlinkRead(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	// create a sensitive file
	sensitiveFile := filepath.Join(tempDir, "sensitive.txt")
	err := os.WriteFile(sensitiveFile, []byte("secret data"), PermUserRW)
	require.NoError(t, err)

	// create a symlink to it
	symlinkPath := filepath.Join(tempDir, "config.json")
	err = os.Symlink(sensitiveFile, symlinkPath)
	require.NoError(t, err)

	// try to read through the symlink
	_, err = SecureFileRead(symlinkPath)

	// should fail with symlink error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symlink")
}

func TestSecureFileRead_PreventHardlinkRead(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	// create a source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("data"), PermUserRW)
	require.NoError(t, err)

	// create a hardlink
	hardlinkPath := filepath.Join(tempDir, "config.json")
	err = os.Link(sourceFile, hardlinkPath)
	require.NoError(t, err)

	// try to read the hardlink
	_, err = SecureFileRead(hardlinkPath)

	// should fail with hardlink error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hardlink")
}

func TestSecureFileOperations_NormalUse(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	testData := []byte(`{"version": "1"}`)

	err := SecureFileWrite(configPath, testData, PermUserRW)
	assert.NoError(t, err)

	readData, err := SecureFileRead(configPath)
	assert.NoError(t, err)
	assert.Equal(t, testData, readData)

	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(PermUserRW), info.Mode().Perm())
}

func TestCheckForSymlinks(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		expectError bool
	}{
		{
			name: "regular file",
			setup: func() string {
				path := filepath.Join(tempDir, "regular.txt")
				os.WriteFile(path, []byte("data"), PermUserRW)
				return path
			},
			expectError: false,
		},
		{
			name: "direct symlink",
			setup: func() string {
				target := filepath.Join(tempDir, "target.txt")
				link := filepath.Join(tempDir, "link.txt")
				os.WriteFile(target, []byte("data"), PermUserRW)
				os.Symlink(target, link)
				return link
			},
			expectError: true,
		},
		{
			name: "symlink in path",
			setup: func() string {
				realDir := filepath.Join(tempDir, "realdir")
				os.Mkdir(realDir, 0755)
				linkDir := filepath.Join(tempDir, "linkdir")
				os.Symlink(realDir, linkDir)
				return filepath.Join(linkDir, "file.txt")
			},
			expectError: true,
		},
		{
			name: "non-existent file",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent.txt")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			err := CheckPathForSymlinks(path)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "symlink")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVerifyNotLink(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	// test regular file
	regularFile := filepath.Join(tempDir, "regular.txt")
	err := os.WriteFile(regularFile, []byte("data"), PermUserRW)
	require.NoError(t, err)

	err = VerifyNotLink(regularFile)
	assert.NoError(t, err)

	// test symlink
	target := filepath.Join(tempDir, "target.txt")
	symlink := filepath.Join(tempDir, "symlink.txt")
	err = os.WriteFile(target, []byte("data"), PermUserRW)
	require.NoError(t, err)
	err = os.Symlink(target, symlink)
	require.NoError(t, err)

	err = VerifyNotLink(symlink)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSymlinkDetected)

	// test hardlink
	hardlink := filepath.Join(tempDir, "hardlink.txt")
	err = os.Link(regularFile, hardlink)
	require.NoError(t, err)

	err = VerifyNotLink(hardlink)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hardlink")
}

func TestSecureFileWrite_PreventSymlinkAttack(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	// create a target file that we don't want to be overwritten
	sensitiveFile := filepath.Join(tempDir, "sensitive.txt")
	err := os.WriteFile(sensitiveFile, []byte("sensitive data"), PermUserRW)
	require.NoError(t, err)

	// create a symlink pointing to the sensitive file
	symlinkPath := filepath.Join(tempDir, "config.json")
	err = os.Symlink(sensitiveFile, symlinkPath)
	require.NoError(t, err)

	// try to write through the symlink using secure write
	err = SecureFileWrite(symlinkPath, []byte("malicious data"), PermUserRW)

	// should fail with symlink error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symlink")

	// verify sensitive file was not modified
	content, err := os.ReadFile(sensitiveFile)
	require.NoError(t, err)
	assert.Equal(t, "sensitive data", string(content))
}

func TestSecureFileWrite_PreventHardlinkAttack(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	// create a source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(sourceFile, []byte("original data"), PermUserRW)
	require.NoError(t, err)

	// create a hardlink
	hardlinkPath := filepath.Join(tempDir, "config.json")
	err = os.Link(sourceFile, hardlinkPath)
	require.NoError(t, err)

	// try to write to the hardlink using secure write
	err = SecureFileWrite(hardlinkPath, []byte("new data"), PermUserRW)

	// should succeed but break the hardlink
	assert.NoError(t, err)

	// verify files have different content (hardlink was broken)
	sourceContent, err := os.ReadFile(sourceFile)
	require.NoError(t, err)
	assert.Equal(t, "original data", string(sourceContent))

	hardlinkContent, err := os.ReadFile(hardlinkPath)
	require.NoError(t, err)
	assert.Equal(t, "new data", string(hardlinkContent))
}

func TestSecureFileWrite_SymlinkInPath(t *testing.T) {
	category.Set(t, category.Unit)

	tempDir := t.TempDir()

	// create a directory symlink
	realDir := filepath.Join(tempDir, "real")
	err := os.Mkdir(realDir, PermUserRWX)
	require.NoError(t, err)

	linkDir := filepath.Join(tempDir, "link")
	err = os.Symlink(realDir, linkDir)
	require.NoError(t, err)

	// try to write through the symlink directory
	targetPath := filepath.Join(linkDir, "config.json")
	err = SecureFileWrite(targetPath, []byte("data"), PermUserRW)

	// should fail because path contains symlink
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symlink")
}
