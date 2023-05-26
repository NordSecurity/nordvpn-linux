package fileshare

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/netip"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"golang.org/x/exp/slices"
)

// Pre-built values for commonly returned responses to decrease verbosity
func empty() *pb.Error {
	return &pb.Error{Response: &pb.Error_Empty{}}
}

func serviceError(code pb.ServiceErrorCode) *pb.Error {
	return &pb.Error{Response: &pb.Error_ServiceError{
		ServiceError: code,
	}}
}

func fileshareError(code pb.FileshareErrorCode) *pb.Error {
	return &pb.Error{Response: &pb.Error_FileshareError{
		FileshareError: code,
	}}
}

// Server implements fileshare rpc receiver
type Server struct {
	pb.UnimplementedFileshareServer
	// Errors on Fileshare methods shouldn't be logged, because they are logged by the library itself.
	fileshare    Fileshare
	eventManager *EventManager
	meshClient   meshpb.MeshnetClient
	filesystem   Filesystem
	osInfo       OsInfo
}

// NewServer is a default constructor for a fileshare server
func NewServer(
	fileshare Fileshare,
	eventManager *EventManager,
	meshClient meshpb.MeshnetClient,
	filesystem Filesystem,
	osInfo OsInfo,
) *Server {
	return &Server{
		fileshare:    fileshare,
		eventManager: eventManager,
		meshClient:   meshClient,
		filesystem:   filesystem,
		osInfo:       osInfo,
	}
}

func (s *Server) isDirectory(path string) (bool, error) {
	fileInfo, err := s.filesystem.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

var (
	errMaxDirectoryDepthReached = errors.New("Max directory depth reached")
	errGetPeersFailed           = errors.New("Failed to get peers from meshnet daemon")
)

// getNumberOfFiles returns number of files in a directory and its subdirectories
// Returns an error if max subdirectory deepth exceeds max_depth, pass a negative number for infinite depth
func (s *Server) getNumberOfFiles(path string, maxDepth int) (int, error) {
	if maxDepth == 0 {
		return 0, errMaxDirectoryDepthReached
	}

	files, err := s.filesystem.ReadDir(path)

	if err != nil {
		return 0, err
	}

	numberOfFiles := 0
	maxDepth--

	for _, file := range files {
		if file.IsDir() {
			nestedFiles, err := s.getNumberOfFiles(path+"/"+file.Name(), maxDepth)
			if err != nil {
				return 0, err
			}
			numberOfFiles += nestedFiles
		} else {
			numberOfFiles++
		}
	}

	return numberOfFiles, err
}

func (s *Server) startTransferStatusStream(srv pb.Fileshare_SendServer, transferID string) error {
	for ev := range s.eventManager.Subscribe(transferID) {
		//exhaustive:ignore
		switch ev.Status {
		case pb.Status_ONGOING:
			if err := srv.Send(&pb.StatusResponse{
				TransferId: ev.TransferID,
				Progress:   ev.Transferred,
				Status:     pb.Status_ONGOING}); err != nil {
				log.Printf("error while streaming transfer %s status: %s", transferID, err)
			}
		case pb.Status_SUCCESS:
			return srv.Send(&pb.StatusResponse{TransferId: ev.TransferID, Status: pb.Status_SUCCESS})
		case pb.Status_FINISHED_WITH_ERRORS:
			return srv.Send(&pb.StatusResponse{TransferId: ev.TransferID, Status: pb.Status_FINISHED_WITH_ERRORS})
		case pb.Status_CANCELED_BY_PEER:
			return srv.Send(&pb.StatusResponse{TransferId: ev.TransferID, Status: pb.Status_CANCELED_BY_PEER})
		case pb.Status_CANCELED:
			return srv.Send(&pb.StatusResponse{TransferId: ev.TransferID, Status: pb.Status_CANCELED})
		}
	}
	return nil
}

// getPeers returns map where peer ip/hostname/pubkey maps to *meshpb.Peer
func (s *Server) getPeers() (map[string]*meshpb.Peer, error) {
	resp, err := s.meshClient.GetPeers(context.Background(), &meshpb.Empty{})

	if err != nil {
		log.Printf("GetPeers failed: %s", err)
		return nil, errGetPeersFailed
	}

	switch resp := resp.Response.(type) {
	case *meshpb.GetPeersResponse_Peers:
		peerNameToPeer := make(map[string]*meshpb.Peer)
		for _, peer := range append(resp.Peers.External, resp.Peers.Local...) {
			peerNameToPeer[peer.Ip] = peer
			peerNameToPeer[peer.Hostname] = peer
			peerNameToPeer[peer.Pubkey] = peer
		}
		return peerNameToPeer, nil
	case *meshpb.GetPeersResponse_ServiceErrorCode:
		log.Printf("GetPeers failed, service error: %s", meshpb.ServiceErrorCode_name[int32(resp.ServiceErrorCode)])
		return nil, errGetPeersFailed
	case *meshpb.GetPeersResponse_MeshnetErrorCode:
		log.Printf("GetPeers failed, meshnet error: %s", meshpb.ServiceErrorCode_name[int32(resp.MeshnetErrorCode)])
		return nil, errGetPeersFailed
	default:
		log.Printf("GetPeers failed, unknown error")
		return nil, errGetPeersFailed
	}
}

// Ping rpc
func (*Server) Ping(ctx context.Context, _ *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

// Send rpc
func (s *Server) Send(req *pb.SendRequest, srv pb.Fileshare_SendServer) error {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return srv.Send(&pb.StatusResponse{Error: serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED)})
	}

	fileCount := 0
	for _, path := range req.Paths {
		isDirectory, err := s.isDirectory(path)

		if err != nil {
			return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND)})
		}

		if isDirectory {
			fileCountInDirectory, err := s.getNumberOfFiles(path, DirDepthLimit)
			switch err {
			case errMaxDirectoryDepthReached:
				return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_DIRECTORY_TOO_DEEP)})
			case nil:
				fileCount += fileCountInDirectory
			default:
				return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND)})
			}
		} else {
			fileCount++
		}

		if fileCount > TransferFileLimit {
			return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_TOO_MANY_FILES)})
		}

		if fileCount == 0 {
			return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_NO_FILES)})
		}
	}

	peers, err := s.getPeers()
	if err != nil {
		return srv.Send(&pb.StatusResponse{Error: serviceError(pb.ServiceErrorCode_INTERNAL_FAILURE)})
	}

	peer, ok := peers[req.Peer]
	if !ok {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_INVALID_PEER)})
	}

	if peer.Status == meshpb.PeerStatus_DISCONNECTED {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_PEER_DISCONNECTED)})
	}

	parsedIP, err := netip.ParseAddr(peer.Ip)
	if err != nil {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_INVALID_PEER)})
	}

	if !peer.IsFileshareAllowed {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_SENDING_NOT_ALLOWED)})
	}

	transferID, err := s.fileshare.Send(parsedIP, req.Paths)
	if err != nil {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_CREATED)})
	}

	if len(req.Paths) > 1 {
		s.eventManager.NewOutgoingTransfer(transferID, peer.Ip, "multiple files")
	} else {
		s.eventManager.NewOutgoingTransfer(transferID, peer.Ip, req.Paths[0])
	}

	// Ignore response here
	fileName := ""
	if len(req.Paths) == 1 {
		fileName = req.Paths[0]
	}
	go s.meshClient.NotifyNewTransfer(context.Background(), &meshpb.NewTransferNotification{
		Identifier: peer.Identifier,
		// This is needed because currently nordvpnd queries API every time when looking
		// for a specific peer. Since this is currently needed only for iOS, this will
		// allows not to call API on every new transfer
		Os:        peer.Os,
		FileName:  fileName,
		FileCount: int32(len(req.Paths)),
	})

	if err := srv.Send(&pb.StatusResponse{TransferId: transferID, Status: pb.Status_REQUESTED}); err != nil {
		return err
	}

	if req.GetSilent() { // report no progress back, if asked
		return nil
	}

	return s.startTransferStatusStream(srv, transferID)
}

func isFileWriteable(fileInfo fs.FileInfo, user *user.User, gids []string) bool {
	var ownerUID int
	var ownerGID int
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		ownerUID = int(stat.Uid)
		ownerGID = int(stat.Gid)
	} else {
		return false
	}

	uid, err := strconv.Atoi(user.Uid)

	if err != nil {
		log.Printf("Failed to convert uid %s to int: %s", user.Uid, err)
		return false
	}

	isOwner := uid == ownerUID

	if isOwner {
		return fileInfo.Mode().Perm()&os.FileMode(0200) != 0
	}

	ownerGIDStr := strconv.Itoa(ownerGID)
	gidIndex := slices.Index(gids, ownerGIDStr)
	isGroup := gidIndex != -1
	if isGroup {
		return fileInfo.Mode().Perm()&os.FileMode(0020) != 0
	}

	return fileInfo.Mode().Perm()&os.FileMode(0002) != 0
}

// Accept rpc
func (s *Server) Accept(req *pb.AcceptRequest, srv pb.Fileshare_AcceptServer) error {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return srv.Send(&pb.StatusResponse{Error: serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED)})
	}

	destinationFileInfo, err := s.filesystem.Lstat(req.DstPath)

	if err != nil {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NOT_FOUND)})
	}

	if destinationFileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_A_SYMLINK)})
	}

	if !destinationFileInfo.IsDir() {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_NOT_A_DIRECTORY)})
	}

	statfs, err := s.filesystem.Statfs(req.DstPath)
	if err != nil {
		log.Printf("doing statfs: %s", err)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_NOT_ENOUGH_SPACE)})
	}

	fileInfo, err := s.filesystem.Stat(req.DstPath)
	if err != nil {
		log.Printf("doing stat: %s", err)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode(pb.ServiceErrorCode_INTERNAL_FAILURE))})
	}

	userInfo, err := s.osInfo.CurrentUser()
	if err != nil {
		log.Printf("getting user info: %s", err)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode(pb.ServiceErrorCode_INTERNAL_FAILURE))})
	}

	userGroups, err := s.osInfo.GetGroupIds(userInfo)
	if err != nil {
		log.Printf("getting user groups: %s", err)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode(pb.ServiceErrorCode_INTERNAL_FAILURE))})
	}

	if !isFileWriteable(fileInfo, userInfo, userGroups) {
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode(pb.FileshareErrorCode_ACCEPT_DIR_NO_PERMISSIONS))})
	}

	transfer, err := s.eventManager.AcceptTransfer(req.TransferId, req.DstPath, req.Files, statfs.Bavail*uint64(statfs.Bsize))

	switch err {
	case ErrTransferNotFound:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND)})
	case ErrTransferAcceptOutgoing:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_OUTGOING)})
	case ErrTransferAlreadyAccepted:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ALREADY_ACCEPTED)})
	case ErrFileNotFound:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND)})
	case ErrSizeLimitExceeded:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_NOT_ENOUGH_SPACE)})
	case nil:
		break
	default:
		log.Printf("error while accepting transfer %s: %s", req.TransferId, err)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_LIB_FAILURE)})
	}

	transferStarted := false
	// if user has given command to accept only one (or some) file in whole transfer
	// given files should be accepted, but other files has to be canceled for whole transfer to get processed at once
	for _, file := range GetAllTransferFiles(transfer) {
		isAccepted := len(req.Files) == 0 || slices.ContainsFunc(req.Files,
			func(acceptedFileId string) bool {
				// user can provide a directory name in order to accept multiple files, so we use HasPrefix instead of comparing ids directly
				return strings.HasPrefix(file.Id, acceptedFileId)
			})

		if isAccepted {
			if err := s.fileshare.Accept(req.TransferId, req.DstPath, file.Id); err == nil {
				transferStarted = true
			}
		} else {
			s.eventManager.SetFileStatus(transfer.Id, file.Id, pb.Status_CANCELED)
		}
	}

	if !transferStarted {
		// Setting transfer status because it will not be set by events because transfer
		// is not even started.
		// Also not handling possible error because we are already in error state.
		_ = s.eventManager.SetTransferStatus(transfer.Id, pb.Status_ACCEPT_FAILURE)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_ALL_FILES_FAILED)})
	}

	if err := srv.Send(&pb.StatusResponse{TransferId: transfer.Id, Status: pb.Status_REQUESTED}); err != nil {
		return err
	}

	if req.GetSilent() { // report no progress back, if asked
		return nil
	}

	return s.startTransferStatusStream(srv, transfer.Id)
}

// Cancel rpc
func (s *Server) Cancel(
	ctx context.Context,
	req *pb.CancelRequest,
) (*pb.Error, error) {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED), nil
	}

	transfer, err := s.eventManager.GetTransfer(req.GetTransferId())
	switch err {
	case ErrTransferNotFound:
		return fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND), nil
	case nil:
		break
	default:
		log.Printf("error while cancelling transfer %s: %s", req.TransferId, err)
		return fileshareError(pb.FileshareErrorCode_LIB_FAILURE), nil
	}

	if transfer.Status != pb.Status_ONGOING && transfer.Status != pb.Status_REQUESTED {
		return fileshareError(pb.FileshareErrorCode_TRANSFER_INVALIDATED), nil
	}

	if err := s.fileshare.Cancel(transfer.Id); err != nil {
		return fileshareError(pb.FileshareErrorCode_LIB_FAILURE), nil
	}

	return empty(), nil
}

// List rpc
func (s *Server) List(ctx context.Context, _ *pb.Empty) (*pb.ListResponse, error) {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return &pb.ListResponse{Error: serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED)}, nil
	}

	peers, err := s.getPeers()
	if err != nil {
		return &pb.ListResponse{Error: serviceError(pb.ServiceErrorCode_INTERNAL_FAILURE)}, nil
	}

	transfers := s.eventManager.GetTransfers()
	for _, transfer := range transfers {
		if peer, ok := peers[transfer.Peer]; ok {
			transfer.Peer = peer.Hostname
		}
	}

	return &pb.ListResponse{
		Error:     empty(),
		Transfers: transfers,
	}, nil
}

// CancelFile rpc
func (s *Server) CancelFile(ctx context.Context, req *pb.CancelFileRequest) (*pb.Error, error) {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED), nil
	}

	transfer, err := s.eventManager.GetTransfer(req.TransferId)
	switch err {
	case ErrTransferNotFound:
		return fileshareError(pb.FileshareErrorCode_TRANSFER_NOT_FOUND), nil
	case nil:
		break
	default:
		log.Printf("error while cancelling transfer %s: %s", req.TransferId, err)
		return fileshareError(pb.FileshareErrorCode_LIB_FAILURE), nil
	}

	file := FindTransferFile(transfer, req.FileId)

	if file == nil {
		return fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND), nil
	}

	if file.Status == pb.Status_CANCELED {
		return fileshareError(pb.FileshareErrorCode_FILE_INVALIDATED), nil
	}

	if file.Status != pb.Status_ONGOING {
		return fileshareError(pb.FileshareErrorCode_FILE_NOT_IN_PROGRESS), nil
	}

	if err := s.fileshare.CancelFile(req.TransferId, req.FileId); err != nil {
		log.Printf("failed to cancel file in transfer %s: %s", req.TransferId, err)
		return fileshareError(pb.FileshareErrorCode_LIB_FAILURE), nil
	}

	return empty(), nil
}

func (s *Server) SetNotifications(ctx context.Context, in *pb.SetNotificationsRequest) (*pb.SetNotificationsResponse, error) {
	if s.eventManager.AreNotificationsEnabled() == in.Enable {
		return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_NOTHING_TO_DO}, nil
	}

	if in.Enable {
		if err := s.eventManager.EnableNotifications(); err != nil {
			log.Println("Failed to enable notifications: ", err)
		}
	} else {
		s.eventManager.DisableNotifications()
	}

	return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_SET_SUCCESS}, nil
}
