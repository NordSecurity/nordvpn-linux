package fileshare

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

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
			SetTransferAllFileStatus(tr, pb.Status_INTERRUPTED)
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

// CombinedStorage combines transfers from two storages
// Originally we had our own storage implementation in JSON file. Later libDrop introduced an
// integrated storage solution, so we migrated to that. But to not lose transfer history when
// updating the app, we still load transfers from the original file storage.
type CombinedStorage struct {
	legacy  Storage
	libdrop Storage
}

func NewCombinedStorage(storagePath string, dropStorage Storage) *CombinedStorage {
	return &CombinedStorage{NewJsonFile(storagePath), dropStorage}
}

func (c *CombinedStorage) Load() (map[string]*pb.Transfer, error) {
	libdropTransfers, err := c.libdrop.Load()
	if err != nil {
		return nil, err
	}

	legacyTransfers, err := c.legacy.Load()
	if err != nil {
		for key, value := range legacyTransfers {
			libdropTransfers[key] = value
		}
	}

	return libdropTransfers, nil
}
