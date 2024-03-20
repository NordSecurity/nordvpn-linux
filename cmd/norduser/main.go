package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"golang.org/x/net/netutil"
	"google.golang.org/grpc"

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
)

var ConnURL = internal.GetNorduserdSocket(os.Geteuid())

func main() {
	var listenerFunction = internal.SystemDListener
	if os.Getenv(internal.ListenPID) != strconv.Itoa(os.Getpid()) {
		listenerFunction = internal.ManualListener(ConnURL, internal.PermUserRWX)
	}

	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf(internal.ErrorPrefix+"Error on listening to UNIX domain socket: %s\n", err)
	}
	listener = netutil.LimitListener(listener, 100)

	server := norduser.NewServer(fileshare_process.GRPCFileshareProcess{})

	grpcServer := grpc.NewServer()
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalln(internal.ErrorPrefix+"Failed to start accept on grpc server: ", err)
		}
	}()

	internal.WaitSignal()

	if _, err := server.StopFileshare(context.Background(), &pb.Empty{}); err != nil {
		log.Println(internal.ErrorPrefix+"Failed to stop fileshare: ", err)
	}

	grpcServer.GracefulStop()
}
