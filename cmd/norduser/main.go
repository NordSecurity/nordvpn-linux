package main

import (
	"log"
	"os"
	"strconv"

	"google.golang.org/grpc"

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
		log.Fatalf("Error on listening to UNIX domain socket: %s\n", err)
	}

	server := norduser.NewServer()

	grpcServer := grpc.NewServer()
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalln("failed to start accept on grpc server: ", err)
		}
	}()

	internal.WaitSignal()

	grpcServer.GracefulStop()
}
