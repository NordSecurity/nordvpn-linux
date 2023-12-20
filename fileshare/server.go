package fileshare

import (
	"context"
	"errors"
	"log"
	"net/netip"
	"strings"

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
	fileshare     Fileshare
	eventManager  *EventManager
	meshClient    meshpb.MeshnetClient
	filesystem    Filesystem
	osInfo        OsInfo
	listChunkSize int
}

// NewServer is a default constructor for a fileshare server
func NewServer(
	fileshare Fileshare,
	eventManager *EventManager,
	meshClient meshpb.MeshnetClient,
	filesystem Filesystem,
	osInfo OsInfo,
	listChunkSize int,
) *Server {
	return &Server{
		fileshare:     fileshare,
		eventManager:  eventManager,
		meshClient:    meshClient,
		filesystem:    filesystem,
		osInfo:        osInfo,
		listChunkSize: listChunkSize,
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
	errMaxDirectoryDepthReached = errors.New("max directory depth reached")
	errGetPeersFailed           = errors.New("failed to get peers from meshnet daemon")
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

// Accept rpc
func (s *Server) Accept(req *pb.AcceptRequest, srv pb.Fileshare_AcceptServer) error {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return srv.Send(&pb.StatusResponse{Error: serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED)})
	}

	transfer, err := s.eventManager.AcceptTransfer(req.TransferId, req.DstPath, req.Files)

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
	case ErrAcceptDirNotFound:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NOT_FOUND)})
	case ErrAcceptDirIsASymlink:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_A_SYMLINK)})
	case ErrAcceptDirIsNotADirectory:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_IS_NOT_A_DIRECTORY)})
	case ErrNoPermissionsToAcceptDirectory:
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_ACCEPT_DIR_NO_PERMISSIONS)})
	case nil:
		break
	default:
		log.Printf("error while accepting transfer %s: %s", req.TransferId, err)
		return srv.Send(&pb.StatusResponse{Error: fileshareError(pb.FileshareErrorCode_LIB_FAILURE)})
	}

	transferStarted := false
	// if user has given command to accept only one (or some) file in whole transfer
	// given files should be accepted, but other files has to be canceled for whole transfer to get processed at once
	for _, file := range transfer.Files {
		isAccepted := len(req.Files) == 0 || slices.ContainsFunc(req.Files,
			func(acceptedFilePath string) bool {
				// user can provide a directory name in order to accept multiple files, so we use HasPrefix instead of comparing ids directly
				return strings.HasPrefix(file.Path, acceptedFilePath)
			})

		if isAccepted {
			if err := s.fileshare.Accept(req.TransferId, req.DstPath, file.Id); err != nil {
				log.Printf("error accepting file %s in transfer %s: %s", file.Id, req.TransferId, err)
			} else {
				transferStarted = true
			}
		} else {
			if err := s.fileshare.CancelFile(req.TransferId, file.Id); err != nil {
				log.Printf("error cancelling file %s in transfer %s: %s", file.Id, req.TransferId, err)
			}
		}
	}

	if !transferStarted {
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
func (s *Server) List(_ *pb.Empty, srv pb.Fileshare_ListServer) error {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return srv.Send(&pb.ListResponse{Error: serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED)})
	}

	peers, err := s.getPeers()
	if err != nil {
		return srv.Send(&pb.ListResponse{Error: serviceError(pb.ServiceErrorCode_INTERNAL_FAILURE)})
	}

	transfers, err := s.eventManager.GetTransfers()
	if err != nil {
		log.Printf("getting transfer list: %s", err)
		return srv.Send(&pb.ListResponse{Error: fileshareError(pb.FileshareErrorCode_LIB_FAILURE)})
	}
	for _, transfer := range transfers {
		if peer, ok := peers[transfer.Peer]; ok {
			transfer.Peer = peer.Hostname
		}
	}

	for chunkStart := 0; chunkStart < len(transfers); chunkStart += s.listChunkSize {
		chunk := transfers[chunkStart:]
		if len(chunk) < s.listChunkSize {
			return srv.Send(&pb.ListResponse{
				Error:     empty(),
				Transfers: chunk,
			})
		}

		err := srv.Send(&pb.ListResponse{
			Error:     empty(),
			Transfers: chunk[:s.listChunkSize],
		})
		if err != nil {
			return err
		}
	}

	return nil
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

	file := FindTransferFileByPath(transfer, req.FilePath)

	if file == nil {
		return fileshareError(pb.FileshareErrorCode_FILE_NOT_FOUND), nil
	}

	if file.Status == pb.Status_CANCELED || file.Status == pb.Status_SUCCESS {
		return fileshareError(pb.FileshareErrorCode_FILE_INVALIDATED), nil
	}

	if file.Status != pb.Status_ONGOING {
		return fileshareError(pb.FileshareErrorCode_FILE_NOT_IN_PROGRESS), nil
	}

	if err := s.fileshare.CancelFile(req.TransferId, file.Id); err != nil {
		log.Printf("failed to cancel file in transfer %s: %s", req.TransferId, err)
		return fileshareError(pb.FileshareErrorCode_LIB_FAILURE), nil
	}

	return empty(), nil
}

func (s *Server) SetNotifications(ctx context.Context, in *pb.SetNotificationsRequest) (*pb.SetNotificationsResponse, error) {
	if in.Enable {
		switch s.eventManager.EnableNotifications(s.fileshare) {
		case ErrNotificationsAlreadyEnabled:
			return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_NOTHING_TO_DO}, nil
		case nil:
			return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_SET_SUCCESS}, nil
		default:
			return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_SET_FAILURE}, nil
		}
	} else {
		switch s.eventManager.DisableNotifications() {
		case ErrNotificationsAlreadyDisabled:
			return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_NOTHING_TO_DO}, nil
		case nil:
			return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_SET_SUCCESS}, nil
		default:
			return &pb.SetNotificationsResponse{Status: pb.SetNotificationsStatus_SET_FAILURE}, nil
		}
	}
}

func (s *Server) PurgeTransfersUntil(ctx context.Context, req *pb.PurgeTransfersUntilRequest) (*pb.Error, error) {
	resp, err := s.meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetValue() {
		return serviceError(pb.ServiceErrorCode_MESH_NOT_ENABLED), nil
	}

	err = s.fileshare.PurgeTransfersUntil(req.Until.AsTime())
	if err != nil {
		log.Printf("error while purging transfers: %s", err)
		return fileshareError(pb.FileshareErrorCode_PURGE_FAILURE), nil
	}

	return empty(), nil
}
