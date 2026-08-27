package norduser

import (
	"errors"
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// createSocketDirectory creates a directory under /run/nordvpn/<uid> where user daemon sockets can be kept
func createSocketDirectory(username string, uidGetter userIDGetter, fsHandle internal.FileSystemHandle) error {
	uids, err := uidGetter.getUserID(username)
	if err != nil {
		return fmt.Errorf("getting user id: %w", err)
	}

	path := internal.GetUserSocketDirectoryPath(int(uids.uid))
	if err := fsHandle.Mkdir(path, internal.PermUserRWX); err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("creating user socket directory: %w", err)
	}

	// Change the directory permissions in case the directory already existed before this function was called.
	if err := fsHandle.Chmod(path, internal.PermUserRWX); err != nil {
		return fmt.Errorf(
			"failed to change user socket directory permissions: %w", err)
	}

	if err := fsHandle.Chown(path, int(uids.uid), int(uids.gid)); err != nil {
		return fmt.Errorf("failed to change socket directory ownership: %w", err)
	}

	return nil
}

// removeSocketDirectory removes the directory under /run/nordvpn/<uid>. If the removal fails the function attempts to
// change the directory ownership to root:root.
func removeSocketDirectory(username string, uidGetter userIDGetter, fsHandle internal.FileSystemHandle) error {
	uids, err := uidGetter.getUserID(username)
	if err != nil {
		return fmt.Errorf("getting user id: %w", err)
	}

	path := internal.GetUserSocketDirectoryPath(int(uids.uid))
	if err := fsHandle.RemoveAll(path); err != nil {
		log.Warn("failed to remove user socket directory:", err)
		// If it's not possible to remove the directory, try to change its ownership so that user will lose access to
		// the socket
		if err := fsHandle.Chown(path, 0, 0); err != nil {
			return fmt.Errorf("failed to remove the socket directory or change its ownership: %w", err)
		}
		// Change directory permissions in case they were changed before the ownership change.
		if err := fsHandle.Chmod(path, internal.PermUserRWX); err != nil {
			return fmt.Errorf(
				"failed to change user socket directory permissions after changing its ownership: %w", err)
		}
	}

	return nil
}
