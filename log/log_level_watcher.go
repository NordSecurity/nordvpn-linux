package log

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// WatchLevelFile starts a goroutine that watches path for writes and updates
// the active log level whenever the file content changes. The file should
// contain one of "debug", "info", or "error" (case-insensitive). The parent
// directory is watched so that file creation is also detected. The goroutine
// stops when the returned CancelFunc is called.
func WatchLevelFile(path string) (CancelFunc, error) {
  Info("setting log level watcher on path", path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	applyLevelFile(path)

	if err := watcher.Add(filepath.Dir(path)); err != nil {
		if closeErr := watcher.Close(); closeErr != nil {
			Warn("closing log level watcher:", closeErr)
		}
		return nil, err
	}

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
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	text := strings.TrimSpace(strings.ToLower(string(data)))
	switch text {
	case "debug":
		SetLevel(LevelDebug)
	case "info":
		SetLevel(LevelInfo)
	case "warn":
		SetLevel(LevelWarn)
	case "error":
		SetLevel(LevelError)
	case "fatal":
		SetLevel(LevelFatal)
	case "off":
		SetLevel(LevelOff)
	default:
		Warn("unknown log level:", text)
		return
	}
}
