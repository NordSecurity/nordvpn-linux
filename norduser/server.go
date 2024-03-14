package norduser

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
)

type Server struct {
	pb.UnimplementedNorduserServer
	fileshareProcessManager fileshare_process.GRPCFileshareProcess
	fileshareRunning        bool
}

func NewServer(fileshareProcessManager fileshare_process.GRPCFileshareProcess) *Server {
	return &Server{
		fileshareProcessManager: fileshareProcessManager,
	}
}

func (s *Server) StartFileshare(context.Context, *pb.Empty) (*pb.StartFileshareResponse, error) {
	log.Println(internal.InfoPrefix + "Starting nordfileshare process")
	status := s.fileshareProcessManager.StartProcess()

	if status == pb.StartFileshareStatus_SUCCESS {
		s.fileshareRunning = true
	}

	return &pb.StartFileshareResponse{StartFileshareStatus: status}, nil
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
