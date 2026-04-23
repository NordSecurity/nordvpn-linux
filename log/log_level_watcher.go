package log

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/filewatch"
	"github.com/fsnotify/fsnotify"
)

// WatchLevelFile starts a goroutine that watches path for writes and updates
// the active log level whenever the file content changes. The file should
// contain one of "debug", "info", or "error" (case-insensitive). The parent
// directory is watched so that file creation is also detected. The goroutine
// stops when the returned CancelFunc is called.
func WatchLevelFile(path string) (CancelFunc, error) {
	Info("setting log level watcher on path", path)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, err
	}

	watcher, err := filewatch.GetFileWatcher(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	applyLevelFile(path)

	go func() {
		defer func() {
			if err := watcher.Close(); err != nil {
				Warn("closing log level watcher:", err)
			}
		}()
		var debounce <-chan time.Time

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name != path {
					continue
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					debounce = time.After(50 * time.Millisecond)
				}
			case <-debounce:
				Info("applying log level from file")
				applyLevelFile(path)
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return func() {
		if err := watcher.Close(); err != nil {
			Warn("closing log level watcher:", err)
		}
	}, nil
}

func applyLevelFile(path string) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return
	}
	defer func() {
		if err := root.Close(); err != nil {
			Errorf("failed to close dir root '%s': %v", filepath.Dir(path), err)
		}
	}()

	f, err := root.Open(filepath.Base(path))
	if err != nil {
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			Errorf("failed to close file '%s': %v", path, err)
		}
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		return
	}

	text := strings.TrimSpace(strings.ToLower(string(data)))
	switch text {
	case "debug":
		SetLevel(levelDebug)
	case "info":
		SetLevel(levelInfo)
	case "warn":
		SetLevel(levelWarn)
	case "error":
		SetLevel(levelError)
	case "fatal":
		SetLevel(levelFatal)
	case "off":
		SetLevel(levelOff)
	default:
		Warn("unknown log level:", text)
		return
	}
}
