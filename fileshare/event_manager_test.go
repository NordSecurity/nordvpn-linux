package fileshare

import (
	"errors"
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

func (*mockEventManagerFileshare) GetTransfersSince(t time.Time) ([]LibdropTransfer, error) {
	return nil, nil
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

type mockStorage struct {
	transfers map[string]*pb.Transfer
	err       error
}

func (m *mockStorage) Load() (map[string]*pb.Transfer, error) {
	return m.transfers, m.err
}

func TestGetTransfers(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(false, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
	eventManager.SetStorage(storage)

	timeNow := time.Now()
	for i := 10; i > 0; i-- {
		storage.transfers[strconv.Itoa(i)] = &pb.Transfer{
			Id:      strconv.Itoa(i),
			Created: timestamppb.New(timeNow.Add(-time.Second * time.Duration(i))),
		}
	}
	storage.transfers["2"].Files = []*pb.File{{Id: "file"}}
	eventManager.liveTransfers["2"] = &LiveTransfer{
		TotalTransferred: 2,
		Files: map[string]*LiveFile{
			"file": {
				Transferred: 3,
			},
		},
	}

	transfers, err := eventManager.GetTransfers()
	assert.NoError(t, err)
	assert.Equal(t, 10, len(transfers))
	// Check if ordered
	for i := 0; i < 9; i++ {
		assert.True(t, transfers[i].Created.AsTime().Before(transfers[i+1].Created.AsTime()))
	}

	assert.Equal(t, "2", transfers[8].Id)
	assert.EqualValues(t, 2, transfers[8].TotalTransferred)
	assert.EqualValues(t, 3, transfers[8].Files[0].Transferred)

	// Almost same functionality, so let's just test GetTransfer here as well
	transfer, err := eventManager.GetTransfer("2")
	assert.NoError(t, err)
	assert.Equal(t, transfers[8], transfer)
}

func TestGetTransfers_Fail(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(false, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.SetStorage(&mockStorage{err: errors.New("storage failure")})
	_, err := eventManager.GetTransfers()
	assert.ErrorContains(t, err, "storage failure")
}

func TestTransferProgress(t *testing.T) {
	category.Set(t, category.Unit)

	eventManager := NewEventManager(false, &mockMeshClient{}, &mockEventManagerOsInfo{}, &mockEventManagerFilesystem{}, "")
	eventManager.SetFileshare(&mockEventManagerFileshare{})
	storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
	eventManager.SetStorage(storage)

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	peer := "12.12.12.12"
	path := "/tmp"
	file1 := "testfile-small"
	file1ID := "file1ID"
	file1sz := 100
	file2 := "testfile-big"
	file2ID := "file2ID"
	file2sz := 1000
	file3 := "file3.txt"
	file3ID := "file3ID"
	file3sz := 1000

	storage.transfers[transferID] = &pb.Transfer{
		Id:        transferID,
		Peer:      peer,
		Path:      path,
		Status:    pb.Status_REQUESTED,
		TotalSize: uint64(file1sz) + uint64(file2sz) + uint64(file3sz),
		Files: []*pb.File{
			{
				Id:     file1ID,
				Path:   file1,
				Size:   uint64(file1sz),
				Status: pb.Status_REQUESTED,
			},
			{
				Id:     file2ID,
				Path:   file2,
				Size:   uint64(file2sz),
				Status: pb.Status_REQUESTED,
			},
			{
				Id:     file3ID,
				Path:   file3,
				Size:   uint64(file3sz),
				Status: pb.Status_REQUESTED,
			},
		},
	}

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

	progCh := eventManager.Subscribe(transferID)

	eventManager.EventFunc(
		fmt.Sprintf(`{
		"type": "TransferStarted",
		"data": {
			"transfer": "%s",
			"file": "%s"
		}
		}`, transferID, file1ID))

	eventManager.EventFunc(
		fmt.Sprintf(`{
		"type": "TransferStarted",
		"data": {
			"transfer": "%s",
			"file": "%s"
		}
		}`, transferID, file2ID))

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
	expectedProgress := uint32(float64(transferredBytes) / float64(file1sz+file2sz+file3sz) * 100)
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
		// Final transfer state is determined from storage
		storage.transfers[transferID].Status = pb.Status_SUCCESS
		eventManager.EventFunc(
			fmt.Sprintf(`{
				"type": "TransferFinished",
				"data": {
					"transfer": "%s",
					"reason": "TransferCanceled"
				}
				}`, transferID))
		waitGroup.Done()
	}()

	progressEvent = <-progCh
	assert.Equal(t, pb.Status_SUCCESS, progressEvent.Status)

	waitGroup.Wait()
	_, ok := eventManager.transferSubscriptions[transferID]
	assert.False(t, ok) // expect subscriber to be removed
	_, ok = eventManager.liveTransfers[transferID]
	assert.False(t, ok) // expect transfer not to be tracked anymore
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
			testName:    "accept transfer success",
			transfer:    "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr: nil,
			files:       []string{},
			sizeLimit:   6,
		},
		{
			testName:    "accept files success",
			transfer:    "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr: nil,
			files:       []string{"test/file_A"},
			sizeLimit:   1,
		},
		{
			testName:    "transfer doesn't exist",
			transfer:    "invalid_transfer",
			expectedErr: ErrTransferNotFound,
			files:       []string{},
			sizeLimit:   6,
		},
		{
			testName:    "file doesn't exist",
			transfer:    "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr: ErrFileNotFound,
			files:       []string{"invalid_file"},
			sizeLimit:   6,
		},
		{
			testName:    "size exceeds limit",
			transfer:    "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr: ErrSizeLimitExceeded,
			files:       []string{},
			sizeLimit:   5,
		},
		{
			testName:    "partial transfer size exceeds limit",
			transfer:    "c13c619c-c70b-49b8-9396-72de88155c43",
			expectedErr: ErrSizeLimitExceeded,
			files:       []string{"test/file_C"},
			sizeLimit:   2,
		},
	}

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

	mockSystemEnvironment := newMockSystemEnvironment(t)

	for _, test := range tests {
		mockSystemEnvironment.mockEventManagerFilesystem.freeSpace = test.sizeLimit

		eventManager := NewEventManager(false,
			&mockMeshClient{},
			&mockSystemEnvironment.mockEventManagerOsInfo,
			&mockSystemEnvironment.mockEventManagerFilesystem,
			"")
		storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
		eventManager.SetStorage(storage)
		storage.transfers[transferID] = &pb.Transfer{
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
			_, err := eventManager.AcceptTransfer(test.transfer, mockSystemEnvironment.destinationDirectory, test.files)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestAcceptTransfer_Outgoing(t *testing.T) {
	category.Set(t, category.Unit)

	mockSystemEnvironment := newMockSystemEnvironment(t)

	eventManager := NewEventManager(false,
		&mockMeshClient{},
		&mockSystemEnvironment.mockEventManagerOsInfo,
		&mockSystemEnvironment.mockEventManagerFilesystem,
		"")
	storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
	eventManager.SetStorage(storage)
	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	storage.transfers[transferID] = &pb.Transfer{
		Id:        transferID,
		Direction: pb.Direction_OUTGOING,
		Status:    pb.Status_REQUESTED,
	}

	_, err := eventManager.AcceptTransfer(transferID, mockSystemEnvironment.destinationDirectory, []string{})
	assert.Equal(t, ErrTransferAcceptOutgoing, err)
}

func TestAcceptTransfer_AlreadyAccepted(t *testing.T) {
	category.Set(t, category.Unit)

	mockSystemEnvironment := newMockSystemEnvironment(t)

	eventManager := NewEventManager(false,
		&mockMeshClient{},
		&mockSystemEnvironment.mockEventManagerOsInfo,
		&mockSystemEnvironment.mockEventManagerFilesystem,
		"")
	storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
	eventManager.SetStorage(storage)
	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"
	storage.transfers[transferID] = &pb.Transfer{
		Id:        transferID,
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_ONGOING,
	}

	_, err := eventManager.AcceptTransfer("c13c619c-c70b-49b8-9396-72de88155c43", mockSystemEnvironment.destinationDirectory, []string{})
	assert.Equal(t, ErrTransferAlreadyAccepted, err)
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

		eventManager := NewEventManager(false,
			&mockMeshClient{},
			&mockEventManagerOsInfo{},
			&mockEventManagerFilesystem{},
			"")
		eventManager.notificationManager = &notificationManager
		eventManager.SetFileshare(&mockEventManagerFileshare{})
		storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
		eventManager.SetStorage(storage)
		storage.transfers[transferID] = &pb.Transfer{
			Id:     transferID,
			Status: pb.Status_ONGOING,
			Files: []*pb.File{
				{Id: fileID, FullPath: filePath, Status: pb.Status_ONGOING},
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

	eventManager := NewEventManager(false,
		&mockMeshClient{},
		&mockEventManagerOsInfo{},
		&mockEventManagerFilesystem{},
		"")
	eventManager.notificationManager = &notificationManager
	eventManager.SetFileshare(&mockEventManagerFileshare{})
	storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
	eventManager.SetStorage(storage)
	storage.transfers[transferID] = &pb.Transfer{
		Id:     transferID,
		Status: pb.Status_ONGOING,
		Files: []*pb.File{
			{Id: fileID, FullPath: filePath, Status: pb.Status_ONGOING},
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
	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

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

	eventManager := NewEventManager(false,
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

		eventManager := NewEventManager(false,
			&mockMeshClient{},
			&osInfo,
			&filesystem,
			"")
		storage := &mockStorage{}
		eventManager.SetStorage(storage)
		storage.transfers = map[string]*pb.Transfer{
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

	eventManager := NewEventManager(false,
		&mockMeshClient{},
		&mockOsEnvironment.mockEventManagerOsInfo,
		&mockOsEnvironment.mockEventManagerFilesystem,
		mockOsEnvironment.destinationDirectory)
	eventManager.notificationManager = &notificationManager
	eventManager.SetStorage(&mockStorage{})

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

		eventManager := NewEventManager(false,
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

		storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
		eventManager.SetStorage(storage)
		storage.transfers[pendingTransferID] = &pb.Transfer{Status: pb.Status_REQUESTED}
		storage.transfers[transferAlreadyCanceledID] = &pb.Transfer{Status: pb.Status_CANCELED}
		storage.transfers[transferFinishedID] = &pb.Transfer{Status: pb.Status_SUCCESS}

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

	eventManager := NewEventManager(false,
		&mockMeshClient{},
		&mockOsEnvironment.mockEventManagerOsInfo,
		&mockOsEnvironment.mockEventManagerFilesystem,
		mockOsEnvironment.destinationDirectory)
	eventManager.notificationManager = &notificationManager
	storage := &mockStorage{transfers: map[string]*pb.Transfer{}}
	eventManager.SetStorage(storage)

	notificationManager.eventManager = eventManager
	notificationManager.defaultDownloadDir = mockOsEnvironment.destinationDirectory

	notificationManager.notifications.transfers = map[uint32]string{}

	peerAutoAcceptIP := "172.20.0.5"
	peerAutoacceptHostname := "internal.peer1.nord"

	eventManager.meshClient = &mockMeshClient{externalPeers: []*meshpb.Peer{
		{
			Ip:                peerAutoAcceptIP,
			Hostname:          peerAutoacceptHostname,
			DoIAllowFileshare: true,
			AlwaysAcceptFiles: true,
		},
	}}

	transferID := "c13c619c-c70b-49b8-9396-72de88155c43"

	storage.transfers[transferID] = &pb.Transfer{
		Id:        transferID,
		Peer:      peerAutoAcceptIP,
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
		Files: []*pb.File{
			{
				Id:   "testfile",
				Size: 1048576,
			},
		},
	}

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
	}`, peerAutoAcceptIP, transferID)

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
		t.Run(test.name, func(t *testing.T) {
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
		})
	}
}
