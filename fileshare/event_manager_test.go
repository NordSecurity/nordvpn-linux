package fileshare

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestIncomingTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	peer := "172.20.0.5"
	meshClient := mockMeshClient{}
	meshClient.externalPeers = []*meshpb.Peer{
		{
			Ip:                "172.20.0.5",
			DoIAllowFileshare: true,
		},
	}

	eventManager := NewEventManager(MockStorage{}, meshClient)
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.EventFunc(fmt.Sprintf(`{
		"type": "RequestReceived",
		"data": {
			"peer": "%s",
			"transfer": "c13c619c-c70b-49b8-9396-72de88155c43",
			"files": [
			  {
				"id": "testfile-small",
				"size": 1048576
			  },
			  {
				"id": "testfile-big",
				"size": 10485760
			  }
			]
		}
	}`, peer))
	transfer, ok := eventManager.transfers["c13c619c-c70b-49b8-9396-72de88155c43"]
	assert.True(t, ok)
	assert.Equal(t, "c13c619c-c70b-49b8-9396-72de88155c43", transfer.Id)
	assert.Equal(t, peer, transfer.Peer)
	assert.Equal(t, pb.Direction_INCOMING, transfer.Direction)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.WithinDuration(t, time.Now(), transfer.Created.AsTime(), time.Second*10)
	assert.Equal(t, 2, len(transfer.Files))
	assert.Equal(t, "", transfer.Path) // Only set after accepting transfer
}

func TestGetTransfers(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	timeNow := time.Now()
	for i := 10; i > 0; i-- {
		eventManager.transfers[strconv.Itoa(i)] = &pb.Transfer{
			Id:      strconv.Itoa(i),
			Created: timestamppb.New(timeNow.Add(-time.Second * time.Duration(i))),
		}
	}
	transfers := eventManager.GetTransfers()
	assert.Equal(t, 10, len(transfers))
	for i := 0; i < 9; i++ {
		assert.True(t, transfers[i].Created.AsTime().Before(transfers[i+1].Created.AsTime()))
	}

	// Test whether we received a copy
	eventManager.transfers[transfers[0].Id].Path = "/test1"
	transfers[0].Path = "/test2"
	assert.Equal(t, "/test1", eventManager.transfers[transfers[0].Id].Path)
}

func TestGetTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }
	eventManager.transfers["test"] = &pb.Transfer{
		Id: "test",
	}

	transfer, err := eventManager.GetTransfer("test")
	assert.NoError(t, err)
	assert.Equal(t, "test", transfer.Id)

	// Test whether we received a copy
	eventManager.transfers["test"].Path = "/test1"
	transfer.Path = "/test1"
	assert.Equal(t, "/test1", eventManager.transfers["test"].Path)
}

func TestOutgoingTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "172.20.0.5", "/tmp")

	transfer, ok := eventManager.transfers["c13c619c-c70b-49b8-9396-72de88155c43"]
	assert.True(t, ok)
	assert.Equal(t, "c13c619c-c70b-49b8-9396-72de88155c43", transfer.Id)
	assert.Equal(t, "172.20.0.5", transfer.Peer)
	assert.Equal(t, pb.Direction_OUTGOING, transfer.Direction)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.WithinDuration(t, time.Now(), transfer.Created.AsTime(), time.Second*10)
	assert.Equal(t, 0, len(transfer.Files)) // Files only added upon confirmation from libdrop
	assert.Equal(t, "/tmp", transfer.Path)

	eventManager.EventFunc(`{
		"type": "RequestQueued",
		"data": {
			"peer": "172.20.0.5",
			"transfer": "c13c619c-c70b-49b8-9396-72de88155c43",
			"files": [
			  {
				"id": "testfile-small",
				"size": 1048576
			  },
			  {
				"id": "testfile-big",
				"size": 10485760
			  }
			]
		}
	}`)
	assert.Equal(t, 2, len(transfer.Files))
}

func TestInvalidTransferProgress(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		eventManager.EventFunc(
			fmt.Sprintf(`{
			"type": "TransferFinished",
			"data": {
				"transfer": "%s",
				"reason": "TransferFailed",
				"data": {
					"status": 3
				}
			}
			}`, transferID))
		waitGroup.Done()
	}()

	waitGroup.Wait()
	testName := "invalid transfer error handling"
	transfer, ok := eventManager.transfers[transferID]
	assert.True(t, ok, testName)
	if ok {
		assert.Equal(t, pb.Status_FINISHED_WITH_ERRORS, transfer.Status, testName)
	}
}

func TestTransferProgress(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	peer := "12.12.12.12"
	path := "/tmp"
	fileCnt := 3
	file1 := "testfile-small"
	file1sz := 100
	file2 := "testfile-big"
	file2sz := 1000
	level2 := "level2"
	file3 := "file3.txt"
	file3sz := 1000

	eventManager.NewOutgoingTransfer(transferID, peer, path)

	eventManager.EventFunc(
		fmt.Sprintf(`{
			"type": "RequestQueued",
			"data": {
				"peer": "%s",
				"transfer": "%s",
				"files": [
				{
					"id": "%s",
					"size": %d,
					"children": {}
				},
				{
					"id": "%s",
					"size": %d,
					"children": {}
				},
				{
					"id": "%s",
					"children": {
						"%s": {
							"id": "%s",
							"size": %d,
							"children": {}
						}
					}
				}
				]
			}
		}`, peer, transferID, file1, file1sz, file2, file2sz, level2, file3, file3, file3sz))

	transfer, ok := eventManager.transfers[transferID]
	assert.True(t, ok)
	assert.Equal(t, transferID, transfer.Id)
	assert.Equal(t, peer, transfer.Peer)
	assert.Equal(t, pb.Direction_OUTGOING, transfer.Direction)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.WithinDuration(t, time.Now(), transfer.Created.AsTime(), time.Second*10)
	assert.Equal(t, path, transfer.Path)
	assert.Equal(t, fileCnt, len(transfer.Files))

	progCh := eventManager.Subscribe(transferID)

	eventManager.EventFunc(
		fmt.Sprintf(`{
		"type": "TransferStarted",
		"data": {
			"transfer": "%s",
			"file": "%s"
		}
		}`, transferID, file1))

	assert.EqualValues(t, transfer.TotalSize, file1sz)

	eventManager.EventFunc(
		fmt.Sprintf(`{
		"type": "TransferStarted",
		"data": {
			"transfer": "%s",
			"file": "%s"
		}
		}`, transferID, file2))

	assert.EqualValues(t, transfer.TotalSize, file1sz+file2sz)

	transferredBytes := file1sz
	go func() {
		eventManager.EventFunc(
			fmt.Sprintf(`{
			"type": "TransferProgress",
			"data": {
				"transfer": "%s",
				"file": "%s",
				"transfered": %d
			}
			}`, transferID, file1, transferredBytes))
	}()

	progressEvent := <-progCh
	assert.Equal(t, pb.Status_ONGOING, progressEvent.Status)
	expectedProgess := uint32(float64(transferredBytes) / float64(file1sz+file2sz) * 100)
	assert.Equal(t, expectedProgess, progressEvent.Transferred)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	go func() {
		eventManager.EventFunc(
			fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileDownloaded",
					"data": {
							"file": "%s"
					}
				}
				}`, transferID, file1))
		eventManager.EventFunc(
			fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileDownloaded",
					"data": {
							"file": "%s"
					}
				}
				}`, transferID, file2))
		eventManager.EventFunc(
			fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileDownloaded",
					"data": {
						"file": "%s"
					}
				}
				}`, transferID, level2+"/"+file3))
		waitGroup.Done()
	}()

	progressEvent = <-progCh
	assert.Equal(t, pb.Status_SUCCESS, progressEvent.Status)

	waitGroup.Wait()
	_, ok = eventManager.transferSubscriptions[transferID]
	assert.False(t, ok) // expect subscriber to be removed
}

func TestAcceptTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		testName       string
		transfer       string
		expectedErr    error
		expectedStatus pb.Status
		files          []string
		sizeLimit      uint64
	}{
		{
			testName:       "accept transfer success",
			transfer:       "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr:    nil,
			expectedStatus: pb.Status_ONGOING,
			files:          []string{},
			sizeLimit:      6,
		},
		{
			testName:       "accept files success",
			transfer:       "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr:    nil,
			expectedStatus: pb.Status_ONGOING,
			files:          []string{"test/file_A"},
			sizeLimit:      1,
		},
		{
			testName:       "transfer doesn't exist",
			transfer:       "invalid_transfer",
			expectedErr:    ErrTransferNotFound,
			expectedStatus: pb.Status_REQUESTED,
			files:          []string{},
			sizeLimit:      6,
		},
		{
			testName:       "file doesn't exist",
			transfer:       "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr:    ErrFileNotFound,
			expectedStatus: pb.Status_REQUESTED,
			files:          []string{"invalid_file"},
			sizeLimit:      6,
		},
		{
			testName:       "size exceeds limit",
			transfer:       "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr:    ErrSizeLimitExceeded,
			expectedStatus: pb.Status_REQUESTED,
			files:          []string{},
			sizeLimit:      5,
		},
		{
			testName:       "partial transfer size exceeds limit",
			transfer:       "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr:    ErrSizeLimitExceeded,
			expectedStatus: pb.Status_REQUESTED,
			files:          []string{"test/file_C"},
			sizeLimit:      2,
		},
	}

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

	for _, test := range tests {
		eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
		eventManager.transfers[transferID] = &pb.Transfer{
			Id:        transferID,
			Direction: pb.Direction_INCOMING,
			Status:    pb.Status_REQUESTED,
			Path:      "/test",
			Files: []*pb.File{
				{Id: "test/file_A", Size: 1},
				{Id: "test/file_B", Size: 2},
				{Id: "test/file_C", Size: 3},
			},
		}
		eventManager.CancelFunc = func(transferID string) error { return nil }

		t.Run(test.testName, func(t *testing.T) {
			transfer, err := eventManager.AcceptTransfer(test.transfer, "/tmp", test.files, test.sizeLimit)

			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedStatus, eventManager.transfers[transferID].Status)

			if transfer != nil {
				assert.Equal(t, "/tmp", transfer.Path)
			}
		})
	}
}

func TestAcceptTransfer_Outgoing(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "172.20.0.5", "/tmp")

	_, err := eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "", []string{}, 999)
	assert.Equal(t, ErrTransferAcceptOutgoing, err)
}

func TestAcceptTransfer_AlreadyAccepted(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.transfers["c13c619c-c70b-49b8-9396-72de88155c43"] = &pb.Transfer{
		Id:        "c13c619c-c70b-49b8-9396-72de88155c43",
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
	}

	_, err := eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "", []string{}, 999)
	assert.NoError(t, err)
	_, err = eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "", []string{}, 999)
	assert.Equal(t, ErrTransferAlreadyAccepted, err)
}

func TestAcceptTransfer_ConcurrentAccepts(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.transfers["c13c619c-c70b-49b8-9396-72de88155c43"] = &pb.Transfer{
		Id:        "c13c619c-c70b-49b8-9396-72de88155c43",
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
	}

	var err1, err2 error
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	go func() {
		_, err1 = eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "", []string{}, 999)
		waitGroup.Done()
	}()
	go func() {
		_, err2 = eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "", []string{}, 999)
		waitGroup.Done()
	}()
	waitGroup.Wait()

	if err1 == nil {
		assert.Equal(t, ErrTransferAlreadyAccepted, err2)
	} else {
		assert.NoError(t, err2)
		assert.Equal(t, ErrTransferAlreadyAccepted, err1)
	}
}

func TestSetTransferStatus(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "172.20.0.5", "/tmp")

	err := eventManager.SetTransferStatus("c13c619c-c70b-49b8-9396-72de88155c43", pb.Status_CANCELED)
	assert.NoError(t, err)
	transfer, err := eventManager.GetTransfer("c13c619c-c70b-49b8-9396-72de88155c43")
	assert.NoError(t, err)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)
}

func TestFinishedTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	finishTests := []struct {
		testName       string
		transferID     string
		peer           string
		reason         string
		transferFileID string
		eventFileID    string
		fileStatus     pb.Status
		transferFound  bool
		transferStatus pb.Status
		disabled       bool
	}{
		{
			testName:       "transfer finished success",
			transferID:     "b537743c-a328-4a3e-b2ec",
			peer:           "1.1.1.1",
			reason:         "FileUploaded",
			transferFileID: "testfile-big",
			eventFileID:    "testfile-big",
			fileStatus:     pb.Status_SUCCESS,
			transferFound:  true,
			transferStatus: pb.Status_SUCCESS,
		},
		{
			testName:       "transfer finished with errors",
			transferID:     "b537743c-a328-4a3e-b2ec",
			peer:           "1.1.1.1",
			reason:         "FileFailed",
			transferFileID: "testfile-big",
			eventFileID:    "testfile-big",
			fileStatus:     pb.Status_BAD_FILE,
			transferFound:  true,
			transferStatus: pb.Status_FINISHED_WITH_ERRORS,
		},
		{
			testName:       "transfer canceled",
			transferID:     "b537743c-a328-4a3e-b2ec",
			peer:           "1.1.1.1",
			reason:         "FileCanceled",
			transferFileID: "testfile-big",
			eventFileID:    "testfile-big",
			fileStatus:     pb.Status_CANCELED,
			transferFound:  true,
			transferStatus: pb.Status_CANCELED,
		},
		{
			testName:       "event has unknown file",
			transferID:     "b537743c-a328-4a3e-b2ec",
			peer:           "1.1.1.1",
			reason:         "FileFailed",
			transferFileID: "testfile-big",
			eventFileID:    "testfile-big-nonono",
			fileStatus:     pb.Status_BAD_FILE,
			transferFound:  true,
			transferStatus: pb.Status_REQUESTED,
		},
	}

	for _, test := range finishTests {
		if test.disabled {
			continue
		}
		t.Run(test.testName, func(t *testing.T) {
			eventManager.transfers[test.transferID] =
				NewIncomingTransfer(test.transferID, test.peer, []*pb.File{{Id: test.transferFileID}})
			eventManager.EventFunc(
				fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "%s",
					"data": {
					  "file": "%s",
					  "by_peer": false,
					  "status": %d
					}
				}
			}`, test.transferID, test.reason, test.eventFileID, test.fileStatus))
			transfer, ok := eventManager.transfers[test.transferID]
			assert.Equal(t, test.transferFound, ok, test.testName)
			assert.Equal(t, test.transferStatus, transfer.Status, test.testName)
		})
	}
}

func TestNewTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	fileID := "file1.xml"

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer(transferID, "172.20.0.5", fileID)

	eventManager.EventFunc(
		fmt.Sprintf(`{
			"type": "RequestQueued",
			"data": {
				"transfer": "%s",
				"files": [
				{
					"id": "%s",
					"size": 1048576
				}
				]
			}
		}`, transferID, fileID))

	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "TransferCanceled",
					"data": {
						"by_peer": false
					}
				}
			}`, transferID))

	transfer, ok := eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)

	transfer.Status = pb.Status_ONGOING
	transfer.Finalized = false
	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileDownloaded",
					"data": {
						"file": "%s",
						"final_path": "testfile-big"
					}
				}
			}`, transferID, fileID))

	transfer, ok = eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, pb.Status_SUCCESS, transfer.Status)

	transfer.Status = pb.Status_ONGOING
	transfer.Finalized = false
	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileUploaded",
					"data": {
						"file": "%s"
					}
				}
			}`, transferID, fileID))

	transfer, ok = eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, pb.Status_SUCCESS, transfer.Status)

	transfer.Status = pb.Status_ONGOING
	transfer.Finalized = false
	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileCanceled",
					"data": {
						"file": "%s",
						"by_peer": true
					}
				}
			}`, transferID, fileID))

	transfer, ok = eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)

	transfer.Status = pb.Status_ONGOING
	transfer.Finalized = false
	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "FileFailed",
					"data": {
						"file": "%s",
						"status": 2
					}
				}
			}`, transferID, fileID))

	transfer, ok = eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, pb.Status_FINISHED_WITH_ERRORS, transfer.Status)

	transfer.Status = pb.Status_ONGOING
	transfer.Finalized = false
	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "TransferFailed",
					"data": {
						"status": 3
					}
				}
			}`, transferID))

	transfer, ok = eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, pb.Status_FINISHED_WITH_ERRORS, transfer.Status)
}

//nolint:errcheck
func TestCheckTransferStatuses_SingleDirWithFiles(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	peer := "2.2.2.2"
	dirID := "tst3"
	file1ID := "file1.xml"
	file2ID := "file2.xml"
	path := "/tmp"

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer(transferID, peer, path)

	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "RequestQueued",
				"data": {
					"transfer": "%s",
					"files": [
						{
							"id": "%s",
							"size": 0,
							"children": {
								"%s": {
									"id": "%s",
									"size": 10,
									"children": {}
								},
								"%s": {
									"id": "%s",
									"size": 20,
									"children": {}
								}
							}
						}
					]
				}
		}`, transferID, dirID, file1ID, file1ID, file2ID, file2ID))

	transfer, ok := eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(2), CountTransferFiles(transfer))
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.Equal(t, uint64(0), transfer.TotalSize) // Size is being appended when downloads for files start

	// TODO: this has to be changed after libdrop fix
	file1ID = dirID + "/" + file1ID
	file2ID = dirID + "/" + file2ID

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetFileStatus(transfer.Files, file1ID, pb.Status_REQUESTED)
	SetFileStatus(transfer.Files, file2ID, pb.Status_REQUESTED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)

	// check canceled
	SetFileStatus(transfer.Files, file1ID, pb.Status_CANCELED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_CANCELED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetFileStatus(transfer.Files, file1ID, pb.Status_REQUESTED)
	SetFileStatus(transfer.Files, file2ID, pb.Status_REQUESTED)

	// check finished success
	SetFileStatus(transfer.Files, file1ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_SUCCESS, transfer.Status)

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetFileStatus(transfer.Files, file1ID, pb.Status_REQUESTED)
	SetFileStatus(transfer.Files, file2ID, pb.Status_REQUESTED)

	// check finished with errors
	SetFileStatus(transfer.Files, file1ID, pb.Status_BAD_FILE)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_FINISHED_WITH_ERRORS, transfer.Status)
}

//nolint:errcheck
func TestCheckTransferStatuses_MultipleInputPaths(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	peer := "2.2.2.2"
	file1ID := "file1.xml"
	file2ID := "file2.xml"
	path := "/tmp"

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer(transferID, peer, path)

	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "RequestQueued",
				"data": {
					"transfer": "%s",
					"files": [
						{
							"id": "%s",
							"size": 100,
							"children": {}
						},
						{
							"id": "%s",
							"size": 200,
							"children": {}
						}
					]
				}
		}`, transferID, file1ID, file2ID))

	transfer, ok := eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, len(transfer.Files))
	assert.Equal(t, 0, len(transfer.Files[0].Children))
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.Equal(t, uint64(0), transfer.TotalSize) // Size is being appended when downloads for files start

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetFileStatus(transfer.Files, file1ID, pb.Status_REQUESTED)
	SetFileStatus(transfer.Files, file2ID, pb.Status_REQUESTED)

	// check canceled
	SetFileStatus(transfer.Files, file1ID, pb.Status_CANCELED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_CANCELED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetFileStatus(transfer.Files, file1ID, pb.Status_REQUESTED)
	SetFileStatus(transfer.Files, file2ID, pb.Status_REQUESTED)

	// check finished success
	SetFileStatus(transfer.Files, file1ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_SUCCESS, transfer.Status)

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetFileStatus(transfer.Files, file1ID, pb.Status_REQUESTED)
	SetFileStatus(transfer.Files, file2ID, pb.Status_REQUESTED)

	// check finished with errors
	SetFileStatus(transfer.Files, file1ID, pb.Status_BAD_FILE)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_FINISHED_WITH_ERRORS, transfer.Status)
}

//nolint:errcheck
func TestCheckTransferStatuses_MultilevelDirComplexStructure(t *testing.T) {
	category.Set(t, category.Unit)

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	peer := "2.2.2.2"
	file1ID := "file1.xml"
	file2ID := "file2.xml"
	path := "/tmp"
	topLevelID := "multilevel"
	level1ID := "level1"
	level1File1ID := "level1-file1.txt"
	level1File2ID := "level1-file2.txt"
	level2ID := "level2"
	level2File1ID := "level2-file1.txt"
	level2File2ID := "level2-file2.txt"
	level3ID := "level3"
	level3File1ID := "level3-file1.txt"
	level3File2ID := "level3-file2.txt"

	meshClient := mockMeshClient{}
	meshClient.externalPeers = []*meshpb.Peer{
		{
			Ip:                peer,
			DoIAllowFileshare: true,
		},
	}

	eventManager := NewEventManager(MockStorage{}, meshClient)
	eventManager.CancelFunc = func(transferID string) error { return nil }

	eventManager.NewOutgoingTransfer(transferID, peer, path)

	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "RequestQueued",
				"data": {
					"transfer": "%s",
					"files": [
						{
							"id": "%s",
							"size": 1,
							"children": {}
						},
						{
							"id": "%s",
							"size": 1,
							"children": {}
						},
						{
							"id": "%s",
							"size": 0,
							"children": {
								"%s": {
									"id": "%s",
									"size": 0,
									"children": {
										"%s": {
											"id": "%s",
											"size": 1,
											"children": {}
										},
										"%s": {
											"id": "%s",
											"size": 1,
											"children": {}
										},
										"%s": {
											"id": "%s",
											"size": 0,
											"children": {
												"%s": {
													"id": "%s",
													"size": 1,
													"children": {}
												},
												"%s": {
													"id": "%s",
													"size": 1,
													"children": {}
												},
												"%s": {
													"id": "%s",
													"size": 0,
													"children": {
														"%s": {
															"id": "%s",
															"size": 1,
															"children": {}
														},
														"%s": {
															"id": "%s",
															"size": 1,
															"children": {}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					]
				}
		}`, transferID, file1ID, file2ID, topLevelID,
			level1ID, level1ID, level1File1ID, level1File1ID, level1File2ID, level1File2ID,
			level2ID, level2ID, level2File1ID, level2File1ID, level2File2ID, level2File2ID,
			level3ID, level3ID, level3File1ID, level3File1ID, level3File2ID, level3File2ID))

	transfer, ok := eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, 3, len(transfer.Files))
	assert.Equal(t, 0, len(transfer.Files[0].Children))
	assert.Equal(t, 1, len(transfer.Files[2].Children))
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.Equal(t, uint64(0), transfer.TotalSize) // Size is being appended when downloads for files start

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetTransferAllFileStatus(transfer, pb.Status_REQUESTED)

	// TODO: this has to be changed after libdrop fix
	level1File1ID = topLevelID + "/" + level1ID + "/" + level1File1ID
	level1File2ID = topLevelID + "/" + level1ID + "/" + level1File2ID
	level2File1ID = topLevelID + "/" + level1ID + "/" + level2ID + "/" + level2File1ID
	level2File2ID = topLevelID + "/" + level1ID + "/" + level2ID + "/" + level2File2ID
	level3File1ID = topLevelID + "/" + level1ID + "/" + level2ID + "/" + level3ID + "/" + level3File1ID
	level3File2ID = topLevelID + "/" + level1ID + "/" + level2ID + "/" + level3ID + "/" + level3File2ID

	// check canceled
	SetFileStatus(transfer.Files, file1ID, pb.Status_CANCELED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_CANCELED)
	SetFileStatus(transfer.Files, level1File1ID, pb.Status_CANCELED)
	SetFileStatus(transfer.Files, level1File2ID, pb.Status_CANCELED)
	SetFileStatus(transfer.Files, level2File1ID, pb.Status_CANCELED)
	SetFileStatus(transfer.Files, level2File2ID, pb.Status_CANCELED)
	SetFileStatus(transfer.Files, level3File1ID, pb.Status_CANCELED)
	SetFileStatus(transfer.Files, level3File2ID, pb.Status_CANCELED)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetTransferAllFileStatus(transfer, pb.Status_REQUESTED)

	// check finished success
	SetFileStatus(transfer.Files, file1ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level1File1ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level1File2ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level2File1ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level2File2ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level3File1ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level3File2ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_SUCCESS, transfer.Status)

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetTransferAllFileStatus(transfer, pb.Status_REQUESTED)

	// check finished with errors
	SetFileStatus(transfer.Files, file1ID, pb.Status_BAD_FILE)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	SetFileStatus(transfer.Files, file2ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level1File1ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level1File2ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level2File1ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level2File2ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level3File1ID, pb.Status_SUCCESS)
	SetFileStatus(transfer.Files, level3File2ID, pb.Status_SUCCESS)
	transfer.Status = GetNewTransferStatus(transfer.Files, transfer.Status)
	assert.Equal(t, pb.Status_FINISHED_WITH_ERRORS, transfer.Status)
}

func TestTransferRequestPermissionsValidation(t *testing.T) {
	meshClient := mockMeshClient{}
	noPermissionPeer := "172.10.0.5"
	meshClient.externalPeers = []*meshpb.Peer{
		{
			Ip:                noPermissionPeer,
			Pubkey:            "aZ9KwmEzystVJ0R1YitV02NzNngmSrZ3JDTj6tkI8T6=",
			Hostname:          "internal.peer1.nord",
			DoIAllowFileshare: false,
		},
	}

	eventManager := NewEventManager(MockStorage{}, meshClient)
	eventManager.CancelFunc = func(transferID string) error { return nil }

	tests := []struct {
		testName string
		peer     string
	}{
		{
			testName: "peer has no send permissions",
			peer:     noPermissionPeer,
		},
		{
			testName: "unknown peer",
			peer:     "1.2.3.4",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			eventManager.EventFunc(fmt.Sprintf(`{
				"type": "RequestReceived",
				"data": {
					"peer": "%s",
					"transfer": "c13c619c-c70b-49b8-9396-72de88155c43",
					"files": [
					{
						"id": "testfile",
						"size": 1
					}
					]
				}
			}`, test.peer))
			assert.True(t, len(eventManager.transfers) == 0)
		})
	}
}

func TestTransferFinalization(t *testing.T) {
	fileUploadedEventFormat := `{
		"type": "TransferFinished",
		"data": {
			"transfer": "c13c619c-c70b-49b8-9396-72de88155c43",
			"reason": "%s",
			"data": {
				"file": "%s",
				"status": %d
			}
		}
	}`

	transferCanceledEvent := `{
		"type": "TransferFinished",
		"data": {
			"transfer": "c13c619c-c70b-49b8-9396-72de88155c43",
			"reason": "TransferCanceled",
			"data": {}
		}
	}`

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

	file1 := "file1"
	file2 := "file2"
	file3 := "file3"

	tests := []struct {
		testName                string
		transferFinishedReasons [3]string
		fileStatuses            [3]int
		file1Status             string
		file2Status             string
		file3Status             string
		finalStatus             pb.Status
	}{
		{
			testName:                "all accepted",
			transferFinishedReasons: [3]string{"FileUploaded", "FileUploaded", "FileUploaded"},
			fileStatuses:            [3]int{int(pb.Status_SUCCESS), int(pb.Status_SUCCESS), int(pb.Status_SUCCESS)},
			finalStatus:             pb.Status_SUCCESS,
		},
		{
			testName:                "one canceled",
			transferFinishedReasons: [3]string{"FileUploaded", "FileCanceled", "FileUploaded"},
			fileStatuses:            [3]int{int(pb.Status_SUCCESS), int(pb.Status_CANCELED), int(pb.Status_SUCCESS)},
			finalStatus:             pb.Status_SUCCESS,
		},
		{
			testName:                "all canceled",
			transferFinishedReasons: [3]string{"FileCanceled", "FileCanceled", "FileCanceled"},
			fileStatuses:            [3]int{int(pb.Status_CANCELED), int(pb.Status_CANCELED), int(pb.Status_CANCELED)},
			finalStatus:             pb.Status_CANCELED,
		},
		{
			testName:                "one failed",
			transferFinishedReasons: [3]string{"FileUploaded", "FileFailed", "FileUploaded"},
			fileStatuses:            [3]int{int(pb.Status_SUCCESS), int(pb.Status_IO), int(pb.Status_SUCCESS)},
			finalStatus:             pb.Status_FINISHED_WITH_ERRORS,
		},
	}

	for _, test := range tests {
		eventManager := NewEventManager(MockStorage{}, mockMeshClient{})

		cancelFuncCalled := false
		eventManager.CancelFunc = func(transferID string) error {
			cancelFuncCalled = true
			return nil
		}

		eventManager.transfers[transferID] = &pb.Transfer{
			Id:     transferID,
			Status: pb.Status_ONGOING,
			Files: []*pb.File{
				{Id: file1, Status: pb.Status_ONGOING},
				{Id: file2, Status: pb.Status_ONGOING},
				{Id: file3, Status: pb.Status_ONGOING}},
			TotalSize:        3,
			TotalTransferred: 0,
			Direction:        pb.Direction_INCOMING,
		}

		t.Run(test.testName, func(t *testing.T) {
			file1UploadedEvent := fmt.Sprintf(fileUploadedEventFormat, test.transferFinishedReasons[0], file1, test.fileStatuses[0])
			eventManager.EventFunc(file1UploadedEvent)
			assert.False(t, cancelFuncCalled, "transfer has been finalized(canceled) before it has finished")
			assert.Equal(t, pb.Status_ONGOING, eventManager.transfers[transferID].Status,
				"expected transfer status: %s, acctual transfer status: %s",
				test.finalStatus, eventManager.transfers[transferID].Status)

			file2UploadedEvent := fmt.Sprintf(fileUploadedEventFormat, test.transferFinishedReasons[1], file2, test.fileStatuses[1])
			eventManager.EventFunc(file2UploadedEvent)
			assert.False(t, cancelFuncCalled, "transfer has been finalized(canceled) before it has finished")
			assert.Equal(t, pb.Status_ONGOING, eventManager.transfers[transferID].Status,
				"expected transfer status: %s, acctual transfer status: %s",
				test.finalStatus, eventManager.transfers[transferID].Status)

			file3UploadedEvent := fmt.Sprintf(fileUploadedEventFormat, test.transferFinishedReasons[2], file3, test.fileStatuses[2])
			eventManager.EventFunc(file3UploadedEvent)
			assert.True(t, cancelFuncCalled, "transfer was not finalized(canceled) after it has finished")
			assert.Equal(t, test.finalStatus, eventManager.transfers[transferID].Status,
				"expected transfer status: %s, acctual transfer status: %s",
				test.finalStatus, eventManager.transfers[transferID].Status)

			cancelFuncCalled = false

			eventManager.EventFunc(transferCanceledEvent)

			assert.False(t, cancelFuncCalled, "transfer has been finalized(canceled) twice")
			assert.Equal(t, test.finalStatus, eventManager.transfers[transferID].Status,
				"expected transfer status: %s, acctual transfer status: %s",
				test.finalStatus, eventManager.transfers[transferID].Status)
		})
	}
}

func TestTransferFinalization_TransferCanceled(t *testing.T) {
	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

	transferCanceledEvent := fmt.Sprintf(`{
		"type": "TransferFinished",
		"data": {
			"transfer": "%s",
			"reason": "TransferCanceled",
			"data": {}
		}
	}`, transferID)

	eventManager := NewEventManager(MockStorage{}, mockMeshClient{})
	cancelFuncCalled := false
	eventManager.CancelFunc = func(transferID string) error {
		cancelFuncCalled = true
		return nil
	}

	eventManager.transfers[transferID] = &pb.Transfer{
		Id:     transferID,
		Status: pb.Status_ONGOING,
		Files: []*pb.File{
			{Id: "file1", Status: pb.Status_ONGOING},
			{Id: "file2", Status: pb.Status_ONGOING},
			{Id: "file3", Status: pb.Status_ONGOING}},
		TotalSize:        3,
		TotalTransferred: 0,
		Direction:        pb.Direction_INCOMING,
	}

	eventManager.EventFunc(transferCanceledEvent)
	assert.False(t, cancelFuncCalled, "canceled transfer has been finalized")
	assert.Equal(t, pb.Status_CANCELED, eventManager.transfers[transferID].Status,
		"expected transfer status: %s, acctual transfer status: %s",
		pb.Status_CANCELED, eventManager.transfers[transferID].Status)
}
