package filewatch

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
)

// GetFileWatcher returns a fsnotify file watcher that is monitoring files provided in pathsToMonitor
func GetFileWatcher(pathsToMonitor ...string) (watcher *fsnotify.Watcher, err error) {
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating new watcher: %w", err)
	}

	defer func() {
		if err != nil && watcher != nil {
			_ = watcher.Close()
		}
	}()

	for _, file := range pathsToMonitor {
		if err := watcher.Add(file); err != nil {
			return nil, fmt.Errorf("adding file to watcher: %w", err)
		}
	}

	return watcher, nil
}
