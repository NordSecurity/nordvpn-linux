package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	daemonpb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/tray"
	"github.com/NordSecurity/systray"
)

func openLogFile(path string) (*os.File, error) {
	// #nosec path is constant
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func addAutostart() (string, error) {
	autostartDesktopFileContents := "[Desktop Entry]" +
		"\nName=NordVPN" +
		"\nExec=nordvpn user" +
		"\nTerminal=false" +
		"\nType=Application" +
		"\nX-GNOME-Autostart-enabled=true" +
		"\nX-GNOME-Autostart-Delay=10" +
		"\nX-KDE-autostart-after=panel" +
		"\nX-MATE-Autostart-Delay=10" +
		"\nComment=This is an autostart for NordVPN user daemon" +
		"\nCategories=Utility;"

	dataDir := os.Getenv(snapconf.EnvSnapUserData)
	path := path.Join(dataDir, ".config", "autostart", "nordvpn.desktop")
	if err := internal.EnsureDir(path); err != nil {
		return "", fmt.Errorf("ensuring path: %w", err)
	}

	return path, internal.FileWrite(path, []byte(autostartDesktopFileContents), internal.PermUserRW)
}

func startTray(quitChan chan<- norduser.StopRequest) {
	try := 0
	// Retry checking systray availability, as it might not be availalble on startup.
	for {
		if systray.IsAvailable() {
			break
		}

		if try == 5 {
			log.Println("Session tray not available, exiting")
			return
		}

		try++
		<-time.After(10)
	}

	daemonURL := fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	conn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	var client daemonpb.DaemonClient
	if err == nil {
		client = daemonpb.NewDaemonClient(conn)
	} else {
		log.Println("Error connecting to the NordVPN daemon: ", err)
		return
	}

	onExit := func() {
		now := time.Now()
		log.Println("Tray exit at", now.String())
	}

	ti := tray.NewTrayInstance(client, quitChan)
	onReady := func() {
		log.Println("Tray ready")
		tray.OnReady(ti)
	}

	systray.Run(onReady, onExit)
}

func shouldEnableFileshare(uid uint32) (bool, error) {
	daemonURL := fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)

	grpcConn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false, fmt.Errorf("can't connect to daemon: %w", err)
	}

	defer func() {
		if err := grpcConn.Close(); err != nil {
			log.Println("failed to close grpc connection")
		}
	}()

	meshClient := meshpb.NewMeshnetClient(grpcConn)

	resp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil {
		return false, fmt.Errorf("running is mesh enabled grpc: %w", err)
	}

	meshStatus := resp.GetStatus()

	return meshStatus.GetUid() == uid && meshStatus.GetValue(), nil
}

func startSnap() {
	usr, err := user.Current()
	if err != nil {
		log.Println("Unable to retrieve current user:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("Unable to retrieve user home directory:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	configDirPath, err := internal.GetConfigDirPath(homeDir)

	if err == nil {
		if logFile, err := openLogFile(filepath.Join(configDirPath, internal.NorduserLogFile)); err == nil {
			log.SetOutput(logFile)
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
		}
	}

	// Always use real home dir here regardless of `$HOME` value
	autostartFile, err := addAutostart()
	if err != nil {
		log.Println("Failed to add autostart: ", err)
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

	listener, err := internal.ManualListener(socketPath, internal.PermUserRWGroupRWOthersRW)()
	if err != nil {
		log.Printf("Failed to open unix socket: %s", err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}

	limitedListener := netutil.LimitListener(listener, 100)

	stopChan := make(chan norduser.StopRequest)
	fileshareProcessManager := fileshare_process.NewFileshareGRPCProcessManager()
	if enable, err := shouldEnableFileshare(uint32(uid)); err != nil {
		log.Println("Failed to determine if fileshare should be enabled on startup: ", err)
	} else if enable {
		if startupCode, err := fileshareProcessManager.StartProcess(); err != nil {
			log.Println("Failed to enable fileshare at startup: ", err)
		} else {
			log.Println("Fileshare enable status at startup: ", startupCode)
		}
	}

	server := norduser.NewServer(fileshareProcessManager, stopChan)

	grpcServer := grpc.NewServer(
		grpc.Creds(internal.NewUnixSocketCredentials(internal.NewFileshareAuthenticator(uint32(uid)))))
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(limitedListener); err != nil {
			log.Println("failed to start accept on grpc server: ", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray(stopChan)

	signals := internal.GetSignalChan()

	log.Println(internal.InfoPrefix, "Daemon has started")
	select {
	case sig := <-signals:
		log.Println("Received signal: ", sig)
	case stopRequest := <-stopChan:
		if stopRequest.DisableAutostart {
			if err := os.Remove(autostartFile); err != nil {
				log.Println("Failed to remove autostart file: ", err)
			}
		}
	}

	if _, err := server.StopFileshare(context.Background(), &pb.Empty{}); err != nil {
		log.Println(internal.ErrorPrefix+"Failed to stop fileshare: ", err)
	}

	grpcServer.GracefulStop()
	log.Println("Norduser process has stopped")
}

func start() {
	listenerFunction := internal.SystemDListener

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println("failed to find home dir: ", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	configDirPath, err := internal.GetConfigDirPath(homeDir)
	if err == nil {
		if logFile, err := openLogFile(filepath.Join(configDirPath, internal.NorduserLogFile)); err == nil {
			log.SetOutput(logFile)
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
		}
	}

	connURL := internal.GetNorduserSocketFork(os.Geteuid())
	if err := os.Remove(connURL); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println("Failed to remove old socket file: ", err)
	}
	listenerFunction = internal.ManualListener(connURL, internal.PermUserRWX)

	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf(internal.ErrorPrefix+"Error on listening to UNIX domain socket: %s\n", err)
	}
	listener = netutil.LimitListener(listener, 100)

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshare_process.NewFileshareGRPCProcessManager(), stopChan)

	grpcServer := grpc.NewServer()
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Println(internal.ErrorPrefix+"Failed to start accept on grpc server: ", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray(stopChan)

	signals := internal.GetSignalChan()

	log.Println(internal.InfoPrefix, "Daemon has started")
	select {
	case sig := <-signals:
		log.Println("Received signal: ", sig)
	case <-stopChan:
		log.Println("Received stop request")
	}

	log.Println("Stopping fileshare")

	if _, err := server.StopFileshare(context.Background(), &pb.Empty{}); err != nil {
		log.Println(internal.ErrorPrefix+"Failed to stop fileshare: ", err)
	}

	grpcServer.GracefulStop()

	log.Println("Norduser process has stopped")
}

func main() {
	if snapconf.IsUnderSnap() {
		startSnap()
	} else {
		start()
	}
}
