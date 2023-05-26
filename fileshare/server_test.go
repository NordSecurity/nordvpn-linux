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

	"github.com/stretchr/testify/assert"
)

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

func findFileInList(t *testing.T, id string, files []*pb.File) *pb.File {
	t.Helper()

	for _, file := range files {
		if file := findFileInTree(t, id, file); file != nil {
			return file
		}
	}

	return nil
}

func findFileInTree(t *testing.T, id string, tree *pb.File) *pb.File {
	t.Helper()

	if tree.Id == id {
		return tree
	}

	for _, child := range tree.Children {
		if file := findFileInTree(t, id, child); file != nil {
			return file
		}
	}

	return nil
}

// checkFilesStatus returns file/status pair for all fileIDs that do not have the expected status
// and filenames of all of the fileIDs not found in files
func checkFilesStatus(t *testing.T, fileIDs []string, files []*pb.File, status pb.Status) ([]string, []*pb.File) {
	notFoundFileIDs := []string{}
	invalidStatusFiles := []*pb.File{}
	for _, fileID := range fileIDs {
		file := findFileInList(t, fileID, files)
		if file == nil {
			notFoundFileIDs = append(notFoundFileIDs, fileID)
		} else if file.Status != status {
			invalidStatusFiles = append(invalidStatusFiles, file)
		}
	}

	return notFoundFileIDs, invalidStatusFiles
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
	response        pb.StatusResponse
	sendReturnValue error
}

func (m *mockSendServer) Send(resp *pb.StatusResponse) error {
	m.response = *resp //nolint:govet
	return m.sendReturnValue
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
			loacalPeers:   localPeers,
			externalPeers: externalPeers,
		}

		mockFileshare := mockServerFileshare{}
		server := NewServer(
			&mockFileshare,
			&EventManager{transfers: make(map[string]*pb.Transfer)},
			mockMeshClient,
			mockFs,
			&mockOsInfo{},
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

	file1 := "file1"
	mockFs.MapFS[file1] = &fstest.MapFile{}
	file2 := "file2"
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
			&EventManager{transfers: make(map[string]*pb.Transfer)},
			mockMeshClient{isEnabled: true},
			mockFs,
			&mockOsInfo{},
		)

		sendServer := mockSendServer{}

		t.Run(test.testName, func(t *testing.T) {
			err := server.Send(
				&pb.SendRequest{Peer: "100.96.115.182", Paths: test.paths, Silent: test.transferSilent},
				&sendServer)
			assert.Equal(t, nil, err)
			assert.Equal(t, test.expectedSendResponse, &sendServer.response)
		})
	}
}

func TestAccept(t *testing.T) {
	category.Set(t, category.Unit)

	fileID := "test_a.txt"
	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"

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
		fileID            string
		acceptPath        string
		transferDirection pb.Direction
		transferStatus    pb.Status
		serverError       error
		respError         *pb.Error
	}{
		{
			testName:          "transfer successfully accepted",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
		},
		{
			testName:          "server error",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			serverError:       errors.New("some error"),
		},
		{
			testName:          "non-existing transfer",
			transferID:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			fileID:            fileID,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND),
		},
		{
			testName:          "outgoing transfer",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_OUTGOING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_OUTGOING),
		},
		{
			testName:          "ongoing transfer",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_ONGOING,
			respError:         fileshareError(pb.FileshareErrorCode_ALREADY_ACCEPTED),
		},
		{
			testName:          "non-existing file",
			transferID:        transferID,
			fileID:            "some_file.txt",
			acceptPath:        acceptDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND),
		},
		{
			testName:          "accept directory does not exist",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        "dddd",
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NOT_FOUND),
		},
		{
			testName:          "symlink accept directory",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        acceptSymlinkName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_A_SYMLINK),
		},
		{
			testName:          "accept directory is not a directory",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        acceptNotDirName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_NOT_A_DIRECTORY),
		},
		{
			testName:          "user belongs to destination directory owner group",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        directoryGroupWriteName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
		},
		{
			testName:          "destination directory has write permissions for other users",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        directoryOtherWriteName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
		},
		{
			testName:          "user has no write permissions to destination directory",
			transferID:        transferID,
			fileID:            fileID,
			acceptPath:        directoryNoPermissionsName,
			transferDirection: pb.Direction_INCOMING,
			transferStatus:    pb.Status_REQUESTED,
			respError:         fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NO_PERMISSIONS),
		},
		{
			testName:          "user belongs to owner group, owner group has no write permissions",
			transferID:        transferID,
			fileID:            fileID,
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
				Id: fileID,
			}},
		}
		eventManager := EventManager{transfers: map[string]*pb.Transfer{
			transferID: &transfer,
		}}
		server := NewServer(
			&mockServerFileshare{},
			&eventManager,
			mockMeshClient{isEnabled: true},
			mockFs,
			&mockOsInfo{
				currentUser: user,
				groupIds:    uidToGids,
			})

		t.Run(test.testName, func(t *testing.T) {
			err := server.Accept(
				&pb.AcceptRequest{TransferId: test.transferID, DstPath: test.acceptPath, Silent: true, Files: []string{test.fileID}},
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

	nestedDirectoryID := "nested"
	innerDirectoryID := "nested/inner"
	outerDirectoryID := "outer"

	file0ID := "nested/a"
	file1ID := "nested/inner/b"
	file2ID := "outer/c"
	file3ID := "outer/d"

	file0 := pb.File{
		Id:   file0ID,
		Size: uint64(10),
	}
	file1 := pb.File{
		Id:   file1ID,
		Size: uint64(10),
	}
	file2 := pb.File{
		Id:   file2ID,
		Size: uint64(10),
	}
	file3 := pb.File{
		Id:   file3ID,
		Size: uint64(10),
	}

	innerDirectory := pb.File{
		Id:       innerDirectoryID,
		Children: map[string]*pb.File{file1ID: &file0},
	}
	nestedDirectory := pb.File{
		Id:       nestedDirectoryID,
		Children: map[string]*pb.File{file1ID: &file1, innerDirectoryID: &innerDirectory},
	}
	outerDirectory := pb.File{
		Id:       outerDirectoryID,
		Children: map[string]*pb.File{file1ID: &file2, file3ID: &file3},
	}

	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"
	transfer := pb.Transfer{
		Id:        transferID,
		Direction: pb.Direction_INCOMING,
		Status:    pb.Status_REQUESTED,
		Files:     []*pb.File{&nestedDirectory, &outerDirectory},
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

	tests := []struct {
		testName              string
		fileIDs               []string
		expectedAcceptedFiles []string
		expectedCanceledFiles []string
		firstFileErr          error
		respErr               *pb.Error
		filesystemSpace       uint64 // maxuint64 if not defined
	}{
		{
			testName:              "accept nested directory",
			fileIDs:               []string{"nested"},
			expectedAcceptedFiles: []string{"nested/a", "nested/inner/b"},
			expectedCanceledFiles: []string{"outer/c", "outer/d"},
		},
		{
			testName:              "accept outer directory",
			fileIDs:               []string{"outer"},
			expectedAcceptedFiles: []string{"outer/c", "outer/d"},
			expectedCanceledFiles: []string{"nested/a", "nested/inner/b"},
		},
		{
			testName:              "accept both directories",
			fileIDs:               []string{"outer", "nested"},
			expectedAcceptedFiles: []string{"outer/c", "outer/d", "nested/a", "nested/inner/b"},
			expectedCanceledFiles: []string{},
		},
		{
			testName:              "accept only nested inner",
			fileIDs:               []string{"nested/inner"},
			expectedAcceptedFiles: []string{"nested/inner/b"},
			expectedCanceledFiles: []string{"outer/c", "outer/d", "nested/a"},
		},
		{
			testName:              "accept single file",
			fileIDs:               []string{"outer/c"},
			expectedAcceptedFiles: []string{"outer/c"},
			expectedCanceledFiles: []string{"outer/d", "nested/a", "nested/inner/b"},
		},
		{
			testName:              "accept single file error",
			fileIDs:               []string{"outer/c"},
			expectedAcceptedFiles: []string{"outer/c"},
			expectedCanceledFiles: []string{"outer/d", "nested/a", "nested/inner/b"},
			firstFileErr:          errors.New("broken file"),
			respErr:               fileshareError(pb.FileshareErrorCode_ACCEPT_ALL_FILES_FAILED),
		},
		{
			testName:              "accept partial file error",
			fileIDs:               []string{"outer/c", "nested/a"},
			expectedAcceptedFiles: []string{"outer/c", "nested/a"},
			expectedCanceledFiles: []string{"outer/d", "nested/inner/b"},
			firstFileErr:          errors.New("broken file"),
			// No error expected because transfer starts with some files
		},
		{
			testName:              "not enough space",
			fileIDs:               []string{"outer", "nested"},
			expectedAcceptedFiles: []string{"outer/c", "outer/d", "nested/a", "nested/inner/b"},
			expectedCanceledFiles: []string{},
			filesystemSpace:       35,
			respErr:               fileshareError(pb.FileshareErrorCode_NOT_ENOUGH_SPACE),
		},
		{
			testName:              "enough space",
			fileIDs:               []string{"outer", "nested/a"},
			expectedAcceptedFiles: []string{"outer/c", "outer/d", "nested/a"},
			expectedCanceledFiles: []string{"nested/inner/b"},
			filesystemSpace:       35,
		},
	}

	for _, test := range tests {
		transfer.Status = pb.Status_REQUESTED
		eventManager := EventManager{transfers: map[string]*pb.Transfer{
			transferID: &transfer,
		}}

		fileshare := &mockServerFileshare{
			acceptedFiles:          []string{},
			canceledFiles:          []string{},
			acceptFirstReturnValue: test.firstFileErr,
		}

		mockFs.freeSpace = test.filesystemSpace

		server := NewServer(
			fileshare,
			&eventManager,
			mockMeshClient{isEnabled: true},
			mockFs,
			&mockOsInfo{
				currentUser: user,
				groupIds:    uidToGids,
			},
		)

		acceptServer := &mockAcceptServer{serverError: nil}

		t.Run(test.testName, func(t *testing.T) {
			err := server.Accept(
				&pb.AcceptRequest{TransferId: transferID, DstPath: "tmp", Silent: true, Files: test.fileIDs},
				acceptServer)
			assert.Equal(t, err, nil)
			if test.respErr != nil {
				assert.Equal(t, test.respErr, acceptServer.response.Error)
				return
			}
			assert.True(t, isFileListEqual(t, test.expectedAcceptedFiles, fileshare.acceptedFiles),
				"expected %v, got %v", test.expectedAcceptedFiles, fileshare.acceptedFiles)

			notFoundFileIDs, invalidStatusFiles := checkFilesStatus(t, test.expectedCanceledFiles, eventManager.transfers[transferID].Files, pb.Status_CANCELED)
			assert.Equal(t, len(notFoundFileIDs), 0, "not all file IDs from %v found in transfer files list, missing files: %v", test.expectedCanceledFiles, notFoundFileIDs)
			assert.Equal(t, len(invalidStatusFiles), 0, "not all file IDs from %v are canceled, files with invalid status: %v", test.expectedCanceledFiles, invalidStatusFiles)
		})
	}
}

func TestCancel(t *testing.T) {
	category.Set(t, category.Unit)

	fileID := "norddrop_tests/test_a.txt"
	file := pb.File{
		Id: fileID,
	}

	transferID := "b537743c-a328-4a3e-b2ec-fc87f98c2164"
	transfer := pb.Transfer{
		Id:    transferID,
		Files: []*pb.File{&file},
	}

	fileshareTests := []struct {
		testName       string
		isMeshEnabled  bool
		cancelError    error
		transferID     string
		fileID         string
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
			fileID:        "",
			response:      fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND),
		},
		{
			testName:       "file not found",
			isMeshEnabled:  true,
			cancelError:    nil,
			transferID:     transferID,
			fileID:         "norddrop_tests/test.txt",
			transferStatus: pb.Status_ONGOING,
			response:       fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND),
		},
		{
			testName:       "lib failure",
			isMeshEnabled:  true,
			cancelError:    errors.New("generic error"),
			transferID:     transferID,
			fileID:         fileID,
			transferStatus: pb.Status_ONGOING,
			response:       fileshareError(pb.FileshareErrorCode_LIB_FAILURE),
		},
		{
			testName:       "double cancel",
			isMeshEnabled:  true,
			cancelError:    nil,
			transferID:     transferID,
			fileID:         fileID,
			transferStatus: pb.Status_CANCELED,
			response:       fileshareError(pb.FileshareErrorCode_FILE_INVALIDATED),
		},
		{
			testName:       "cancel success",
			isMeshEnabled:  true,
			cancelError:    nil,
			transferID:     transferID,
			fileID:         fileID,
			transferStatus: pb.Status_ONGOING,
			response:       empty(),
		},
	}

	for _, test := range fileshareTests {
		eventManager := EventManager{transfers: map[string]*pb.Transfer{
			transferID: &transfer,
		}}

		server := NewServer(
			&mockServerFileshare{cancelReturnValue: test.cancelError},
			&eventManager,
			mockMeshClient{isEnabled: test.isMeshEnabled},
			newMockFilesystem(),
			&mockOsInfo{},
		)
		eventManager.transfers[transferID].Files[0].Status = test.transferStatus

		t.Run(test.testName, func(t *testing.T) {
			resp, err := server.CancelFile(context.Background(), &pb.CancelFileRequest{TransferId: test.transferID, FileId: test.fileID})
			assert.NoError(t, err)
			assert.Equal(t, test.response, resp)
		})
	}
}
