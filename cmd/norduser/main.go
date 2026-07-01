package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path"
	"slices"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/net/netutil"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/tray"
)

// Value set when building the application
var Environment = ""

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

func startTray() {
	log.Info("Starting systray")
	tray.Start()
	log.Info("Exiting systray")
}

func shouldEnableFileshare(uid uint32) (bool, error) {
	daemonURL := fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)

	//nolint:staticcheck
	grpcConn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false, fmt.Errorf("connecting to main daemon: %w", err)
	}

	defer func() {
		if err := grpcConn.Close(); err != nil {
			log.Error("Failed to close grpc connection")
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

func waitForShutdown(stopChan <-chan norduser.StopRequest,
	fileshareManagementChan chan<- norduser.FileshareManagementMsg,
	fileshareShutdownChan <-chan interface{},
	logoutChan <-chan interface{},
	grpcServer *grpc.Server,
	onShutdown func(bool),
) {
	restart := false
	signals := internal.GetSignalChan()

	select {
	case sig := <-signals:
		log.Info("Received signal:", sig)
		if sig == unix.SIGHUP {
			restart = true
		}
	case stopRequest := <-stopChan:
		if stopRequest.DisableAutostart {
			onShutdown(stopRequest.DisableAutostart)
		}
		if stopRequest.Restart {
			restart = true
		}
	case <-logoutChan:
		log.Info("User has logged out")
	}

	grpcServer.Stop()
	fileshareManagementChan <- norduser.Shutdown
	// shutdownChan will be closed once the shutdown operation is finished
	<-fileshareShutdownChan

	tray.Stop()
	<-time.After(500 * time.Millisecond)

	log.Info("Norduser daemon has stopped")

	if restart {
		log.Info("Norduser daemon restarting")
		execpath, err := os.Executable()
		if err == nil {
			// #nosec G204 - restart, reusing arguments is acceptable
			err = syscall.Exec(execpath, os.Args, os.Environ())
			if err != nil {
				log.Error("Norduser daemon restart error:", err)
			}
		}
	}
}

func startFileshare(uid uint32) (chan<- norduser.FileshareManagementMsg, <-chan interface{}) {
	fileshareManagementChan, fileshareShutdownChan := norduser.StartFileshareManagementLoop()
	if enable, err := shouldEnableFileshare(uid); err != nil {
		log.Error("Failed to determine if fileshare should be enabled on startup:", err)
	} else if enable {
		fileshareManagementChan <- norduser.Start
	}

	return fileshareManagementChan, fileshareShutdownChan
}

func startSnap() {
	group, err := user.LookupGroup(internal.NordvpnGroup)
	if err != nil {
		log.Error("Unable to retrieve nordvpn group:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	usr, err := user.Current()
	if err != nil {
		log.Error("Unable to retrieve current user:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	gids, err := usr.GroupIds()
	if err != nil {
		log.Error("Unable to retrieve group ids:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	if slices.Index(gids, group.Gid) == -1 {
		log.Error("User does not belong to the nordvpn group")
		os.Exit(int(childprocess.CodeUserNotInGroup))
	}

	// Always use real home dir here regardless of `$HOME` value
	autostartFile, err := addAutostart()
	if err != nil {
		log.Error("Failed to add autostart:", err)
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Errorf("Invalid unix user id, failed to convert from string: %s", usr.Uid)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	logoutChan := make(chan interface{})
	go func() {
		if err := norduser.WaitForLogout(usr.Username, logoutChan); err != nil {
			log.Error("failed to start logout monitor:", err)
		}
	}()

	// #nosec G115
	processStatus := process.NewNorduserGRPCProcessManager(uint32(uid)).ProcessStatus()
	if processStatus == childprocess.Running {
		os.Exit(int(childprocess.CodeAlreadyRunning))
	}

	socketPath := internal.GetNorduserSocketSnap(uid)

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Error("Failed to remove old socket file:", err)
	}

	listener, err := internal.ManualListener(socketPath, internal.PermUserRWGroupRWOthersRW)()
	if err != nil {
		log.Errorf("Failed to open unix socket: %s", err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}
	limitedListener := netutil.LimitListener(listener, 100)

	// #nosec G115
	fileshareManagementChan, fileshareShutdownChan := startFileshare(uint32(uid))

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshareManagementChan, stopChan)

	// #nosec G115
	grpcServer := grpc.NewServer(
		grpc.Creds(internal.NewUnixSocketCredentials(internal.NewFileshareAuthenticator(uint32(uid)))))
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(limitedListener); err != nil {
			log.Error("failed to start accept on grpc server:", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray()

	log.Info("Daemon has started")

	waitForShutdown(stopChan, fileshareManagementChan, fileshareShutdownChan, logoutChan, grpcServer,
		func(disable bool) {
			if !disable {
				return
			}

			if err := os.Remove(autostartFile); err != nil {
				log.Error("Failed to remove autostart file:", err)
			}
		})
}

func start() {
	connURL := internal.GetNorduserSocketFork(os.Geteuid())
	if err := os.Remove(connURL); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Error("Failed to remove old socket file:", err)
	}
	listenerFunction := internal.ManualListener(connURL, internal.PermUserRWX)

	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf("Error on listening to UNIX domain socket: %s", err)
	}
	listener = netutil.LimitListener(listener, 100)

	usr, err := user.Current()
	if err != nil {
		log.Error("Unable to retrieve current user:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Error("Failed to parse user id:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	// #nosec G115
	fileshareManagementChan, fileshareShutdownChan := startFileshare(uint32(uid))

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshareManagementChan, stopChan)

	grpcServer := grpc.NewServer()
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("Failed to start accept on grpc server:", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray()

	log.Info("Norduser daemon has started")

	// logoutChan is not needed in non-snap environment, as startup/shutdown on login/logout is managed by the main daemon
	waitForShutdown(stopChan, fileshareManagementChan, fileshareShutdownChan, make(<-chan interface{}),
		grpcServer,
		func(disable bool) {})
}

func main() {
	stopLevelWatcher := log.SetupLogger(
		internal.UserLogOutput(internal.NorduserdLogFileName, internal.MaxUserLogSizeMB),
		internal.LogLevelFile,
		log.DefaultLevel(),
	)
	defer stopLevelWatcher()

	if snapconf.IsUnderSnap() {
		startSnap()
	} else {
		start()
	}
}
