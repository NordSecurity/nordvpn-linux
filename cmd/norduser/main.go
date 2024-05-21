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
	"slices"
	"strconv"
	"time"

	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	daemonpb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
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
	for {
		if systray.IsAvailable() {
			break
		}
		<-time.After(10 * time.Second)
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
		return false, fmt.Errorf("connecting to main daemon: %w", err)
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

func setupLog() {
	log.SetOutput(os.Stdout)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configDirPath, err := internal.GetConfigDirPath(homeDir)

	if err == nil {
		if logFile, err := openLogFile(filepath.Join(configDirPath, internal.NorduserLogFile)); err == nil {
			log.SetOutput(logFile)
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
		}
	}
}

func waitForShutdown(stopChan <-chan norduser.StopRequest,
	fileshareManagementChan chan<- norduser.FileshareManagementMsg,
	fileshareShutdownChan <-chan interface{},
	grpcServer *grpc.Server,
	onShutdown func(bool)) {
	signals := internal.GetSignalChan()

	select {
	case sig := <-signals:
		log.Println("Received signal: ", sig)
	case stopRequest := <-stopChan:
		if stopRequest.DisableAutostart {
			onShutdown(stopRequest.DisableAutostart)
		}
	}

	grpcServer.GracefulStop()
	fileshareManagementChan <- norduser.Shutdown
	// shutdownChan will be closed once the shutdown operation is finished
	<-fileshareShutdownChan

	systray.Quit()
	// We need to give systray some time to clean up after quting. Otherwise, when the main app is restarted
	// two trays will be visible for a split second.
	<-time.After(500 * time.Millisecond)

	log.Println("Norduser process has stopped")
}

func startFileshare(uid uint32) (chan<- norduser.FileshareManagementMsg, <-chan interface{}) {
	fileshareManagementChan, fileshareShutdownChan := norduser.StartFileshareManagementLoop()
	if enable, err := shouldEnableFileshare(uid); err != nil {
		log.Println("Failed to determine if fileshare should be enabled on startup: ", err)
	} else if enable {
		fileshareManagementChan <- norduser.Start
	}

	return fileshareManagementChan, fileshareShutdownChan
}

func startSnap() {
	setupLog()
	group, err := user.LookupGroup(internal.NordvpnGroup)
	if err != nil {
		log.Println("Unable to retrieve nordvpn group: ", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	usr, err := user.Current()
	if err != nil {
		log.Println("Unable to retrieve current user: ", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	gids, err := usr.GroupIds()
	if err != nil {
		log.Println("Unable to retrieve group ids: ", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	if slices.Index(gids, group.Gid) == -1 {
		log.Println("User does not belong to the nordvpn group.")
		os.Exit(int(childprocess.CodeUserNotInGroup))
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

	socketPath := internal.GetNorduserSocketSnap(uid)

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println("Failed to remove old socket file: ", err)
	}

	listener, err := internal.ManualListener(socketPath, internal.PermUserRWGroupRWOthersRW)()
	if err != nil {
		log.Printf("Failed to open unix socket: %s", err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}
	limitedListener := netutil.LimitListener(listener, 100)

	fileshareManagementChan, fileshareShutdownChan := startFileshare(uint32(uid))

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshareManagementChan, stopChan)

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

	log.Println(internal.InfoPrefix, "Daemon has started")

	waitForShutdown(stopChan, fileshareManagementChan, fileshareShutdownChan, grpcServer,
		func(disable bool) {
			if !disable {
				return
			}

			if err := os.Remove(autostartFile); err != nil {
				log.Println("failed to remove autostart file: ", err)
			}
		})
}

func start() {
	listenerFunction := internal.SystemDListener

	setupLog()

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

	usr, err := user.Current()
	if err != nil {
		log.Println("Unable to retrieve current user:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Println("Failed to parse user id: ", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	fileshareManagementChan, fileshareShutdownChan := startFileshare(uint32(uid))

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshareManagementChan, stopChan)

	grpcServer := grpc.NewServer()
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Println(internal.ErrorPrefix+"Failed to start accept on grpc server: ", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray(stopChan)

	log.Println(internal.InfoPrefix, "Daemon has started")

	waitForShutdown(stopChan, fileshareManagementChan, fileshareShutdownChan, grpcServer, func(disable bool) {})
}

func main() {
	if snapconf.IsUnderSnap() {
		startSnap()
	} else {
		start()
	}
}
