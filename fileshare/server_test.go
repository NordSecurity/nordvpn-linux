package fileshare

import (
	"context"
	"errors"
	"io/fs"
	"math"
	"net/netip"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"testing"
	"testing/fstest"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"golang.org/x/exp/slices"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
)

const exampleFileID2 = "file2"

type mockServerFileshare struct {
	Fileshare
	cancelReturnValue      error
	acceptFirstReturnValue error // Only first return has this value, subsequent always have nil
	destinationPeer        string
	acceptedFiles          []string
	canceledFiles          []string
}

func isFileListEqual(t *testing.T, lhs []string, rhs []string) bool {
	t.Helper()

	if len(lhs) != len(rhs) {
		return false
	}

	for _, file := range lhs {
		if !slices.Contains(rhs, file) {
			return false
		}
	}

	return true
}

func (m *mockServerFileshare) CancelFile(string, fileID string) error {
	m.canceledFiles = append(m.canceledFiles, fileID)
	return m.cancelReturnValue
}

func (mockServerFileshare) Enable(listenAddress netip.Addr) error { return nil }

func (mockServerFileshare) Disable() error { return nil }

func (m *mockServerFileshare) Send(peer netip.Addr, paths []string) (string, error) {
	m.destinationPeer = peer.String()
	return "", nil
}

func (m *mockServerFileshare) Accept(transferID, dstPath string, fileID string) error {
	m.acceptedFiles = append(m.acceptedFiles, fileID)

	err := m.acceptFirstReturnValue
	m.acceptFirstReturnValue = nil
	return err
}

type mockAcceptServer struct {
	pb.Fileshare_AcceptServer
	serverError error
	response    *pb.StatusResponse
}

func (m *mockAcceptServer) Send(resp *pb.StatusResponse) error {
	m.response = resp
	return m.serverError
}

type mockSendServer struct {
	pb.Fileshare_SendServer
	response        *pb.StatusResponse
	sendReturnValue error
}

func (m *mockSendServer) Send(resp *pb.StatusResponse) error {
	m.response = resp
	return m.sendReturnValue
}

type mockListServer struct {
	pb.Fileshare_ListServer
	responses []*pb.ListResponse
}

func (m *mockListServer) Send(resp *pb.ListResponse) error {
	m.responses = append(m.responses, resp)
	return nil
}

type mockFilesystem struct {
	fstest.MapFS
	freeSpace uint64
}

func newMockFilesystem() mockFilesystem {
	mockFilesystem := mockFilesystem{}
	mockFilesystem.MapFS = make(fstest.MapFS)
	return mockFilesystem
}

func (mf mockFilesystem) Lstat(path string) (fs.FileInfo, error) {
	fileInfo, err := mf.MapFS.Stat(path)
	return fileInfo, err
}

func (mf mockFilesystem) Statfs(path string) (unix.Statfs_t, error) {
	if mf.freeSpace == 0 {
		return unix.Statfs_t{Bavail: math.MaxUint64, Bsize: 1}, nil
	}
	return unix.Statfs_t{Bavail: mf.freeSpace, Bsize: 1}, nil
}

func populateMapFs(t *testing.T, mapfs *fstest.MapFS, directoryName string, fileCount int) {
	t.Helper()

	(*mapfs)[directoryName] = &fstest.MapFile{Mode: fs.ModeDir}
	for filename := 0; filename < fileCount; filename++ {
		(*mapfs)[directoryName+"/"+strconv.Itoa(filename)] = &fstest.MapFile{}
	}
}

type mockOsInfo struct {
	currentUser user.User
	groupIds    map[string][]string
}

func (mOS *mockOsInfo) CurrentUser() (*user.User, error) {
	return &mOS.currentUser, nil
}

func (mOS *mockOsInfo) GetGroupIds(userInfo *user.User) ([]string, error) {
	return mOS.groupIds[userInfo.Uid], nil
}

type mockMeshClient struct {
	meshpb.MeshnetClient
	isEnabled      bool
	localPeers     []*meshpb.Peer
	externalPeers  []*meshpb.Peer
	selfPeer       *meshpb.Peer
	getPeersCalled bool
}

// IsEnabled mock implementation
func (m *mockMeshClient) IsEnabled(ctx context.Context, in *meshpb.Empty, opts ...grpc.CallOption) (*meshpb.ServiceBoolResponse, error) {
	return &meshpb.ServiceBoolResponse{Response: &meshpb.ServiceBoolResponse_Value{Value: m.isEnabled}}, nil
}

// GetPeers mock implementation
func (m *mockMeshClient) GetPeers(ctx context.Context, in *meshpb.Empty, opts ...grpc.CallOption) (*meshpb.GetPeersResponse, error) {
	response := &meshpb.GetPeersResponse{
		Response: &meshpb.GetPeersResponse_Peers{
			Peers: &meshpb.PeerList{
				Local:    m.localPeers,
				External: m.externalPeers,
				Self:     m.selfPeer,
			},
		},
	}
	m.getPeersCalled = true
	return response, nil
}

// NotifyNewTransfer mock implementation
func (m *mockMeshClient) NotifyNewTransfer(ctx context.Context, in *meshpb.NewTransferNotification, opts ...grpc.CallOption) (*meshpb.NotifyNewTransferResponse, error) {
	return &meshpb.NotifyNewTransferResponse{
		Response: &meshpb.NotifyNewTransferResponse_Empty{},
	}, nil
}

func getTransfers(t *testing.T, numberOfTransfers int) map[string]*pb.Transfer {
	t.Helper()

	transfersMap := make(map[string]*pb.Transfer, numberOfTransfers)
	for transfer := 0; transfer < numberOfTransfers; transfer++ {
		transferID := strconv.Itoa(transfer)
		pbTransfer := &pb.Transfer{
			Id: transferID,
		}

		transfersMap[transferID] = pbTransfer
	}

	return transfersMap
}

func TestSend(t *testing.T) {
	category.Set(t, category.Unit)

	mockFs := newMockFilesystem()
	directory := "dir"
	populateMapFs(t, &mockFs.MapFS, directory, 3)

	internalPeer2IP := "219.150.143.226"
	internalPeer2Pubkey := "FofTQLNKWoHwep2syHdzEg3RGVErLDizgeMArzwMdWT="
	internalPeer2Hostname := "internal.peer2.nord"
	localPeers := []*meshpb.Peer{
		{
			Ip:                 "38.30.202.86",
			Pubkey:             "aZ9KwmEzystVJ0R1YitV02NzNngmSrZ3JDTj6tkI8T6=",
			Hostname:           "internal.peer1.nord",
			IsFileshareAllowed: true,
			Status:             1,
		},
		{
			Ip:                 internalPeer2IP,
			Pubkey:             internalPeer2Pubkey,
			Hostname:           internalPeer2Hostname,
			IsFileshareAllowed: true,
			Status:             1,
		},
	}

	externalPeer2IP := "124.252.136.82"
	externalPeer2Pubkey := "yaisO7jHDcEeb6NTasfhr3duUGIJKQipv4bC9SSDvQP="
	externalPeer2Hostname := "external.peer1.nord"

	peerSendingfilesNotAllowedIP := "116.51.81.30"
	peerSendingfilesNotAllowedPubkey := "TndF1zMx38gd3PF5ho1eSc2FqtkojwlYdOxcmLZn8OU"
	peerSendingfilesNotAllowedHostname := "external.peer3.nord"

	externalPeers := []*meshpb.Peer{
		{
			Ip:                 "124.252.136.82",
			Pubkey:             "yaisO7jHDcEeb6NTasfhr3duUGIJKQipv4bC9SSDvQP=",
			Hostname:           "external.peer1.nord",
			IsFileshareAllowed: true,
			Status:             1,
		},
		{
			Ip:                 externalPeer2IP,
			Pubkey:             externalPeer2Pubkey,
			Hostname:           externalPeer2Hostname,
			IsFileshareAllowed: true,
			Status:             1,
		},
		{
			Ip:                 peerSendingfilesNotAllowedIP,
			Pubkey:             peerSendingfilesNotAllowedPubkey,
			Hostname:           peerSendingfilesNotAllowedHostname,
			IsFileshareAllowed: false,
			Status:             1,
		},
	}

	fileshareTests := []struct {
		testName                string
		path                    string
		peer                    string
		transferSilent          bool
		expectedError           *pb.Error
		expectedDestinationPeer string
	}{
		{
			testName:                "send to internal peer by hostname",
			path:                    directory,
			peer:                    internalPeer2Hostname,
			transferSilent:          true,
			expectedError:           nil,
			expectedDestinationPeer: internalPeer2IP,
		},
		{
			testName:                "send to external peer by hostname",
			path:                    directory,
			peer:                    externalPeer2Hostname,
			transferSilent:          true,
			expectedError:           nil,
			expectedDestinationPeer: externalPeer2IP,
		},
		{
			testName:                "send to internal peer by ip",
			path:                    directory,
			peer:                    internalPeer2IP,
			transferSilent:          true,
			expectedError:           nil,
			expectedDestinationPeer: internalPeer2IP,
		},
		{
			testName:                "send to external peer by ip",
			path:                    directory,
			peer:                    externalPeer2IP,
			transferSilent:          true,
			expectedError:           nil,
			expectedDestinationPeer: externalPeer2IP,
		},
		{
			testName:                "send to internal peer by pubkey",
			path:                    directory,
			peer:                    internalPeer2Pubkey,
			transferSilent:          true,
			expectedError:           nil,
			expectedDestinationPeer: internalPeer2IP,
		},
		{
			testName:                "send to external peer by pubkey",
			path:                    directory,
			peer:                    externalPeer2Pubkey,
			transferSilent:          true,
			expectedError:           nil,
			expectedDestinationPeer: externalPeer2IP,
		},
		{
			testName:                "invalid peer",
			path:                    directory,
			peer:                    "no peer",
			transferSilent:          true,
			expectedError:           fileshareError(pb.FileshareErrorCode_INVALID_PEER),
			expectedDestinationPeer: "",
		},
		{
			testName:                "sending files not allowed ip",
			path:                    directory,
			peer:                    peerSendingfilesNotAllowedIP,
			transferSilent:          true,
			expectedError:           fileshareError(pb.FileshareErrorCode_SENDING_NOT_ALLOWED),
			expectedDestinationPeer: "",
		},
		{
			testName:                "sending files not allowed pubkey",
			path:                    directory,
			peer:                    peerSendingfilesNotAllowedPubkey,
			transferSilent:          true,
			expectedError:           fileshareError(pb.FileshareErrorCode_SENDING_NOT_ALLOWED),
			expectedDestinationPeer: "",
		},
		{
			testName:                "sending files not allowed hostname",
			path:                    directory,
			peer:                    peerSendingfilesNotAllowedHostname,
			transferSilent:          true,
			expectedError:           fileshareError(pb.FileshareErrorCode_SENDING_NOT_ALLOWED),
			expectedDestinationPeer: "",
		},
	}

	for _, test := range fileshareTests {
		mockMeshClient := mockMeshClient{
			isEnabled:     true,
			localPeers:    localPeers,
			externalPeers: externalPeers,
		}

		mockFileshare := mockServerFileshare{}
		server := NewServer(
			&mockFileshare,
			&EventManager{},
			&mockMeshClient,
			mockFs,
			&mockOsInfo{},
			0,
			nil,
		)

		sendServer := mockSendServer{}

		t.Run(test.testName, func(t *testing.T) {
			err := server.Send(
				&pb.SendRequest{
					Peer:   test.peer,
					Paths:  []string{test.path},
					Silent: test.transferSilent},
				&sendServer,
			)
			assert.Equal(t, nil, err)
			assert.Equal(t, test.expectedError, sendServer.response.GetError())
			assert.Equal(t, test.expectedDestinationPeer, mockFileshare.destinationPeer)
		})
	}
}

func TestSendDirectoryFilesystemErrorHandling(t *testing.T) {
	category.Set(t, category.Unit)

	mockFs := newMockFilesystem()
	directoryTooManyFiles := "directory_too_many_files"
	populateMapFs(t, &mockFs.MapFS, directoryTooManyFiles, 1001)

	directoryTooDeepName := "directory_too_deep"
	currentDir := directoryTooDeepName
	for directory := 0; directory < 8; directory++ {
		mockFs.MapFS[currentDir] = &fstest.MapFile{Mode: fs.ModeDir}
		currentDir = currentDir + "/" + strconv.Itoa(directory)
	}

	tooManyFilesCumulative1 := "directory_too_many_files_cumulative_1"
	populateMapFs(t, &mockFs.MapFS, tooManyFilesCumulative1, 600)

	tooManyFilesCumulative2 := "directory_too_many_files_cumulative_2"
	populateMapFs(t, &mockFs.MapFS, tooManyFilesCumulative2, 600)

	exectFileLimit := "directory_exact_limit"
	populateMapFs(t, &mockFs.MapFS, exectFileLimit, 1000)

	file1 := exampleFileID1
	mockFs.MapFS[file1] = &fstest.MapFile{}
	file2 := exampleFileID2
	mockFs.MapFS[file2] = &fstest.MapFile{}
	file3 := "file3"
	mockFs.MapFS[file3] = &fstest.MapFile{}

	emptyDirectory := "empty"
	mockFs.MapFS[emptyDirectory] = &fstest.MapFile{Mode: fs.ModeDir}

	fileshareTests := []struct {
		testName             string
		paths                []string
		transferSilent       bool
		expectedSendResponse *pb.StatusResponse
	}{
		{
			testName:             "too many files",
			paths:                []string{directoryTooManyFiles},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_TOO_MANY_FILES)},
		},
		{
			testName:             "file doesent exist",
			paths:                []string{"nofile"},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND)},
		},
		{
			testName:             "directory too deep",
			paths:                []string{directoryTooDeepName},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_DIRECTORY_TOO_DEEP)},
		},
		{
			testName:             "too many files in multidirectory transfer",
			paths:                []string{tooManyFilesCumulative1, tooManyFilesCumulative2},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_TOO_MANY_FILES)},
		},
		{
			testName:             "too many files in multifile transfer",
			paths:                []string{exectFileLimit, file1},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_TOO_MANY_FILES)},
		},
		{
			testName:             "file in multifile transfer doesnt exist",
			paths:                []string{file1, file2, file3, "nofile"},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND)},
		},
		{
			testName:             "no files",
			paths:                []string{emptyDirectory},
			transferSilent:       true,
			expectedSendResponse: &pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_NO_FILES)},
		},
	}

	for _, test := range fileshareTests {
		server := NewServer(
			&mockServerFileshare{},
			&EventManager{},
			&mockMeshClient{isEnabled: true},
			mockFs,
			&mockOsInfo{},
			0,
			nil,
		)

		sendServer := mockSendServer{}

		t.Run(test.testName, func(t *testing.T) {
			err := server.Send(
				&pb.SendRequest{Peer: "100.96.115.182", Paths: test.paths, Silent: test.transferSilent},
				&sendServer)
			assert.Equal(t, nil, err)
			assert.Equal(t, test.expectedSendResponse, sendServer.response)
		})
	}
}

func TestAccept(t *testing.T) {
	category.Set(t, category.Unit)

	filePath := "test_a.txt"
	transferID := exampleUUID

	mockFs := newMockFilesystem()

	currentUserUID := uint32(1000)
	currentUserGID := uint32(1000)

	acceptDirName := "tmp"
	statCurrentUserOwner := &syscall.Stat_t{
		Uid: currentUserUID,
		Gid: currentUserGID,
	}
	mockFs.MapFS[acceptDirName] = &fstest.MapFile{Mode: os.ModeDir | 0777, Sys: statCurrentUserOwner}

	acceptSymlinkName := "link"
	mockFs.MapFS[acceptSymlinkName] = &fstest.MapFile{Mode: os.ModeSymlink | 0777, Sys: statCurrentUserOwner}

	acceptNotDirName := "not_dir"
	mockFs.MapFS[acceptNotDirName] = &fstest.MapFile{Mode: 0777, Sys: statCurrentUserOwner}

	currentUserUIDStr := strconv.Itoa(int(currentUserUID))
	currentUserGIDStr := strconv.Itoa(int(currentUserGID))
	user := user.User{
		Uid: currentUserUIDStr,
	}
	uidToGids := map[string][]string{
		currentUserUIDStr: {currentUserGIDStr},
	}

	statCurrentUserGroupOwner := &syscall.Stat_t{
		Uid: 2000,
		Gid: currentUserGID,
	}

	mockOsInfo := mockOsInfo{
		currentUser: user,
		groupIds:    uidToGids,
	}

	directoryGroupWriteName := "group_write"
	mockFs.MapFS[directoryGroupWriteName] = &fstest.MapFile{Mode: os.ModeDir | 0220, Sys: statCurrentUserGroupOwner}

	directoryGroupNoWrite := "group_no_write"
	mockFs.MapFS[directoryGroupNoWrite] = &fstest.MapFile{Mode: os.ModeDir | 0200, Sys: statCurrentUserGroupOwner}

	statNoOwner := &syscall.Stat_t{
		Uid: 2000,
		Gid: 2000,
	}

	directoryOtherWriteName := "other_write"
	mockFs.MapFS[directoryOtherWriteName] = &fstest.MapFile{Mode: os.ModeDir | 0002, Sys: statNoOwner}

	directoryNoPermissionsName := "no_permissions"
	mockFs.MapFS[directoryNoPermissionsName] = &fstest.MapFile{Mode: os.ModeDir | 0000, Sys: statNoOwner}

	fileshareTests := []struct {
		testName          string
		transferID        string
		filePath          string
		acceptPath        string
		transferDirection pb.Direction
		transferStatus    pb.Status
		serverError       error
		respError         *pb.Error
	}{
		{
			testName:          "transfer successfully accepted",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
		},
		{
			testName:          "server error",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			serverError:       errors.New("some error"),
		},
		{
			testName:          "non-existing transfer",
			transferID:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			filePath:          filePath,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND),
		},
		{
			testName:          "outgoing transfer",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_OUTGOING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_OUTGOING),
		},
		{
			testName:          "ongoing transfer",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_ONGOING,
			respError:         fileshareError(pb.FileshareErrorCode_ALREADY_ACCEPTED),
		},
		{
			testName:          "non-existing file",
			transferID:        transferID,
			filePath:          "some_file.txt",
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND),
		},
		{
			testName:          "accept directory does not exist",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        "dddd",
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NOT_FOUND),
		},
		{
			testName:          "symlink accept directory",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        acceptSymlinkName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_A_SYMLINK),
		},
		{
			testName:          "accept directory is not a directory",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        acceptNotDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_NOT_A_DIRECTORY),
		},
		{
			testName:          "user belongs to destination directory owner group",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        directoryGroupWriteName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
		},
		{
			testName:          "destination directory has write permissions for other users",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        directoryOtherWriteName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
		},
		{
			testName:          "user has no write permissions to destination directory",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        directoryNoPermissionsName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NO_PERMISSIONS),
		},
		{
			testName:          "user belongs to owner group, owner group has no write permissions",
			transferID:        transferID,
			filePath:          filePath,
			acceptPath:        directoryGroupNoWrite,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NO_PERMISSIONS),
		},
	}

	for _, test := range fileshareTests {
		acceptServer := &mockAcceptServer{serverError: test.serverError}
		transfer := pb.Transfer{
			Id:        transferID,
			Direction: test.transferDirection,
			Status:    test.transferStatus,
			Files: []*pb.File{{
				Path: filePath,
			}},
		}
		eventManager := EventManager{
			storage: &mockStorage{
				transfers: map[string]*pb.Transfer{
					transferID: &transfer,
				},
			},
			filesystem: &mockFs,
			osInfo:     &mockOsInfo}
		server := NewServer(
			&mockServerFileshare{},
			&eventManager,
			&mockMeshClient{isEnabled: true},
			mockFs,
			&mockOsInfo,
			0,
			nil)

		t.Run(test.testName, func(t *testing.T) {
			err := server.Accept(
				&pb.AcceptRequest{TransferId: test.transferID, DstPath: test.acceptPath, Silent: true, Files: []string{test.filePath}},
				acceptServer)
			assert.ErrorIs(t, err, test.serverError)
			assert.Equal(t, test.respError, acceptServer.response.Error)
		})
	}
}

func TestAcceptDirectory(t *testing.T) {
	category.Set(t, category.Unit)

	// nested/
	// ├── a
	// └── inner
	// 	└── b
	// outer/
	// └── c
	// └── d

	nestedFilePath := "nested/a"
	nestedFileID := "nestedFileID"
	nestedInnerFilePath := "nested/inner/b"
	nestedInnerFileID := "nestedInnerFileID"
	outerFile1Path := "outer/c"
	outerFile1ID := "outerFile1ID"
	outerFile2Path := "outer/d"
	outerFile2ID := "outerFile2ID"

	file0 := pb.File{
		Path: nestedFilePath,
		Id:   nestedFileID,
		Size: uint64(10),
	}
	file1 := pb.File{
		Path: nestedInnerFilePath,
		Id:   nestedInnerFileID,
		Size: uint64(10),
	}
	file2 := pb.File{
		Path: outerFile1Path,
		Id:   outerFile1ID,
		Size: uint64(10),
	}
	file3 := pb.File{
		Path: outerFile2Path,
		Id:   outerFile2ID,
		Size: uint64(10),
	}

	transferID := exampleUUID
	transfer := pb.Transfer{
		Id:        transferID,
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
		Files:     []*pb.File{&file0, &file1, &file2, &file3},
	}

	currentUserUID := uint32(1000)
	currentUserGID := uint32(1000)

	stat_t := &syscall.Stat_t{
		Uid: currentUserUID,
		Gid: currentUserGID,
	}

	mockFs := newMockFilesystem()
	mockFs.MapFS["tmp"] = &fstest.MapFile{Mode: fs.ModeDir | 0777, Sys: stat_t}

	currentUserUIDStr := strconv.Itoa(int(currentUserUID))
	currentUserGIDStr := strconv.Itoa(int(currentUserGID))
	user := user.User{
		Uid: currentUserUIDStr,
	}
	uidToGids := map[string][]string{
		currentUserUIDStr: {currentUserGIDStr},
	}

	mockOsInfo := mockOsInfo{
		currentUser: user,
		groupIds:    uidToGids,
	}

	tests := []struct {
		testName              string
		filePaths             []string
		expectedAcceptedFiles []string
		expectedCanceledFiles []string
		firstFileErr          error
		respErr               *pb.Error
		filesystemSpace       uint64 // maxuint64 if not defined
	}{
		{
			testName:              "accept nested directory",
			filePaths:             []string{"nested"},
			expectedAcceptedFiles: []string{nestedFileID, nestedInnerFileID},
			expectedCanceledFiles: []string{outerFile1ID, outerFile2ID},
		},
		{
			testName:              "accept outer directory",
			filePaths:             []string{"outer"},
			expectedAcceptedFiles: []string{outerFile1ID, outerFile2ID},
			expectedCanceledFiles: []string{nestedFileID, nestedInnerFileID},
		},
		{
			testName:              "accept both directories",
			filePaths:             []string{"outer", "nested"},
			expectedAcceptedFiles: []string{outerFile1ID, outerFile2ID, nestedFileID, nestedInnerFileID},
			expectedCanceledFiles: []string{},
		},
		{
			testName:              "accept only nested inner",
			filePaths:             []string{"nested/inner"},
			expectedAcceptedFiles: []string{nestedInnerFileID},
			expectedCanceledFiles: []string{outerFile1ID, outerFile2ID, nestedFileID},
		},
		{
			testName:              "accept single file",
			filePaths:             []string{"outer/c"},
			expectedAcceptedFiles: []string{outerFile1ID},
			expectedCanceledFiles: []string{outerFile2ID, nestedFileID, nestedInnerFileID},
		},
		{
			testName:              "accept single file error",
			filePaths:             []string{"outer/c"},
			expectedAcceptedFiles: []string{outerFile1ID},
			expectedCanceledFiles: []string{outerFile2ID, nestedFileID, nestedInnerFileID},
			firstFileErr:          errors.New("broken file"),
			respErr:               fileshareError(pb.FileshareErrorCode_ACCEPT_ALL_FILES_FAILED),
		},
		{
			testName:              "accept partial file error",
			filePaths:             []string{"outer/c", "nested/a"},
			expectedAcceptedFiles: []string{outerFile1ID, nestedFileID},
			expectedCanceledFiles: []string{outerFile2ID, nestedInnerFileID},
			firstFileErr:          errors.New("broken file"),
			// No error expected because transfer starts with some files
		},
		{
			testName:              "not enough space",
			filePaths:             []string{"outer", "nested"},
			expectedAcceptedFiles: []string{outerFile1ID, outerFile2ID, nestedFileID, nestedInnerFileID},
			expectedCanceledFiles: []string{},
			filesystemSpace:       35,
			respErr:               fileshareError(pb.FileshareErrorCode_NOT_ENOUGH_SPACE),
		},
		{
			testName:              "enough space",
			filePaths:             []string{"outer", "nested/a"},
			expectedAcceptedFiles: []string{outerFile1ID, outerFile2ID, nestedFileID},
			expectedCanceledFiles: []string{nestedInnerFileID},
			filesystemSpace:       35,
		},
	}

	for _, test := range tests {
		transfer.Status = pb.Status_REQUESTED
		eventManager := EventManager{
			storage: &mockStorage{
				transfers: map[string]*pb.Transfer{
					transferID: &transfer,
				},
			},
			filesystem: &mockFs,
			osInfo:     &mockOsInfo}

		fileshare := &mockServerFileshare{
			acceptedFiles:          []string{},
			canceledFiles:          []string{},
			acceptFirstReturnValue: test.firstFileErr,
		}

		mockFs.freeSpace = test.filesystemSpace

		server := NewServer(
			fileshare,
			&eventManager,
			&mockMeshClient{isEnabled: true},
			mockFs,
			&mockOsInfo,
			0,
			nil,
		)

		acceptServer := &mockAcceptServer{serverError: nil}

		t.Run(test.testName, func(t *testing.T) {
			err := server.Accept(
				&pb.AcceptRequest{TransferId: transferID, DstPath: "tmp", Silent: true, Files: test.filePaths},
				acceptServer)
			assert.Equal(t, err, nil)
			if test.respErr != nil {
				assert.Equal(t, test.respErr, acceptServer.response.Error)
				return
			}
			assert.True(t, isFileListEqual(t, test.expectedAcceptedFiles, fileshare.acceptedFiles),
				"expected %v, got %v", test.expectedAcceptedFiles, fileshare.acceptedFiles)
		})
	}
}

func TestCancel(t *testing.T) {
	category.Set(t, category.Unit)

	filePath := "norddrop_tests/test_a.txt"
	fileID := "file ID"
	file := pb.File{
		Path: filePath,
		Id:   fileID,
	}

	transferID := exampleUUID
	transfer := pb.Transfer{
		Id:    transferID,
		Files: []*pb.File{&file},
	}

	fileshareTests := []struct {
		testName       string
		isMeshEnabled  bool
		cancelError    error
		transferID     string
		filePath       string
		transferStatus pb.Status
		response       *pb.Error
	}{
		{
			testName:      "mesh not enabled",
			isMeshEnabled: false,
			response:      serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED),
		},
		{
			testName:      "transfer not found",
			isMeshEnabled: true,
			cancelError:   nil,
			transferID:    "b537743c-a328-4a3e-b2ec",
			filePath:      "",
			response:      fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND),
		},
		{
			testName:       "file not found",
			isMeshEnabled:  true,
			cancelError:    nil,
			transferID:     transferID,
			filePath:       "norddrop_tests/test.txt",
			transferStatus: pb.Status_ONGOING,
			response:       fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND),
		},
		{
			testName:       "lib failure",
			isMeshEnabled:  true,
			cancelError:    errors.New("generic error"),
			transferID:     transferID,
			filePath:       filePath,
			transferStatus: pb.Status_ONGOING,
			response:       fileshareError(pb.FileshareErrorCode_LIB_FAILURE),
		},
		{
			testName:       "double cancel",
			isMeshEnabled:  true,
			cancelError:    nil,
			transferID:     transferID,
			filePath:       filePath,
			transferStatus: pb.Status_CANCELED,
			response:       fileshareError(pb.FileshareErrorCode_FILE_INVALIDATED),
		},
		{
			testName:       "cancel success",
			isMeshEnabled:  true,
			cancelError:    nil,
			transferID:     transferID,
			filePath:       filePath,
			transferStatus: pb.Status_ONGOING,
			response:       empty(),
		},
	}

	for _, test := range fileshareTests {
		transfer.Files[0].Status = test.transferStatus
		eventManager := EventManager{
			storage: &mockStorage{
				transfers: map[string]*pb.Transfer{
					transferID: &transfer,
				},
			},
		}

		server := NewServer(
			&mockServerFileshare{cancelReturnValue: test.cancelError},
			&eventManager,
			&mockMeshClient{isEnabled: test.isMeshEnabled},
			newMockFilesystem(),
			&mockOsInfo{},
			0,
			nil,
		)

		t.Run(test.testName, func(t *testing.T) {
			resp, err := server.CancelFile(context.Background(), &pb.CancelFileRequest{TransferId: test.transferID, FilePath: test.filePath})
			assert.NoError(t, err)
			assert.Equal(t, test.response, resp)
		})
	}
}

func TestList(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name              string
		numberOfTransfers int
		expectedChunks    int
	}{
		{
			name:              "less than one chunk",
			numberOfTransfers: 2,
			expectedChunks:    1,
		},
		{
			name:              "one chunk",
			numberOfTransfers: 5,
			expectedChunks:    1,
		},
		{
			name:              "one and a half chunk",
			numberOfTransfers: 7,
			expectedChunks:    2,
		},
		{
			name:              "two chunks",
			numberOfTransfers: 10,
			expectedChunks:    2,
		},
	}

	for _, test := range tests {
		expectedTransfers := getTransfers(t, test.numberOfTransfers)
		storage := &mockStorage{
			transfers: expectedTransfers,
		}
		eventManager := EventManager{
			storage: storage,
		}

		server := NewServer(
			&mockEventManagerFileshare{},
			&eventManager,
			&mockMeshClient{isEnabled: true},
			newMockFilesystem(),
			&mockOsInfo{},
			5,
			nil,
		)

		listServer := mockListServer{}

		t.Run(test.name, func(t *testing.T) {
			server.List(&pb.Empty{}, &listServer)

			assert.Len(t,
				listServer.responses,
				test.expectedChunks,
				"Invalid number of chunks received: Expected: %d, Got: %d",
				test.expectedChunks,
				len(listServer.responses))

			receivedTransfers := map[string]*pb.Transfer{}
			for _, response := range listServer.responses {
				for _, transfer := range response.Transfers {
					receivedTransfers[transfer.Id] = transfer
				}
			}

			assert.Equal(t,
				expectedTransfers,
				receivedTransfers,
				"Invalid transfers received from server.")
		})
	}
}
