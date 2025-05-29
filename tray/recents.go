package tray

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type RecentConnections struct {
	mu   sync.RWMutex
	List []string `json:"recent"`
	Max  int      `json:"-"`
}

func NewRecentConnections() *RecentConnections {
	return &RecentConnections{
		List: []string{},
		Max:  3,
	}
}

func (rc *RecentConnections) Add(country string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	for i, c := range rc.List {
		if c == country {
			rc.List = append(rc.List[:i], rc.List[i+1:]...)
			break
		}
	}
	rc.List = append([]string{country}, rc.List...)
	if len(rc.List) > rc.Max {
		rc.List = rc.List[:rc.Max]
	}
}

func (rc *RecentConnections) Snapshot() []string {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return append([]string(nil), rc.List...)
}

func (rc *RecentConnections) Save() error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	path, err := getRecentsFilePath()
	if err != nil {
		return err
	}
	tmpPath := path + ".tmp"

	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(rc); err != nil {
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	if dirFd, err := os.Open(filepath.Dir(path)); err == nil {
		_ = dirFd.Sync()
		dirFd.Close()
	}

	return nil
}

func (rc *RecentConnections) Load() error {
	path, err := getRecentsFilePath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	return json.Unmarshal(data, rc)
}

func getRecentsFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(dir, "nordvpn-tray")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "recent.json"), nil
}
