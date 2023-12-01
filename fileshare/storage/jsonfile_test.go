package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

const historySizeMaxBytes = 4 * 1024 * 1024

// This is used only in tests now. It was easier to just copy this into tests instead of refactoring them
// because we are still keeping Load, so it has to be tested.
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

func TestSimpleSaveLoad(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"

	transfers := make(map[string]*pb.Transfer)
	transfers[transferID] = makeTransfer(transferID, 1, 5, false)
	transfers[transferID].Status = pb.Status_REQUESTED
	i := 0
	for _, file := range transfers[transferID].Files[0].Children {
		if i++; i%2 > 0 {
			file.Status = pb.Status_REQUESTED
		} else {
			file.Status = pb.Status_ONGOING
		}
	}

	storagePath := t.TempDir()
	jsonFile := NewJsonFile(storagePath)
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := NewJsonFile(storagePath)
	loadedTransfers, err := loadedJsonFile.Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedTransfers)
	assert.NotNil(t, loadedTransfers[transferID])

	assert.Equal(t, pb.Status_INTERRUPTED, loadedTransfers[transferID].Status)

	for _, file := range loadedTransfers[transferID].Files[0].Children {
		assert.Equal(t, pb.Status_INTERRUPTED, file.Status)
	}
}

func TestLargeSaveLoad(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"

	transfers := make(map[string]*pb.Transfer)

	const transfersCount = 10

	for i := range [transfersCount]byte{} {
		transferID = fmt.Sprintf("%s-%d", transferID, i)
		transfers[transferID] = makeTransfer(transferID, fileshare.DirDepthLimit, fileshare.TransferFileLimit, true)
	}

	fmt.Printf("transfers count before: %d\n", len(transfers))

	storagePath := t.TempDir()
	jsonFile := NewJsonFile(storagePath)
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := NewJsonFile(storagePath)
	loadedTransfers, err := loadedJsonFile.Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedTransfers)

	fmt.Printf("transfers count after load: %d\n", len(loadedTransfers))

	assert.GreaterOrEqual(t, transfersCount, len(loadedTransfers))
}

func TestNormalSaveLoad(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"

	transfers := make(map[string]*pb.Transfer)

	const transfersCount = 50

	for i := range [transfersCount]byte{} {
		transferID = fmt.Sprintf("%s-%d", transferID, i)
		transfers[transferID] = makeTransfer(transferID, 2, 5, false)
	}

	fmt.Printf("transfers count before: %d\n", len(transfers))

	storagePath := t.TempDir()
	jsonFile := NewJsonFile(storagePath)
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := NewJsonFile(storagePath)
	loadedTransfers, err := loadedJsonFile.Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedTransfers)

	fmt.Printf("transfers count after load: %d\n", len(loadedTransfers))

	assert.GreaterOrEqual(t, transfersCount, len(loadedTransfers))
}

func makeTransfer(transferID string, dirLevels, fileCount int, makeBigNames bool) *pb.Transfer {
	nameSize := 10
	if makeBigNames {
		nameSize = 200
	}

	var files []*pb.File
	for i := 0; i < fileCount; i++ {
		filePath := fmt.Sprintf("%s-%d", strings.Repeat("A", nameSize), i)
		files = append(files, &pb.File{
			Id:     fmt.Sprintf("%s-%d", "asdf845sad84fsadf485sa5d487", i),
			Path:   filePath,
			Size:   10,
			Status: pb.Status_REQUESTED,
		})
	}

	transfer := &pb.Transfer{
		Id:    transferID,
		Files: files,
	}

	return transfer
}

func TestCompatibilityLoad(t *testing.T) {
	category.Set(t, category.Unit)

	category.Set(t, category.Unit)

	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"

	transfers := make(map[string]*pb.Transfer)

	const transfersCount = 50

	for i := range [transfersCount]byte{} {
		transferID = fmt.Sprintf("%s-%d", transferID, i)
		transfers[transferID] = makeLegacyTransfer(transferID, 2, 5, false)
	}

	fmt.Printf("transfers count before: %d\n", len(transfers))

	storagePath := t.TempDir()
	jsonFile := NewJsonFile(storagePath)
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := NewJsonFile(storagePath)
	loadedTransfers, err := loadedJsonFile.Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedTransfers)

	fmt.Printf("transfers count after load: %d\n", len(loadedTransfers))

	assert.GreaterOrEqual(t, transfersCount, len(loadedTransfers))

	for _, transfer := range loadedTransfers {
		assert.Equal(t, 5, len(transfer.Files))
		for _, file := range transfer.Files {
			assert.Equal(t, 0, len(file.Children))
			assert.Equal(t, 2, strings.Count(file.Path, "/")) // ensure that all files are files and not dirs
		}
	}
}

func makeLegacyTransfer(transferID string, dirLevels, fileCount int, makeBigNames bool) *pb.Transfer {
	var crrDir, topDir *pb.File
	nameSize := 10
	if makeBigNames {
		nameSize = 200
	}

	for i := 0; i < dirLevels; i++ {
		dirName := fmt.Sprintf("%s-%d", strings.Repeat("A", nameSize), i)
		if crrDir == nil {
			crrDir = &pb.File{
				Id:       dirName,
				Size:     uint64(0),
				Children: map[string]*pb.File{},
				Status:   pb.Status_REQUESTED,
			}
			topDir = crrDir
		} else {
			crrDir.Children[dirName] = &pb.File{
				Id:       dirName,
				Size:     uint64(0),
				Children: map[string]*pb.File{},
				Status:   pb.Status_REQUESTED,
			}
			crrDir = crrDir.Children[dirName]
		}
	}

	for i := 0; i < fileCount; i++ {
		fileID := fmt.Sprintf("%s-%d", strings.Repeat("A", nameSize), i)
		crrDir.Children[fileID] = &pb.File{
			Id:   fileID,
			Size: 10,
		}
	}

	transfer := &pb.Transfer{
		Id:    transferID,
		Files: []*pb.File{topDir},
	}

	setTransferFiles(transfer, []*pb.File{topDir})

	return transfer
}

// Old utility functions copied over to construct a legacy transfer
func setTransferFiles(tr *pb.Transfer, files []*pb.File) {
	tr.Files = files
	setTransferAllFilePath(tr)
}

func setTransferAllFilePath(tr *pb.Transfer) {
	for _, file := range tr.Files {
		setAllFilePath(file, "")
	}
}

func setAllFilePath(file *pb.File, path string) {
	if path != "" {
		path += "/"
	}
	file.Id = path + file.Id
	if len(file.Children) > 0 {
		for _, childFile := range file.Children {
			setAllFilePath(childFile, file.Id)
		}
	}
}
