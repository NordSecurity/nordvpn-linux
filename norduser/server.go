package norduser

import (
	"context"

	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
)

type StopRequest struct {
	DisableAutostart bool
	Restart          bool
}

type Server struct {
	pb.UnimplementedNorduserServer
	fileshareManagementChan chan<- FileshareManagementMsg
	stopChan                chan<- StopRequest
}

func NewServer(fileshareManagementChan chan<- FileshareManagementMsg, stopChan chan<- StopRequest) *Server {
	return &Server{
		fileshareManagementChan: fileshareManagementChan,
		stopChan:                stopChan,
	}
}

func (s *Server) Ping(context.Context, *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func (s *Server) StartFileshare(context.Context, *pb.Empty) (*pb.Empty, error) {
	s.fileshareManagementChan <- Start

	return &pb.Empty{}, nil
}

func (s *Server) StopFileshare(context.Context, *pb.Empty) (*pb.Empty, error) {
	s.fileshareManagementChan <- Stop

	return &pb.Empty{}, nil
}

func (s *Server) Stop(_ context.Context, req *pb.StopNorduserRequest) (*pb.Empty, error) {
	select {
	case s.stopChan <- StopRequest{DisableAutostart: req.Disable, Restart: req.Restart}:
	default:
	}
	return &pb.Empty{}, nil
}
