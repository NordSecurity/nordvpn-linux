package fileshare

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

const historyFile = "history"

// JsonFile is a implementation of user's fileshare history storage.
type JsonFile struct{}

// Load user's history
func (JsonFile) Load() (map[string]*pb.Transfer, error) {
	currentUser, _ := user.Current()
	// we have to hardcode config directory, using os.UserConfigDir is not viable as nordfileshared
	// is spawned by nordvpnd(owned by root) and inherits roots environment variables
	historyFilePath := path.Join(currentUser.HomeDir, internal.ConfigDirectory, internal.UserDataPath, historyFile)

	jsonBytes, err := os.ReadFile(filepath.Clean(historyFilePath))
	if err != nil {
		return nil, fmt.Errorf("loading transfers history file: %w", err)
	}

	var transfers map[string]*pb.Transfer = make(map[string]*pb.Transfer)
	if err := json.Unmarshal(jsonBytes, &transfers); err != nil {
		return nil, fmt.Errorf("unmarshalling transfers history: %w", err)
	}

	for _, tr := range transfers {
		if tr.Status == pb.Status_REQUESTED || tr.Status == pb.Status_ONGOING {
			tr.Status = pb.Status_INTERRUPTED
			SetTransferAllFileStatus(tr, pb.Status_INTERRUPTED)
		}
	}

	return transfers, nil
}

// Save user's history
func (JsonFile) Save(transfers map[string]*pb.Transfer) (err error) {
	currentUser, _ := user.Current()
	// we have to hardcode config directory, using os.UserConfigDir is not viable as nordfileshared
	// is spawned by nordvpnd(owned by root) and inherits roots environment variables
	historyFilePath := path.Join(currentUser.HomeDir, internal.ConfigDirectory, internal.UserDataPath, historyFile)

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
