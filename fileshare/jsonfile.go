package fileshare

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
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

// Save user's history
func (jf JsonFile) Save(transfers map[string]*pb.Transfer) (err error) {
	historyFilePath := path.Join(jf.storagePath, historyFile)
	if err := internal.EnsureDir(historyFilePath); err != nil {
		return fmt.Errorf("trying to save transfers history: %w", err)
	}

	var trBytes []byte
	for {
		trBytes, err = json.Marshal(transfers)
		if err != nil {
			return err
		}

		if len(trBytes) < historySizeMaxBytes {
			break
		}

		// truncate history; find the oldest completed transfer and remove it
		log.Printf("truncating transfers history json size: %d (max limit: %d)\n", len(trBytes), historySizeMaxBytes)
		var oldestTransfer *pb.Transfer
		for _, tr := range transfers {
			if tr.Status == pb.Status_ONGOING {
				continue
			}
			if oldestTransfer == nil {
				oldestTransfer = tr
			} else if tr.Created.AsTime().Before(oldestTransfer.Created.AsTime()) {
				oldestTransfer = tr
			}
		}

		if oldestTransfer == nil {
			log.Println("cannot truncate transfers history")
			break
		} else {
			delete(transfers, oldestTransfer.Id)
		}
	}

	// write (overwrite if exists) and close file
	return os.WriteFile(historyFilePath, trBytes, internal.PermUserRW)
}
