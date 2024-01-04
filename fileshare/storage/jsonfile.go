package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"golang.org/x/exp/maps"
)

const historyFile = "history"

// JsonFile is a implementation of user's fileshare history storage.
type JsonFile struct {
	storagePath string
}

func NewJsonFile(storagePath string) JsonFile {
	return JsonFile{storagePath: storagePath}
}

// Load user's history
func (jf JsonFile) Load() (map[string]*pb.Transfer, error) {
	historyFilePath := path.Join(jf.storagePath, historyFile)
	jsonBytes, err := os.ReadFile(filepath.Clean(historyFilePath))
	if err != nil {
		return nil, fmt.Errorf("loading transfers history file: %w", err)
	}

	var transfers map[string]*pb.Transfer = make(map[string]*pb.Transfer)
	if err := json.Unmarshal(jsonBytes, &transfers); err != nil {
		return nil, fmt.Errorf("unmarshalling transfers history: %w", err)
	}

	for _, tr := range transfers {
		tr.Files = flatten(tr.Files)
		if tr.Status == pb.Status_REQUESTED || tr.Status == pb.Status_ONGOING {
			tr.Status = pb.Status_INTERRUPTED
			fileshare.SetTransferAllFileStatus(tr, pb.Status_INTERRUPTED)
		}
	}

	return transfers, nil
}

// Previously libdrop returned file trees as a tree, but now it returns a flat file list.
// For the users that upgrade nordvpn this converts old format to the new.
func flatten(files []*pb.File) []*pb.File {
	var flatFiles []*pb.File
	for _, file := range files {
		if len(file.Children) > 0 {
			flatFiles = append(flatFiles, flatten(maps.Values(file.Children))...)
		} else {
			if file.Path == "" {
				file.Path = file.Id
			}
			flatFiles = append(flatFiles, file)
		}
	}
	return flatFiles
}

func (jf JsonFile) PurgeTransfersUntil(until time.Time) error {
	historyFilePath := path.Join(jf.storagePath, historyFile)
	info, err := os.Stat(filepath.Clean(historyFilePath))

	if err == nil {
		if info.ModTime().Before(until) {
			if err := os.Remove(filepath.Clean(historyFilePath)); err != nil {
				return fmt.Errorf("removing transfers history file: %w", err)
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stating transfers history file: %w", err)
	}

	return nil
}
