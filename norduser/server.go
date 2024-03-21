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

type Server struct {
	pb.UnimplementedNorduserServer
	fileshareProcessManager childprocess.ChildProcessManager
	fileshareRunning        bool
	stopChan                chan<- interface{}
}

func NewServer(fileshareProcessManager childprocess.ChildProcessManager, stopChan chan<- interface{}) *Server {
	return &Server{
		fileshareProcessManager: fileshareProcessManager,
		stopChan:                stopChan,
	}
}

func (s *Server) Ping(context.Context, *pb.Empty) (*pb.Empty, error) {
	log.Println("ping request")
	return &pb.Empty{}, nil
}

func (s *Server) StartFileshare(context.Context, *pb.Empty) (*pb.StartFileshareResponse, error) {
	log.Println("Starting nordfileshare process")
	returnCode, err := s.fileshareProcessManager.StartProcess()

	if err != nil {
		log.Println("Failed to start fileshare process: ", err)
		return &pb.StartFileshareResponse{StartFileshareStatus: pb.StartFileshareStatus_FAILED_TO_ENABLE}, nil
	}

	returnStatus := pb.StartFileshareStatus_FAILED_TO_ENABLE
	if returnCode == 0 {
		s.fileshareRunning = true
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
	if s.fileshareRunning {
		if err := s.fileshareProcessManager.StopProcess(); err != nil {
			log.Println(internal.ErrorPrefix+"Failed to stop fileshare process: ", err.Error())
			return &pb.StopFileshareResponse{Success: false}, nil
		}

		s.fileshareRunning = false
	}
	return &pb.StopFileshareResponse{Success: true}, nil
}

func (s *Server) Stop(context.Context, *pb.Empty) (*pb.Empty, error) {
	s.stopChan <- struct{}{}
	return &pb.Empty{}, nil
}
