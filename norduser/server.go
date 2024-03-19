package norduser

import (
	"context"
	"log"

	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
)

type Server struct {
	pb.UnimplementedNorduserServer
}

func NewServer() *Server {
	return &Server{}
}

func (*Server) StartFileshare(context.Context, *pb.Empty) (*pb.Empty, error) {
	log.Println("start fileshare call")
	return &pb.Empty{}, nil
}

func (*Server) StopFileshare(context.Context, *pb.Empty) (*pb.Empty, error) {
	log.Println("stop fileshare call")
	return &pb.Empty{}, nil
}
