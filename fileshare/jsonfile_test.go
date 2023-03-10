package fileshare

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

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

	jsonFile := JsonFile{}
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := JsonFile{}
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
		transfers[transferID] = makeTransfer(transferID, DirDepthLimit, TransferFileLimit, true)
	}

	fmt.Printf("transfers count before: %d\n", len(transfers))

	jsonFile := JsonFile{}
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := JsonFile{}
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

	jsonFile := JsonFile{}
	assert.NoError(t, jsonFile.Save(transfers))

	loadedJsonFile := JsonFile{}
	loadedTransfers, err := loadedJsonFile.Load()
	assert.NoError(t, err)
	assert.NotNil(t, loadedTransfers)

	fmt.Printf("transfers count after load: %d\n", len(loadedTransfers))

	assert.GreaterOrEqual(t, transfersCount, len(loadedTransfers))
}

func makeTransfer(transferID string, dirLevels, fileCount int, makeBigNames bool) *pb.Transfer {
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
			}
			topDir = crrDir
		} else {
			crrDir.Children[dirName] = &pb.File{
				Id:       dirName,
				Size:     uint64(0),
				Children: map[string]*pb.File{},
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

	SetTransferFiles(transfer, []*pb.File{topDir})

	return transfer
}
