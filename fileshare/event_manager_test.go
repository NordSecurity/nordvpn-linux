package fileshare

import (
	"fmt"
	"io/fs"
	"math"
	"net/netip"
	"os"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"golang.org/x/sys/unix"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockNotification struct {
	id      uint32
	summary string
	body    string
	actions []Action
}

type mockNotifier struct {
	notifications []mockNotification
	nextID        uint32
}

func (mn *mockNotifier) SendNotification(summary string, body string, actions []Action) (uint32, error) {
	notificationID := mn.nextID
	mn.notifications = append(mn.notifications, mockNotification{id: notificationID, summary: summary, body: body, actions: actions})
	mn.nextID++
	return notificationID, nil
}

func (mn *mockNotifier) Close() error {
	return nil
}

func (mn *mockNotifier) getLastNotification() mockNotification {
	return mn.notifications[len(mn.notifications)-1]
}

func NewMockNotificationManager(osInfo *mockEventManagerOsInfo) NotificationManager {
	return NotificationManager{
		notifications: newNotificationStorage(),
	}
}

type mockEventManagerFilesystem struct {
	fstest.MapFS
	freeSpace uint64
}

func (mf mockEventManagerFilesystem) Lstat(path string) (fs.FileInfo, error) {
	fileInfo, err := mf.MapFS.Stat(path)
	return fileInfo, err
}

func (mf mockEventManagerFilesystem) Statfs(path string) (unix.Statfs_t, error) {
	return unix.Statfs_t{Bavail: mf.freeSpace, Bsize: 1}, nil
}

type mockEventManagerFileshare struct {
	canceledTransferIDs []string
	acceptedTransferIDS []string
}

// Enable starts service listening at provided address
func (*mockEventManagerFileshare) Enable(listenAddress netip.Addr) error {
	return nil
}

// Disable tears down fileshare service
func (*mockEventManagerFileshare) Disable() error {
	return nil
}

// Send sends the provided file or dir to provided peer and returns transfer ID
func (*mockEventManagerFileshare) Send(peer netip.Addr, paths []string) (string, error) {
	return "", nil
}

// Accept accepts provided files from provided request and starts download process
func (mfs *mockEventManagerFileshare) Accept(transferID, dstPath string, fileID string) error {
	mfs.acceptedTransferIDS = append(mfs.acceptedTransferIDS, transferID)
	return nil
}

// Cancel file transfer by ID.
func (mfs *mockEventManagerFileshare) Cancel(transferID string) error {
	mfs.canceledTransferIDs = append(mfs.canceledTransferIDs, transferID)
	return nil
}

// CancelFile id in a transfer
func (*mockEventManagerFileshare) CancelFile(transferID string, fileID string) error {
	return nil
}

func (mfs *mockEventManagerFileshare) getLastAcceptedTransferID() string {
	length := len(mfs.acceptedTransferIDS)
	if length == 0 {
		return ""
	}

	return mfs.acceptedTransferIDS[length-1]
}

func (mfs *mockEventManagerFileshare) getLastCanceledTransferID() string {
	length := len(mfs.canceledTransferIDs)
	if length == 0 {
		return ""
	}

	return mfs.canceledTransferIDs[length-1]
}

type mockEventManagerOsInfo struct {
	currentUser user.User
	groupIds    map[string][]string
}

func (mOS *mockEventManagerOsInfo) CurrentUser() (*user.User, error) {
	return &mOS.currentUser, nil
}

func (mOS *mockEventManagerOsInfo) GetGroupIds(userInfo *user.User) ([]string, error) {
	return mOS.groupIds[userInfo.Uid], nil
}

type mockFileshare struct{}

func (mfs *mockFileshare) Enable(listenAddress netip.Addr) error { return nil }

func (mfs *mockFileshare) Disable() error { return nil }

func (mfs *mockFileshare) Send(peer netip.Addr, paths []string) (string, error) { return "", nil }

func (mfs *mockFileshare) Accept(transferID, dstPath string, fileID string) error { return nil }

func (mfs *mockFileshare) Cancel(transferID string) error { return nil }

func (mfs *mockFileshare) CancelFile(transferID string, fileID string) error { return nil }

type mockSystemEnvironment struct {
	mockEventManagerOsInfo
	mockEventManagerFilesystem
	destinationDirectory string
	currentUserUID       uint32
	currentUserGID       uint32
}

func newMockSystemEnvironment(t *testing.T) mockSystemEnvironment {
	t.Helper()

	currentUserUID := uint32(1000)
	currentUSerUIDString := strconv.Itoa(int(currentUserUID))
	currentUserGID := uint32(1000)
	currentUserGIDString := strconv.Itoa(int(currentUserGID))

	stat_t := &syscall.Stat_t{
		Uid: currentUserUID,
		Gid: currentUserGID,
	}

	destinationDirectoryFilename := "tmp"
	directories := fstest.MapFS{
		destinationDirectoryFilename: &fstest.MapFile{Mode: os.ModeDir | 0777, Sys: stat_t},
	}

	osInfo := mockEventManagerOsInfo{
		currentUser: user.User{Uid: currentUSerUIDString},
		groupIds: map[string][]string{
			currentUSerUIDString: {currentUserGIDString},
		},
	}

	filesystem := mockEventManagerFilesystem{
		freeSpace: math.MaxUint64,
		MapFS:     directories,
	}

	return mockSystemEnvironment{
		mockEventManagerOsInfo:     osInfo,
		mockEventManagerFilesystem: filesystem,
		destinationDirectory:       destinationDirectoryFilename,
		currentUserUID:             currentUserUID,
		currentUserGID:             currentUserGID,
	}
}

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

	eventManager := NewEventManager(false, NoopStorage{}, &meshClient, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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

	eventManager := NewEventManager(false, NoopStorage{}, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.notificationManager = nil
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

	eventManager := NewEventManager(false, NoopStorage{}, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})
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

	eventManager := NewEventManager(false, NoopStorage{}, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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

	eventManager := NewEventManager(false, NoopStorage{}, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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

	eventManager := NewEventManager(false, NoopStorage{}, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	peer := "12.12.12.12"
	path := "/tmp"
	fileCnt := 3
	file1 := "testfile-small"
	file1ID := "file1ID"
	file1sz := 100
	file2 := "testfile-big"
	file2ID := "file2ID"
	file2sz := 1000
	file3 := "file3.txt"
	file3ID := "file3ID"
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
					"path": "%s",
					"size": %d
				},
				{
					"id": "%s",
					"path": "%s",
					"size": %d
				},
				{
					"id": "%s",
					"path": "%s",
					"size": %d
				}
				]
			}
		}`, peer, transferID, file1ID, file1, file1sz, file2ID, file2, file2sz, file3ID, file3, file3sz))

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
		}`, transferID, file1ID))

	assert.EqualValues(t, transfer.TotalSize, file1sz)

	eventManager.EventFunc(
		fmt.Sprintf(`{
		"type": "TransferStarted",
		"data": {
			"transfer": "%s",
			"file": "%s"
		}
		}`, transferID, file2ID))

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
			}`, transferID, file1ID, transferredBytes))
	}()

	progressEvent := <-progCh
	assert.Equal(t, pb.Status_ONGOING, progressEvent.Status)
	expectedProgress := uint32(float64(transferredBytes) / float64(file1sz+file2sz) * 100)
	assert.Equal(t, expectedProgress, progressEvent.Transferred)

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
				}`, transferID, file1ID))
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
				}`, transferID, file2ID))
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
				}`, transferID, file3ID))
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
		// number of available blocks of size 1
		sizeLimit uint64
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

	mockSystemEnvironment := newMockSystemEnvironment(t)

	for _, test := range tests {
		mockSystemEnvironment.mockEventManagerFilesystem.freeSpace = test.sizeLimit

		eventManager := NewEventManager(false, NoopStorage{},
			&mockMeshClient{},
			&mockSystemEnvironment.mockEventManagerOsInfo,
			&mockSystemEnvironment.mockEventManagerFilesystem,
			"")
		eventManager.notificationManager = nil
		eventManager.transfers[transferID] = &pb.Transfer{
			Id:        transferID,
			Direction: pb.Direction_INCOMING,
			Status:    pb.Status_REQUESTED,
			Path:      "/test",
			Files: []*pb.File{
				{Path: "test/file_A", Id: "fileA", Size: 1},
				{Path: "test/file_B", Id: "fileB", Size: 2},
				{Path: "test/file_C", Id: "fileC", Size: 3},
			},
		}
		eventManager.SetFileshare(&mockEventManagerFileshare{})

		t.Run(test.testName, func(t *testing.T) {
			transfer, err := eventManager.AcceptTransfer(test.transfer, mockSystemEnvironment.destinationDirectory, test.files)

			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expectedStatus, eventManager.transfers[transferID].Status)

			if transfer != nil {
				assert.Equal(t, mockSystemEnvironment.destinationDirectory, transfer.Path)
			}
		})
	}
}

func TestAcceptTransfer_Outgoing(t *testing.T) {
	category.Set(t, category.Unit)

	mockSystemEnvironment := newMockSystemEnvironment(t)

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockSystemEnvironment.mockEventManagerOsInfo,
		&mockSystemEnvironment.mockEventManagerFilesystem,
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	eventManager.NewOutgoingTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "172.20.0.5", "/tmp")

	_, err := eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", mockSystemEnvironment.destinationDirectory, []string{})
	assert.Equal(t, ErrTransferAcceptOutgoing, err)
}

func TestAcceptTransfer_AlreadyAccepted(t *testing.T) {
	category.Set(t, category.Unit)

	mockSystemEnvironment := newMockSystemEnvironment(t)

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockSystemEnvironment.mockEventManagerOsInfo,
		&mockSystemEnvironment.mockEventManagerFilesystem,
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	eventManager.transfers["c13c619c-c70b-49b8-9396-72de88155c43"] = &pb.Transfer{
		Id:        "c13c619c-c70b-49b8-9396-72de88155c43",
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
	}

	_, err := eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", mockSystemEnvironment.destinationDirectory, []string{})
	assert.NoError(t, err)
	_, err = eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", mockSystemEnvironment.destinationDirectory, []string{})
	assert.Equal(t, ErrTransferAlreadyAccepted, err)
}

func TestAcceptTransfer_ConcurrentAccepts(t *testing.T) {
	category.Set(t, category.Unit)

	mockSystemEnvironment := newMockSystemEnvironment(t)

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockSystemEnvironment.mockEventManagerOsInfo,
		&mockSystemEnvironment.mockEventManagerFilesystem,
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	eventManager.transfers["c13c619c-c70b-49b8-9396-72de88155c43"] = &pb.Transfer{
		Id:        "c13c619c-c70b-49b8-9396-72de88155c43",
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
	}

	var err1, err2 error
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)
	go func() {
		_, err1 = eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", mockSystemEnvironment.destinationDirectory, []string{})
		waitGroup.Done()
	}()
	go func() {
		_, err2 = eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", mockSystemEnvironment.destinationDirectory, []string{})
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

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	eventManager.NewOutgoingTransfer("c13c619c-c70b-49b8-9396-72de88155c43", "172.20.0.5", "/tmp")

	err := eventManager.SetTransferStatus("c13c619c-c70b-49b8-9396-72de88155c43", pb.Status_CANCELED)
	assert.NoError(t, err)
	transfer, err := eventManager.GetTransfer("c13c619c-c70b-49b8-9396-72de88155c43")
	assert.NoError(t, err)
	assert.Equal(t, pb.Status_CANCELED, transfer.Status)
}

func TestFinishedTransfer(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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
	file1Path := "tst3/file1.xml"
	file1ID := "file1"
	file2Path := "tst3/file2.xml"
	file2ID := "file2"
	path := "/tmp"

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	eventManager.NewOutgoingTransfer(transferID, peer, path)

	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "RequestQueued",
				"data": {
					"transfer": "%s",
					"files": [
						{
							"id": "%s",
							"path": "%s",
							"size": 10
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 20
						}
					]
				}
		}`, transferID, file1ID, file1Path, file2ID, file2Path))

	transfer, ok := eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, 2, len(transfer.Files))
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.Equal(t, uint64(0), transfer.TotalSize) // Size is being appended when downloads for files start

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

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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
	file1ID := "file1"
	file1Path := "file1.xml"
	file2ID := "file2"
	file2Path := "file2.xml"
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

	level1File1Path := topLevelID + "/" + level1ID + "/" + level1File1ID
	level1File2Path := topLevelID + "/" + level1ID + "/" + level1File2ID
	level2File1Path := topLevelID + "/" + level1ID + "/" + level2ID + "/" + level2File1ID
	level2File2Path := topLevelID + "/" + level1ID + "/" + level2ID + "/" + level2File2ID
	level3File1Path := topLevelID + "/" + level1ID + "/" + level2ID + "/" + level3ID + "/" + level3File1ID
	level3File2Path := topLevelID + "/" + level1ID + "/" + level2ID + "/" + level3ID + "/" + level3File2ID

	meshClient := mockMeshClient{}
	meshClient.externalPeers = []*meshpb.Peer{
		{
			Ip:                peer,
			DoIAllowFileshare: true,
		},
	}

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	eventManager.NewOutgoingTransfer(transferID, peer, path)

	eventManager.EventFunc(
		fmt.Sprintf(`{
				"type": "RequestQueued",
				"data": {
					"transfer": "%s",
					"files": [
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						},
						{
							"id": "%s",
							"path": "%s",
							"size": 1
						}
					]
				}
		}`, transferID, file1ID, file1Path, file2ID, file2Path,
			level1File1ID, level1File1Path, level1File2ID, level1File2Path,
			level2File1ID, level2File1Path, level2File2ID, level2File2Path,
			level3File1ID, level3File1Path, level3File2ID, level3File2Path))

	transfer, ok := eventManager.transfers[transferID]
	assert.Equal(t, true, ok)
	assert.Equal(t, 8, len(transfer.Files))
	assert.Equal(t, pb.Status_REQUESTED, transfer.Status)
	assert.Equal(t, uint64(0), transfer.TotalSize) // Size is being appended when downloads for files start

	// init transfer
	transfer.Status = pb.Status_REQUESTED
	SetTransferAllFileStatus(transfer, pb.Status_REQUESTED)

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

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	eventManager.SetFileshare(&mockEventManagerFileshare{})

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
		eventManager := NewEventManager(false, NoopStorage{},
			&mockMeshClient{},
			&mockEventManagerOsInfo{},
			&mockEventManagerFilesystem{},
			"")
		eventManager.notificationManager = nil
		mockFileshare := mockEventManagerFileshare{}
		eventManager.SetFileshare(&mockFileshare)

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

			assert.Empty(t, mockFileshare.canceledTransferIDs, "Transfer has been finalized(canceled) before it has finished")
			assert.Equal(t, pb.Status_ONGOING, eventManager.transfers[transferID].Status,
				"Invalid transfer status.")

			file2UploadedEvent := fmt.Sprintf(fileUploadedEventFormat, test.transferFinishedReasons[1], file2, test.fileStatuses[1])
			eventManager.EventFunc(file2UploadedEvent)

			assert.Empty(t, mockFileshare.canceledTransferIDs, "Transfer has been finalized(canceled) before it has finished")
			assert.Equal(t, pb.Status_ONGOING, eventManager.transfers[transferID].Status,
				"Invalid transfer status")

			file3UploadedEvent := fmt.Sprintf(fileUploadedEventFormat, test.transferFinishedReasons[2], file3, test.fileStatuses[2])
			eventManager.EventFunc(file3UploadedEvent)

			assert.Equal(t, transferID, mockFileshare.getLastCanceledTransferID(), "Transfer was not finalized(canceled) after it has finished")
			assert.Equal(t, test.finalStatus, eventManager.transfers[transferID].Status,
				"Invalid transfer status")

			eventManager.EventFunc(transferCanceledEvent)

			assert.Equal(t, len(mockFileshare.canceledTransferIDs), 1, "Transfer has been finalized(canceled) twice")
			assert.Equal(t, test.finalStatus, eventManager.transfers[transferID].Status,
				"Invalid transfer status")
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

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = nil
	mockFileshare := mockEventManagerFileshare{}
	eventManager.SetFileshare(&mockFileshare)

	eventManager.transfers[transferID] = &pb.Transfer{
		Id:     transferID,
		Status: pb.Status_ONGOING,
		Files: []*pb.File{
			{Id: "file1", Status: pb.Status_ONGOING},
			{Id: "file2", Status: pb.Status_ONGOING},
			{Id: "file3", Status: pb.Status_SUCCESS}},
		TotalSize:        3,
		TotalTransferred: 0,
		Direction:        pb.Direction_INCOMING,
	}

	eventManager.EventFunc(transferCanceledEvent)
	assert.Empty(t, mockFileshare.canceledTransferIDs, "Canceled transfer has been finalized")
	assert.Equal(t, pb.Status_CANCELED, eventManager.transfers[transferID].Status,
		"Invalid transfer status")
}

func TestTransferFinishedNotifications(t *testing.T) {
	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	fileID := "file_id"
	filePath := "file_path"

	initializeEventManager := func(direction pb.Direction) (*EventManager, *mockNotifier) {
		notifier := mockNotifier{
			notifications: []mockNotification{},
			nextID:        0,
		}
		notificationManager := NewMockNotificationManager(&mockEventManagerOsInfo{})
		notificationManager.notifier = &notifier

		eventManager := NewEventManager(false, NoopStorage{},
			&mockMeshClient{},
			&mockEventManagerOsInfo{},
			&mockEventManagerFilesystem{},
			"")
		eventManager.notificationManager = &notificationManager
		eventManager.SetFileshare(&mockFileshare{})
		eventManager.transfers[transferID] = &pb.Transfer{
			Id:     transferID,
			Status: pb.Status_ONGOING,
			Files: []*pb.File{
				{Id: fileID, Path: filePath, Status: pb.Status_ONGOING},
			},
			TotalSize:        1,
			TotalTransferred: 0,
			Direction:        direction,
		}

		return eventManager, &notifier
	}

	tests := []struct {
		name            string
		status          pb.Status
		direction       pb.Direction
		reason          string
		expectedSummary string
		expectedBody    string
		expectedActions []Action
	}{
		{
			name:            "download finished success",
			status:          pb.Status_SUCCESS,
			direction:       pb.Direction_INCOMING,
			reason:          "FileDownloaded",
			expectedSummary: "downloaded",
			expectedActions: []Action{{actionKeyOpenFile, "Open"}},
		},
		{
			name:            "download finished failure",
			status:          pb.Status_TRANSPORT,
			direction:       pb.Direction_INCOMING,
			reason:          "FileFailed",
			expectedSummary: "transport problem",
			expectedActions: nil,
		},
		{
			name:            "download canceled",
			status:          pb.Status_CANCELED,
			direction:       pb.Direction_INCOMING,
			reason:          "FileCanceled",
			expectedSummary: "canceled",
			expectedActions: nil,
		},
		{
			name:            "upload finished success",
			status:          pb.Status_SUCCESS,
			direction:       pb.Direction_OUTGOING,
			reason:          "FileUploaded",
			expectedSummary: "uploaded",
			expectedActions: nil,
		},
	}

	for _, test := range tests {
		eventManager, notifier := initializeEventManager(test.direction)

		t.Run(test.name, func(t *testing.T) {
			eventManager.EventFunc(fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "%s",
					"data": {
						"file": "%s",
						"status": %d
					}
				}
			}`, transferID, test.reason, fileID, test.status))

			assert.Equal(t, 1, len(notifier.notifications),
				"TransferFinished event was received, but EventManager did not send any notifications.")

			notification := notifier.notifications[0]

			assert.Equal(t, test.expectedSummary, notification.summary,
				"Invalid notification summary")
			assert.Equal(t, filePath, notification.body,
				"Notification body should be a filename")
			assert.Equal(t, test.expectedActions, notification.actions,
				"Actions associated with notifications are invalid.")
		})
	}
}

func TestTransferFinishedNotificationsOpenFile(t *testing.T) {
	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	fileID := "file_id"
	filePath := "file_path"

	notifier := mockNotifier{
		notifications: []mockNotification{},
		nextID:        0,
	}

	openedFiles := []string{}
	openFileFunc := func(filename string) {
		openedFiles = append(openedFiles, filename)
	}

	notificationManager := NewMockNotificationManager(&mockEventManagerOsInfo{})
	notificationManager.notifier = &notifier
	notificationManager.openFileFunc = openFileFunc

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = &notificationManager
	eventManager.SetFileshare(&mockEventManagerFileshare{})
	eventManager.transfers[transferID] = &pb.Transfer{
		Id:     transferID,
		Status: pb.Status_ONGOING,
		Files: []*pb.File{
			{Id: fileID, Path: filePath, Status: pb.Status_ONGOING},
		},
		TotalSize:        1,
		TotalTransferred: 0,
		Direction:        pb.Direction_INCOMING,
	}

	eventManager.EventFunc(fmt.Sprintf(`{
		"type": "TransferFinished",
		"data": {
			"transfer": "%s",
			"reason": "FileDownloaded",
			"data": {
				"file": "%s",
				"status": %d
			}
		}
	}`, transferID, fileID, pb.Status_SUCCESS))

	notification := notifier.notifications[0]

	notificationManager.OpenFile(notification.id)
	assert.Equal(t, 1, len(openedFiles), "Open event was emitted, but no files were opened.")
	assert.Equal(t, filePath, openedFiles[0], "Invalid file opened.")

	notificationManager.OpenFile(notification.id)
	assert.Equal(t, 1, len(openedFiles), "File was opened but it was already opened once.")
}

func TestTransferRequestNotification(t *testing.T) {
	notifier := mockNotifier{
		notifications: []mockNotification{},
		nextID:        0,
	}

	openedFiles := []string{}
	openFileFunc := func(filename string) {
		openedFiles = append(openedFiles, filename)
	}

	notificationManager := NewMockNotificationManager(&mockEventManagerOsInfo{})
	notificationManager.notifier = &notifier
	notificationManager.openFileFunc = openFileFunc

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = &notificationManager
	eventManager.SetFileshare(&mockEventManagerFileshare{})

	peer := "172.20.0.5"
	hostname := "peer.nord"
	eventManager.meshClient = &mockMeshClient{externalPeers: []*meshpb.Peer{
		{
			Ip:                peer,
			Hostname:          hostname,
			DoIAllowFileshare: true,
		},
	}}

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	event := fmt.Sprintf(`{
		"type": "RequestReceived",
		"data": {
			"peer": "%s",
			"transfer": "%s",
			"files": [
			  {
				"id": "testfile",
				"size": 1048576
			  }
			]
		}
	}`, peer, transferID)

	eventManager.EventFunc(event)

	assert.Equal(t, 1, len(notifier.notifications),
		"Transfer request notification was not sent after transfer request event was received.")

	transferRequestNotification := notifier.getLastNotification()
	assert.Equal(t, notifyNewTransferSummary, transferRequestNotification.summary)

	expectedNotificationBody := fmt.Sprintf(notifyNewTransferBody, transferID, hostname)
	assert.Equal(t, expectedNotificationBody, transferRequestNotification.body,
		"Invalid notification body.")

	expectedActions := []Action{
		{
			Action: transferAcceptAction,
			Key:    actionKeyAcceptTransfer,
		},
		{
			Action: transferCancelAction,
			Key:    actionKeyCancelTransfer,
		},
	}

	assert.Equal(t, expectedActions, transferRequestNotification.actions)
}

func TestTransferRequestNotificationAccept(t *testing.T) {
	peer := "172.20.0.5"

	pendingTransferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	pendingTransferNotificationID := uint32(0)

	transferFinishedID := "022cb1eb-ee22-431a-80c5-ba3050493c17"
	transferFinishedNotificationID := uint32(1)

	type testEnv struct {
		notificationManager *NotificationManager
		eventManager        *EventManager
		notifier            *mockNotifier
		fileshare           *mockEventManagerFileshare
	}

	setup := func(
		destinationDirectory string, freeSpace uint64) testEnv {
		currentUserUID := uint32(1000)
		currentUSerUIDString := strconv.Itoa(int(currentUserUID))
		currentUserGID := uint32(1000)
		currentUserGIDString := strconv.Itoa(int(currentUserGID))

		stat_t := &syscall.Stat_t{
			Uid: currentUserUID,
			Gid: currentUserGID,
		}
		directories := fstest.MapFS{
			"directory": &fstest.MapFile{Mode: os.ModeDir | 0777, Sys: stat_t},
			"symlink":   &fstest.MapFile{Mode: os.ModeSymlink | 0777, Sys: stat_t},
			"file":      &fstest.MapFile{Mode: 0777, Sys: stat_t},
		}

		filesystem := mockEventManagerFilesystem{
			MapFS:     directories,
			freeSpace: freeSpace,
		}

		notifier := mockNotifier{
			notifications: []mockNotification{},
			nextID:        uint32(pendingTransferNotificationID),
		}

		osInfo := mockEventManagerOsInfo{
			currentUser: user.User{Uid: currentUSerUIDString},
			groupIds: map[string][]string{
				currentUSerUIDString: {currentUserGIDString},
			},
		}

		notificationManager := NewMockNotificationManager(&osInfo)
		notificationManager.notifier = &notifier

		eventManager := NewEventManager(false, NoopStorage{},
			&mockMeshClient{},
			&osInfo,
			&filesystem,
			"")
		eventManager.notificationManager = &notificationManager
		eventManager.SetFileshare(&mockEventManagerFileshare{})
		eventManager.transfers[pendingTransferID] = &pb.Transfer{
			Status:    pb.Status_REQUESTED,
			Direction: pb.Direction_INCOMING,
			Files: []*pb.File{
				{
					Size: 1000,
				},
			}}

		eventManager.transfers = map[string]*pb.Transfer{
			pendingTransferID: {
				Status:    pb.Status_REQUESTED,
				Direction: pb.Direction_INCOMING,
				Files: []*pb.File{
					{
						Size: 1000,
					},
				},
			},
			transferFinishedID: {
				Status:    pb.Status_SUCCESS,
				Direction: pb.Direction_INCOMING,
				Files: []*pb.File{
					{
						Size: 1000,
					},
				},
			},
		}

		fileshare := &mockEventManagerFileshare{}

		notificationManager.eventManager = eventManager
		notificationManager.fileshare = fileshare
		notificationManager.defaultDownloadDir = destinationDirectory

		notificationManager.notifications.transfers = map[uint32]string{
			pendingTransferNotificationID:  pendingTransferID,
			transferFinishedNotificationID: transferFinishedID,
		}

		eventManager.meshClient = &mockMeshClient{externalPeers: []*meshpb.Peer{
			{
				Ip:                peer,
				DoIAllowFileshare: true,
			},
		}}

		return testEnv{
			notificationManager: &notificationManager,
			eventManager:        eventManager,
			notifier:            &notifier,
			fileshare:           fileshare,
		}
	}

	tests := []struct {
		name                      string
		destinationDirectoryName  string
		notificationID            uint32
		transferID                string
		freeSpace                 uint64
		expectedTransferStatus    pb.Status
		expectedErrorNotification string // empty for no error notifications
	}{
		{
			name:                      "transfer succesfully accepted",
			destinationDirectoryName:  "directory",
			notificationID:            pendingTransferNotificationID,
			transferID:                pendingTransferID,
			freeSpace:                 math.MaxUint64,
			expectedTransferStatus:    pb.Status_ONGOING,
			expectedErrorNotification: "",
		},
		{
			name:                      "destination directory is a symlink",
			destinationDirectoryName:  "symlink",
			notificationID:            pendingTransferNotificationID,
			transferID:                pendingTransferID,
			freeSpace:                 math.MaxUint64,
			expectedTransferStatus:    pb.Status_REQUESTED,
			expectedErrorNotification: downloadDirIsASymlinkError,
		},
		{
			name:                      "destination directory is a file",
			destinationDirectoryName:  "file",
			notificationID:            pendingTransferNotificationID,
			transferID:                pendingTransferID,
			freeSpace:                 math.MaxUint64,
			expectedTransferStatus:    pb.Status_REQUESTED,
			expectedErrorNotification: downloadDirIsNotADirError,
		},
		{
			name:                      "directory doesn't exist",
			destinationDirectoryName:  "no_dir",
			notificationID:            pendingTransferNotificationID,
			transferID:                pendingTransferID,
			freeSpace:                 math.MaxUint64,
			expectedTransferStatus:    pb.Status_REQUESTED,
			expectedErrorNotification: downloadDirNotFoundError,
		},
		{
			name:                      "not enough free space",
			destinationDirectoryName:  "directory",
			notificationID:            pendingTransferNotificationID,
			transferID:                pendingTransferID,
			freeSpace:                 1,
			expectedTransferStatus:    pb.Status_REQUESTED,
			expectedErrorNotification: notEnoughSpaceOnDeviceError,
		},
		{
			name:                      "transfer already finished",
			destinationDirectoryName:  "directory",
			notificationID:            transferFinishedNotificationID,
			transferID:                transferFinishedID,
			freeSpace:                 math.MaxUint64,
			expectedTransferStatus:    pb.Status_SUCCESS,
			expectedErrorNotification: transferInvalidated,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testEnv := setup(test.destinationDirectoryName, test.freeSpace)
			testEnv.notificationManager.AcceptTransfer(test.notificationID)

			assert.Equal(t, test.expectedTransferStatus, testEnv.eventManager.transfers[test.transferID].Status,
				"Invalid transfer status after accept notification action has been executed")

			if test.expectedErrorNotification == "" {
				assert.Empty(t, testEnv.notifier.notifications,
					"Unexpected notifications received: %v",
					testEnv.notifier.notifications)

				acceptedTransfer := testEnv.fileshare.getLastAcceptedTransferID()
				assert.Equal(t, test.transferID, acceptedTransfer, "Invalid transfer was accepted")
				return
			}

			assert.Equal(t, 1, len(testEnv.notifier.notifications), "Accept error notification was not received")

			errorNotification := testEnv.notifier.getLastNotification()
			assert.Equal(t, acceptFailedNotificationSummary, errorNotification.summary,
				"Error notification has invalid summary.")
			assert.Equal(t, test.expectedErrorNotification, errorNotification.body,
				"Error notification has invalid body.")
			assert.Equal(t, 0, len(errorNotification.actions),
				"Unexpected actions found in error notification: \n%v",
				errorNotification.actions)
		})
	}
}

func TestTransterRequestNotificationAcceptInvalidTransfer(t *testing.T) {
	peer := "172.20.0.5"

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	transferNotificationID := uint32(0)

	mockOsEnvironment := newMockSystemEnvironment(t)

	notifier := mockNotifier{
		notifications: []mockNotification{},
		nextID:        uint32(transferNotificationID),
	}

	notificationManager := NewMockNotificationManager(&mockOsEnvironment.mockEventManagerOsInfo)
	notificationManager.notifier = &notifier

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockOsEnvironment.mockEventManagerOsInfo,
		&mockOsEnvironment.mockEventManagerFilesystem,
		mockOsEnvironment.destinationDirectory)
	eventManager.notificationManager = &notificationManager

	notificationManager.eventManager = eventManager
	notificationManager.fileshare = &mockEventManagerFileshare{}
	notificationManager.defaultDownloadDir = mockOsEnvironment.destinationDirectory

	notificationManager.notifications.transfers = map[uint32]string{
		transferNotificationID: transferID,
	}

	eventManager.meshClient = &mockMeshClient{externalPeers: []*meshpb.Peer{
		{
			Ip:                peer,
			DoIAllowFileshare: true,
		},
	}}

	notificationManager.AcceptTransfer(transferNotificationID)

	assert.Equal(t, 1, len(notifier.notifications), "Accept error notification was not received")

	errorNotification := notifier.getLastNotification()
	assert.Equal(t, acceptFailedNotificationSummary, errorNotification.summary,
		"Error notification has invalid summary.")
	assert.Equal(t, genericError, errorNotification.body,
		"Error notification has invalid body.")
	assert.Equal(t, 0, len(errorNotification.actions),
		"Unexpected actions found in error notification: \n%v",
		errorNotification.actions)
}

func TestTransferRequestNotificationCancel(t *testing.T) {
	peer := "172.20.0.5"

	pendingTransferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	pendingTransferNotificationID := uint32(0)

	transferAlreadyCanceledID := "5f4c3ec4-d4fe-4335-beb6-5db2ffbae351"
	transferAlreadyCanceledNotificationID := uint32(1)

	transferFinishedID := "022cb1eb-ee22-431a-80c5-ba3050493c17"
	transferFinishedNotificationID := uint32(2)

	invalidTransferID := "022cb1eb-invalid-ba3050493c17"
	invalidTransferNotificationID := uint32(3)

	setup := func() (*NotificationManager, *mockEventManagerFileshare, *mockNotifier) {
		notifier := mockNotifier{
			notifications: []mockNotification{},
			nextID:        uint32(pendingTransferNotificationID),
		}

		notificationManager := NewMockNotificationManager(&mockEventManagerOsInfo{})
		notificationManager.notifications.transfers[pendingTransferNotificationID] = pendingTransferID
		notificationManager.notifications.transfers[transferAlreadyCanceledNotificationID] = transferAlreadyCanceledID
		notificationManager.notifications.transfers[transferFinishedNotificationID] = transferFinishedID
		notificationManager.notifications.transfers[invalidTransferNotificationID] = invalidTransferID
		notificationManager.notifier = &notifier

		eventManager := NewEventManager(false, NoopStorage{},
			&mockMeshClient{},
			&mockEventManagerOsInfo{},
			&mockEventManagerFilesystem{},
			"")
		eventManager.notificationManager = &notificationManager
		eventManager.SetFileshare(&mockEventManagerFileshare{})

		notificationManager.eventManager = eventManager
		fileshare := mockEventManagerFileshare{}
		notificationManager.fileshare = &fileshare

		eventManager.meshClient = &mockMeshClient{externalPeers: []*meshpb.Peer{
			{
				Ip:                peer,
				DoIAllowFileshare: true,
			},
		}}

		eventManager.transfers[pendingTransferID] = &pb.Transfer{Status: pb.Status_REQUESTED}
		eventManager.transfers[transferAlreadyCanceledID] = &pb.Transfer{Status: pb.Status_CANCELED}
		eventManager.transfers[transferFinishedID] = &pb.Transfer{Status: pb.Status_SUCCESS}

		return &notificationManager, &fileshare, &notifier
	}

	tests := []struct {
		name                      string
		notificationID            uint32
		transferID                string
		expectedErrorNotification string // empty for no error notifications
	}{
		{
			name:                      "transfer succesfully canceled",
			notificationID:            pendingTransferNotificationID,
			transferID:                pendingTransferID,
			expectedErrorNotification: "",
		},
		{
			name:                      "transfer already canceled",
			notificationID:            transferAlreadyCanceledNotificationID,
			transferID:                transferAlreadyCanceledID,
			expectedErrorNotification: transferInvalidated,
		},
		{
			name:                      "transfer finished",
			notificationID:            transferFinishedNotificationID,
			transferID:                transferAlreadyCanceledID,
			expectedErrorNotification: transferInvalidated,
		},
		{
			name:                      "transfer does not exist",
			notificationID:            invalidTransferNotificationID,
			transferID:                invalidTransferID,
			expectedErrorNotification: genericError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			notificationManager, fileshare, notifier := setup()
			notificationManager.CancelTransfer(test.notificationID)

			if test.expectedErrorNotification == "" {
				// Assert that only the transfer request notification was received
				assert.NotEmpty(t, fileshare.canceledTransferIDs, "No transfers were canceled")
				assert.Equal(t, test.transferID, fileshare.getLastCanceledTransferID(),
					"Invalid transfer was canceled")
				assert.Empty(t, notifier.notifications,
					"Unexpected notification received: %v",
					notifier.notifications)
				return
			}

			assert.Equal(t, 1, len(notifier.notifications), "Cancel error notification was not received")

			errorNotification := notifier.getLastNotification()
			assert.Equal(t, cancelFailedNotificationSummary, errorNotification.summary,
				"Error notification has invalid summary.")
			assert.Equal(t, test.expectedErrorNotification, errorNotification.body,
				"Error notification has invalid body.")
			assert.Equal(t, 0, len(errorNotification.actions),
				"Unexpected actions found in error notification: \n%v",
				errorNotification.actions)
		})
	}
}

func TestAutoaccept(t *testing.T) {
	mockOsEnvironment := newMockSystemEnvironment(t)

	symlinkDirectoryName := "symlink"
	mockOsEnvironment.MapFS[symlinkDirectoryName] =
		&fstest.MapFile{Mode: os.ModeSymlink | 0777,
			Sys: &syscall.Stat_t{
				Uid: mockOsEnvironment.currentUserUID,
				Gid: mockOsEnvironment.currentUserGID,
			}}

	fileDirectoryName := "not_dir"
	mockOsEnvironment.MapFS[fileDirectoryName] =
		&fstest.MapFile{Mode: 0777,
			Sys: &syscall.Stat_t{
				Uid: mockOsEnvironment.currentUserUID,
				Gid: mockOsEnvironment.currentUserGID,
			}}

	notifier := mockNotifier{
		notifications: []mockNotification{},
		nextID:        0,
	}

	notificationManager := NewMockNotificationManager(&mockOsEnvironment.mockEventManagerOsInfo)
	notificationManager.notifier = &notifier

	eventManager := NewEventManager(false, NoopStorage{},
		&mockMeshClient{},
		&mockOsEnvironment.mockEventManagerOsInfo,
		&mockOsEnvironment.mockEventManagerFilesystem,
		mockOsEnvironment.destinationDirectory)
	eventManager.notificationManager = &notificationManager

	notificationManager.eventManager = eventManager
	notificationManager.defaultDownloadDir = mockOsEnvironment.destinationDirectory

	notificationManager.notifications.transfers = map[uint32]string{}

	peerAutacceptIP := "172.20.0.5"
	peerAutoacceptHostname := "internal.peer1.nord"

	eventManager.meshClient = &mockMeshClient{externalPeers: []*meshpb.Peer{
		{
			Ip:                peerAutacceptIP,
			Hostname:          peerAutoacceptHostname,
			DoIAllowFileshare: true,
			AlwaysAcceptFiles: true,
		},
	}}

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	event := fmt.Sprintf(`{
		"type": "RequestReceived",
		"data": {
			"peer": "%s",
			"transfer": "%s",
			"files": [
			  {
				"id": "testfile",
				"size": 1048576
			  }
			]
		}
	}`, peerAutacceptIP, transferID)

	tests := []struct {
		name                        string
		defaultDownloadDirectory    string
		acceptedTransferID          string
		expectedNotificationSummary string
		expectedNotificationBody    string
	}{
		{
			name:                        "autoaccept ok",
			defaultDownloadDirectory:    mockOsEnvironment.destinationDirectory,
			acceptedTransferID:          transferID,
			expectedNotificationSummary: notifyNewAutoacceptTransfer,
			expectedNotificationBody:    fmt.Sprintf(notifyNewTransferBody, transferID, peerAutoacceptHostname),
		},
		{
			name:                        "autoaccept dir is a symlink",
			defaultDownloadDirectory:    symlinkDirectoryName,
			acceptedTransferID:          "",
			expectedNotificationSummary: notifyAutoacceptFailed,
			expectedNotificationBody: fmt.Sprintf(
				"%s\n%s",
				downloadDirIsASymlinkError,
				fmt.Sprintf(notifyNewTransferBody, transferID, peerAutoacceptHostname)),
		},
		{
			name:                        "autoaccept dir is not a directory",
			defaultDownloadDirectory:    fileDirectoryName,
			acceptedTransferID:          "",
			expectedNotificationSummary: notifyAutoacceptFailed,
			expectedNotificationBody: fmt.Sprintf(
				"%s\n%s",
				downloadDirIsNotADirError,
				fmt.Sprintf(notifyNewTransferBody, transferID, peerAutoacceptHostname)),
		},
		{
			name:                        "autoaccept dir not found",
			defaultDownloadDirectory:    "not-found",
			acceptedTransferID:          "",
			expectedNotificationSummary: notifyAutoacceptFailed,
			expectedNotificationBody: fmt.Sprintf(
				"%s\n%s",
				downloadDirNotFoundError,
				fmt.Sprintf(notifyNewTransferBody, transferID, peerAutoacceptHostname)),
		},
	}

	for _, test := range tests {
		mockFileshare := mockEventManagerFileshare{}

		eventManager.fileshare = &mockFileshare
		eventManager.defaultDownloadDir = test.defaultDownloadDirectory
		eventManager.EventFunc(event)

		if test.acceptedTransferID != "" {
			assert.NotEmpty(t, mockFileshare.acceptedTransferIDS,
				"Incoming transfer was not accepted")
			assert.Equal(t, transferID, mockFileshare.getLastAcceptedTransferID(),
				"Invalid transfer was accepted")
		} else {
			assert.Empty(t, mockFileshare.acceptedTransferIDS,
				"Unexpected incoming transfer was accepted")
		}

		notification := notifier.getLastNotification()
		assert.Equal(t, test.expectedNotificationSummary, notification.summary,
			"Invalid notification summary")

		assert.Equal(t, test.expectedNotificationBody, notification.body, "Invalid notification body")

		assert.Empty(t, notification.actions,
			"Unexpected actions found in autoaccepted transfer notification")
	}
}
