package norduser

import (
	"context"
	"log"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
)

var errorCodeToProtobuffError = map[childprocess.StartupErrorCode]pb.StartFileshareStatus{
	childprocess.CodeAlreadyRunning:             pb.StartFileshareStatus_ALREADY_RUNNING,
	childprocess.CodeAlreadyRunningForOtherUser: pb.StartFileshareStatus_ALREADY_RUNNING_FOR_OTHER_USER,
	childprocess.CodeFailedToCreateUnixScoket:   pb.StartFileshareStatus_FAILED_TO_CREATE_UNIX_SOCKET,
	childprocess.CodeMeshnetNotEnabled:          pb.StartFileshareStatus_MESHNET_NOT_ENABLED,
	childprocess.CodeAddressAlreadyInUse:        pb.StartFileshareStatus_ADDRESS_ALREADY_IN_USE,
	childprocess.CodeFailedToEnable:             pb.StartFileshareStatus_FAILED_TO_ENABLE,
}

type StopRequest struct {
	DisableAutostart bool
}

type Server struct {
	pb.UnimplementedNorduserServer
	fileshareProcessManager childprocess.ChildProcessManager
	stopChan                chan<- StopRequest
}

func NewServer(fileshareProcessManager childprocess.ChildProcessManager, stopChan chan<- StopRequest) *Server {
	return &Server{
		fileshareProcessManager: fileshareProcessManager,
		stopChan:                stopChan,
	}
}

func (s *Server) Ping(context.Context, *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *Server) StartFileshare(context.Context, *pb.Empty) (*pb.StartFileshareResponse, error) {
	log.Println("Starting nordfileshare process")

	fileshareStatus := s.fileshareProcessManager.ProcessStatus()
	if fileshareStatus == childprocess.Running || fileshareStatus == childprocess.RunningForOtherUser {
		log.Println("Received start fileshare request but fileshare is already running, status is: ", fileshareStatus)
		return &pb.StartFileshareResponse{StartFileshareStatus: pb.StartFileshareStatus_SUCCESS}, nil
	}

	returnCode, err := s.fileshareProcessManager.StartProcess()

	if err != nil {
		log.Println("Failed to start fileshare process: ", err)
		return &pb.StartFileshareResponse{StartFileshareStatus: pb.StartFileshareStatus_FAILED_TO_ENABLE}, nil
	}

	returnStatus := pb.StartFileshareStatus_FAILED_TO_ENABLE
	if returnCode == 0 {
		returnStatus = pb.StartFileshareStatus_SUCCESS
		log.Println("Fileshare started")
	} else {
		log.Println("Failed to start fileshare, return code: ", returnCode)
		if s, ok := errorCodeToProtobuffError[returnCode]; ok {
			returnStatus = s
		}
	}

	return &pb.StartFileshareResponse{StartFileshareStatus: returnStatus}, nil
}

func (s *Server) StopFileshare(context.Context, *pb.Empty) (*pb.StopFileshareResponse, error) {
	log.Println(internal.InfoPrefix + "Stopping nordfileshare process")

	fileshareStatus := s.fileshareProcessManager.ProcessStatus()
	if fileshareStatus == childprocess.NotRunning {
		log.Println("Received stop fileshare request but fileshare is already not running, status is: ", fileshareStatus)
		return &pb.StopFileshareResponse{Success: true}, nil
	}

	if err := s.fileshareProcessManager.StopProcess(false); err != nil {
		log.Println("Failed to stop fileshare process: ", err.Error())
		return &pb.StopFileshareResponse{Success: false}, nil
	}

	return &pb.StopFileshareResponse{Success: true}, nil
}

func (s *Server) Stop(_ context.Context, req *pb.StopNorduserRequest) (*pb.Empty, error) {
	select {
	case s.stopChan <- StopRequest{DisableAutostart: req.Disable}:
	default:
	}
	return &pb.Empty{}, nil
}
