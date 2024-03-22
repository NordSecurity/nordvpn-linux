package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"golang.org/x/net/netutil"
	"google.golang.org/grpc"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

var (
	ConnURL = internal.GetNorduserdSocket(os.Geteuid())
	PidFile = internal.GetFilesharedPid(os.Getuid())
)

func openLogFile(path string) (*os.File, error) {
	// #nosec path is constant
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func startSnap() {
	usr, err := user.Current()
	if err != nil {
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	configDirPath, err := internal.GetConfigDirPath(usr.HomeDir)
	if err == nil {
		if logFile, err := openLogFile(filepath.Join(configDirPath, internal.NorduserLogFile)); err == nil {
			log.SetOutput(logFile)
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
		}
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Printf("Invalid unix user id, failed to convert from string: %s", usr.Uid)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	processStatus := process.NewNorduserGRPCProcessManager(uint32(uid)).ProcessStatus()
	if processStatus == childprocess.Running {
		os.Exit(int(childprocess.CodeAlreadyRunning))
	}

	socketPath := internal.GetNorduserSocketSnap(uint32(uid))

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println("Failed to remove old socket file: ", err)
	}

	listener, err := internal.ManualListener(socketPath, internal.PermUserRWGroupRWOthersRW, PidFile)()
	if err != nil {
		log.Printf("Failed to open unix socket: %s", err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}

	limitedListener := netutil.LimitListener(listener, 100)

	stopChan := make(chan interface{})

	server := norduser.NewServer(fileshare_process.NewFileshareGRPCProcessManager(), make(chan<- interface{}))
	grpcServer := grpc.NewServer(
		grpc.Creds(internal.NewUnixSocketCredentials(internal.NewFileshareAuthenticator(uint32(uid)))))

	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(limitedListener); err != nil {
			log.Fatalln("failed to start accept on grpc server: ", err)
		}
	}()

	signals := internal.GetSignalChan()

	log.Println(internal.InfoPrefix, "Daemon has started")
	select {
	case sig := <-signals:
		log.Println("Received signal: ", sig)
	case <-stopChan:
	}

	if _, err := server.StopFileshare(context.Background(), &pb.Empty{}); err != nil {
		log.Println(internal.ErrorPrefix+"Failed to stop fileshare: ", err)
	}

	grpcServer.GracefulStop()
	log.Println("Daemon has stopped")
}

func start() {
	var listenerFunction = internal.SystemDListener
	if os.Getenv(internal.ListenPID) != strconv.Itoa(os.Getpid()) {
		listenerFunction = internal.ManualListener(ConnURL, internal.PermUserRWX, PidFile)
	}

	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf(internal.ErrorPrefix+"Error on listening to UNIX domain socket: %s\n", err)
	}
	listener = netutil.LimitListener(listener, 100)

	server := norduser.NewServer(fileshare_process.NewFileshareGRPCProcessManager(), make(chan<- interface{}))

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

func main() {
	if snapconf.IsUnderSnap() {
		startSnap()
	} else {
		start()
	}
}
