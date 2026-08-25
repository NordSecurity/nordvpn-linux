package norduser

import (
	"errors"
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock/fs"
	"github.com/stretchr/testify/assert"
)

func TestCreateRemoveSocketDirectory(t *testing.T) {
	category.Set(t, category.Unit)

	user1Name := "user1"
	user1UID := 1000
	user1GID := 1000
	user1IDs := userIDs{
		uid: uint32(user1UID),
		gid: uint32(user1GID),
	}

	user1SocketDirPath := internal.RunDir + "/1000"

	usernameToIDs := map[string]userIDs{
		user1Name: user1IDs,
	}
	uidGetterMock := userIDGetterMock{
		UsernameToIDs: usernameToIDs,
	}

	tests := []struct {
		name         string
		username     string
		mkdirErr     error
		chownErr     error
		chmodErr     error
		removeAllErr error
		// removalChownErr is injected after the directory is created, so that it only affects the ownership fallback
		// performed when the removal fails
		removalChownErr error
		// removalChmodErr is injected after the directory is created, so that it only affects the permissions fallback
		// performed when the removal fails
		removalChmodErr error
		// widenedDirMode simulates the user widening the permissions of the socket directory they own, before the
		// removal is attempted. Ignored when zero.
		widenedDirMode os.FileMode
		// existingDir seeds a socket directory that survived a previous daemon run, before the creation is attempted
		existingDir *fs.DirectoryInfo
		// socketCreationFail means that creation is expected to fail
		socketCreationFail bool
		// socketDirLeftAfterFailedCreation means that a directory is expected to be left behind when creation fails
		socketDirLeftAfterFailedCreation bool
		// socketRemovalFail means that removal is expected to fail
		socketRemovalFail bool
		// socketDirLeftAfterRemoval means that a directory is expected to be left behind after removal, with its
		// ownership changed to root
		socketDirLeftAfterRemoval bool
		// socketDirOwnedByRootAfterFailedRemoval means that the ownership fallback is expected to succeed even though
		// the removal has failed
		socketDirOwnedByRootAfterFailedRemoval bool
	}{
		{
			name:     "directory creation removal success",
			username: user1Name,
		},
		{
			name:               "unknown user",
			username:           "unknown-user",
			socketCreationFail: true,
			socketRemovalFail:  true,
		},
		{
			name:               "directory creation fails",
			username:           user1Name,
			mkdirErr:           errors.New("failed to create directory"),
			socketCreationFail: true,
		},
		{
			name:                             "directory ownership change fails",
			username:                         user1Name,
			chownErr:                         errors.New("failed to change ownership"),
			socketCreationFail:               true,
			socketDirLeftAfterFailedCreation: true,
		},
		{
			name:                             "directory permissions change fails",
			username:                         user1Name,
			chmodErr:                         errors.New("failed to change permissions"),
			socketCreationFail:               true,
			socketDirLeftAfterFailedCreation: true,
		},
		{
			// Directories under /run are preserved between the daemon restarts, so the daemon has to handle an already
			// existing socket directory. The user owns it, so they could have widened its permissions.
			name:     "directory already exists with widened permissions",
			username: user1Name,
			mkdirErr: os.ErrExist,
			existingDir: &fs.DirectoryInfo{
				Mode: 0777,
				Uid:  user1UID,
				Gid:  user1GID,
			},
		},
		{
			// A directory that could not be removed is left owned by root, so the ownership has to be handed back to
			// the user when they rejoin the group.
			name:     "directory left over from a failed removal",
			username: user1Name,
			mkdirErr: os.ErrExist,
			existingDir: &fs.DirectoryInfo{
				Mode: internal.PermUserRWX,
				Uid:  0,
				Gid:  0,
			},
		},
		{
			name:                      "directory removal fails",
			username:                  user1Name,
			removeAllErr:              errors.New("failed to remove directory"),
			widenedDirMode:            0777,
			socketDirLeftAfterRemoval: true,
		},
		{
			name:              "directory removal and ownership change fail",
			username:          user1Name,
			removeAllErr:      errors.New("failed to remove directory"),
			removalChownErr:   errors.New("failed to change ownership"),
			socketRemovalFail: true,
		},
		{
			name:                                   "directory removal and permissions change fail",
			username:                               user1Name,
			removeAllErr:                           errors.New("failed to remove directory"),
			removalChmodErr:                        errors.New("failed to change permissions"),
			widenedDirMode:                         0777,
			socketRemovalFail:                      true,
			socketDirOwnedByRootAfterFailedRemoval: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fsMock := fs.NewSystemFileHandleMock(t)
			fsMock.MkdirErr = test.mkdirErr
			if test.existingDir != nil {
				fsMock.Directories[user1SocketDirPath] = *test.existingDir
			}
			fsMock.ChownErr = test.chownErr
			fsMock.ChmodErr = test.chmodErr

			err := createSocketDirectory(test.username, &uidGetterMock, &fsMock)

			if test.socketCreationFail {
				assert.Error(t, err, "Expected an error when creating user socket directory.")

				if test.socketDirLeftAfterFailedCreation {
					directory, ok := fsMock.Directories[user1SocketDirPath]
					assert.True(t, ok, "User socket directory should be left behind.")
					// Ownership was never handed over to the user, so the directory remains inaccessible to them.
					assert.Equal(t, 0, directory.Uid, "Socket directory should not be owned by the user.")
					assert.Equal(t, 0, directory.Gid, "Socket directory should not be owned by the user group.")
				} else {
					assert.Empty(t, fsMock.Directories, "No directory should be left behind when creation fails.")
				}
			} else {
				assert.NoError(t, err, "Unexpected error when creating user socket directory.")

				directory, ok := fsMock.Directories[user1SocketDirPath]
				assert.True(t, ok, "User socket directory was not created.")
				assert.Equal(t, os.FileMode(internal.PermUserRWX), directory.Mode.Perm(), "Invalid permissions set for the directory.")
				assert.Equal(t, user1UID, directory.Uid, "Invalid UID set for user socket directory.")
				assert.Equal(t, user1GID, directory.Gid, "Invalid GID set for user socket directory.")
			}

			if test.widenedDirMode != 0 {
				directory := fsMock.Directories[user1SocketDirPath]
				directory.Mode = test.widenedDirMode
				fsMock.Directories[user1SocketDirPath] = directory
			}

			fsMock.RemoveAllErr = test.removeAllErr
			if test.removalChownErr != nil {
				fsMock.ChownErr = test.removalChownErr
			}
			fsMock.ChmodErr = test.removalChmodErr
			err = removeSocketDirectory(test.username, &uidGetterMock, &fsMock)

			if test.socketRemovalFail {
				assert.Error(t, err, "Expected an error when removing user socket directory.")

				if directory, ok := fsMock.Directories[user1SocketDirPath]; ok {
					if test.socketDirOwnedByRootAfterFailedRemoval {
						// Only the permissions change failed, so the user has already lost the ownership.
						assert.Equal(t, 0, directory.Uid, "Socket directory ownership was not changed to root.")
						assert.Equal(t, 0, directory.Gid, "Socket directory group ownership was not changed to root.")
					} else {
						// The ownership change failed, so the user retains access to the directory.
						assert.Equal(t, user1UID, directory.Uid, "Socket directory UID should be left untouched.")
						assert.Equal(t, user1GID, directory.Gid, "Socket directory GID should be left untouched.")
					}
				}
			} else {
				assert.NoError(t, err, "Unexpected error when removing user socket directory.")

				directory, ok := fsMock.Directories[user1SocketDirPath]
				if test.socketDirLeftAfterRemoval {
					assert.True(t, ok, "User socket directory should be left behind when it cannot be removed.")
					// The directory could not be removed, so the user should at least lose access to it.
					assert.Equal(t, 0, directory.Uid, "Socket directory ownership was not changed to root.")
					assert.Equal(t, 0, directory.Gid, "Socket directory group ownership was not changed to root.")
					assert.Equal(t, os.FileMode(internal.PermUserRWX), directory.Mode.Perm(),
						"Socket directory permissions were not restricted after the ownership change.")
				} else {
					assert.False(t, ok, "User socket directory was not removed.")
				}
			}
		})
	}
}
