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
	"syscall"
	"time"

	"github.com/NordSecurity/systray"
	"golang.org/x/net/netutil"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/clientid"
	daemonpb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser"
	"github.com/NordSecurity/nordvpn-linux/norduser/pb"
	"github.com/NordSecurity/nordvpn-linux/norduser/process"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/tray"
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
		"\nExec=env BAMF_DESKTOP_FILE_HINT=/var/lib/snapd/desktop/applications/nordvpn_nordvpn.desktop /snap/bin/nordvpn user" +
		"\nTerminal=false" +
		"\nType=Application" +
		"\nX-SnapInstanceName=nordvpn" +
		"\nX-GNOME-Autostart-enabled=true" +
		"\nX-GNOME-Autostart-Delay=20" +
		"\nX-KDE-autostart-after=panel" +
		"\nX-MATE-Autostart-Delay=20" +
		"\nComment=This is an autostart for NordVPN user daemon" +
		"\nCategories=Utility;"

	dataDir := os.Getenv(snapconf.EnvSnapUserData)
	p := path.Join(dataDir, ".config", "autostart", "nordvpn.desktop")
	if err := internal.EnsureDir(p); err != nil {
		return "", fmt.Errorf("ensuring path: %w", err)
	}

	return p, internal.FileWrite(p, []byte(autostartDesktopFileContents), internal.PermUserRW)
}

func startTray(quitChan chan<- norduser.StopRequest) {
	daemonURL := fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	cliendIDMetadataInterceptor := clientid.NewInsertClientIDInterceptor(daemonpb.ClientID_TRAY)

	log.Printf("%s Tray: dialing daemon at %q", internal.InfoPrefix, daemonURL)
	conn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(cliendIDMetadataInterceptor.SetMetadataUnaryInterceptor),
		grpc.WithStreamInterceptor(cliendIDMetadataInterceptor.SetMetadataStreamInterceptor),
	)

	var client daemonpb.DaemonClient
	if err == nil {
		client = daemonpb.NewDaemonClient(conn)
		log.Printf("%s Tray: connected to daemon", internal.InfoPrefix)
	} else {
		log.Println(internal.ErrorPrefix, "Error connecting to the NordVPN daemon:", err)
		return
	}
	ReportTelemetry(conn, ReportOnStart, false)

	log.Printf("%s Tray: dialing fileshare daemon at %q", internal.InfoPrefix, fileshare_process.FileshareURL)
	fileshareConn, err := grpc.Dial(
		fileshare_process.FileshareURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	var fileshareClient filesharepb.FileshareClient
	if err == nil {
		fileshareClient = filesharepb.NewFileshareClient(fileshareConn)
		log.Printf("%s Tray: connected to fileshare daemon", internal.InfoPrefix)
	} else {
		log.Println(internal.ErrorPrefix, "Error connecting to the NordVPN fileshare daemon:", err)
		return
	}

	ti := tray.NewTrayInstance(client, fileshareClient, quitChan)
	log.Printf("%s Tray: starting tray instance", internal.InfoPrefix)
	ti.Start()

	onExit := func() {
		log.Println(internal.InfoPrefix, "Exiting systray")
		ReportTelemetry(conn, ReportOnExit, true)
		ti.OnExit()
	}

	onReady := func() {
		log.Println(internal.InfoPrefix, "Starting systray")
		ti.OnReady()
	}

	trayStatus := ti.WaitInitialTrayStatus()
	log.Printf("%s Tray: initial status %v", internal.InfoPrefix, trayStatus)
	if trayStatus == tray.Enabled {
		attempt := 0
		for {
			attempt++
			if systray.IsAvailable() {
				log.Printf("%s Tray: backend available, running systray (attempt %d)", internal.InfoPrefix, attempt)
				systray.Run(onReady, onExit)
				break
			}
			log.Printf("%s Tray: backend not yet available; retrying in 10s (attempt %d)", internal.InfoPrefix, attempt)
			<-time.After(10 * time.Second)
		}
	} else {
		log.Printf("%s Tray: disabled by initial status; systray will not start", internal.InfoPrefix)
	}
}

func shouldEnableFileshare(uid uint32) (bool, error) {
	daemonURL := fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
	log.Printf("%s Fileshare: checking mesh status via %q for uid=%d", internal.InfoPrefix, daemonURL, uid)

	grpcConn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return false, fmt.Errorf("connecting to main daemon: %w", err)
	}

	defer func() {
		if err := grpcConn.Close(); err != nil {
			log.Println(internal.ErrorPrefix, "Failed to close grpc connection")
		}
	}()

	meshClient := meshpb.NewMeshnetClient(grpcConn)

	resp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil {
		return false, fmt.Errorf("running is mesh enabled grpc: %w", err)
	}

	meshStatus := resp.GetStatus()
	enabled := meshStatus.GetUid() == uid && meshStatus.GetValue()
	log.Printf("%s Fileshare: mesh enabled=%t (meshUid=%d reqUid=%d)", internal.InfoPrefix, enabled, meshStatus.GetUid(), uid)

	return enabled, nil
}

func setupLog() {
	log.SetOutput(os.Stdout)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	cacheDirPath, err := internal.GetCacheDirPath(homeDir)

	if err == nil {
		if logFile, err := openLogFile(filepath.Join(cacheDirPath, internal.NorduserdLogFileName)); err == nil {
			log.SetOutput(logFile)
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
			// Startup banner
			pid := os.Getpid()
			euid := os.Geteuid()
			log.Printf("%s ==== norduserd start pid=%d euid=%d time=%s ====", internal.InfoPrefix, pid, euid, time.Now().Format(time.RFC3339Nano))
			log.Printf("%s Env: XDG_CURRENT_DESKTOP=%q DESKTOP_SESSION=%q WAYLAND_DISPLAY=%q DISPLAY=%q", internal.InfoPrefix,
				os.Getenv("XDG_CURRENT_DESKTOP"), os.Getenv("DESKTOP_SESSION"), os.Getenv("WAYLAND_DISPLAY"), os.Getenv("DISPLAY"))
			log.Printf("%s SNAP env: SNAP=%q SNAP_NAME=%q SNAP_INSTANCE_NAME=%q SNAP_USER_DATA=%q", internal.InfoPrefix,
				os.Getenv("SNAP"), os.Getenv("SNAP_NAME"), os.Getenv("SNAP_INSTANCE_NAME"), os.Getenv(snapconf.EnvSnapUserData))
		}
	}
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
		log.Println(internal.InfoPrefix, "Received signal:", sig)
		if sig == unix.SIGHUP {
			restart = true
		}
	case stopRequest := <-stopChan:
		log.Printf("%s Stop request received: restart=%t disableAutostart=%t", internal.InfoPrefix, stopRequest.Restart, stopRequest.DisableAutostart)
		if stopRequest.DisableAutostart {
			onShutdown(stopRequest.DisableAutostart)
		}
		if stopRequest.Restart {
			restart = true
		}
	case <-logoutChan:
		log.Println(internal.InfoPrefix, "User has logged out")
	}

	log.Println(internal.InfoPrefix, "Stopping gRPC server…")
	grpcServer.Stop()
	log.Println(internal.InfoPrefix, "Signalling fileshare to shutdown…")
	fileshareManagementChan <- norduser.Shutdown
	// shutdownChan will be closed once the shutdown operation is finished
	<-fileshareShutdownChan
	log.Println(internal.InfoPrefix, "Fileshare shutdown completed")

	log.Println(internal.InfoPrefix, "Stopping systray…")
	systray.Quit()
	// We need to give systray some time to clean up after quiting. Otherwise, when the main app is restarted
	// two trays will be visible for a split second.
	<-time.After(500 * time.Millisecond)

	log.Println(internal.InfoPrefix, "Norduser daemon has stopped")

	if restart {
		log.Println(internal.InfoPrefix, "Norduser daemon restarting")
		execpath, err := os.Executable()
		if err == nil {
			err = syscall.Exec(execpath, os.Args, os.Environ())
			if err != nil {
				log.Println(internal.InfoPrefix, "Norduser daemon restart error:", err)
			}
		} else {
			log.Println(internal.ErrorPrefix, "Failed to resolve executable path for restart:", err)
		}
	}
}

func startFileshare(uid uint32) (chan<- norduser.FileshareManagementMsg, <-chan interface{}) {
	log.Printf("%s Fileshare: start management loop for uid=%d", internal.InfoPrefix, uid)
	fileshareManagementChan, fileshareShutdownChan := norduser.StartFileshareManagementLoop()
	if enable, err := shouldEnableFileshare(uid); err != nil {
		log.Println(internal.ErrorPrefix, "Failed to determine if fileshare should be enabled on startup:", err)
	} else if enable {
		log.Println(internal.InfoPrefix, "Fileshare: enabling on startup")
		fileshareManagementChan <- norduser.Start
	} else {
		log.Println(internal.InfoPrefix, "Fileshare: not enabled for this user/session")
	}

	return fileshareManagementChan, fileshareShutdownChan
}

func startSnap() {
	setupLog()
	log.Println(internal.InfoPrefix, "Starting in SNAP environment")
	group, err := user.LookupGroup(internal.NordvpnGroup)
	if err != nil {
		log.Println(internal.ErrorPrefix, "Unable to retrieve nordvpn group:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	usr, err := user.Current()
	if err != nil {
		log.Println(internal.ErrorPrefix, "Unable to retrieve current user:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	gids, err := usr.GroupIds()
	if err != nil {
		log.Println(internal.ErrorPrefix, "Unable to retrieve group ids:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	if slices.Index(gids, group.Gid) == -1 {
		log.Println(internal.ErrorPrefix, "User does not belong to the nordvpn group")
		os.Exit(int(childprocess.CodeUserNotInGroup))
	}

	autostartFile, err := addAutostart()
	if err != nil {
		log.Println(internal.ErrorPrefix, "Failed to add autostart:", err)
	} else {
		log.Printf("%s Autostart desktop file ensured at %q", internal.InfoPrefix, autostartFile)
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Printf("%s Invalid unix user id, failed to convert from string: %s", internal.ErrorPrefix, usr.Uid)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	logoutChan := make(chan interface{})
	go func() {
		log.Println(internal.InfoPrefix, "Starting logout monitor…")
		if err := norduser.WaitForLogout(usr.Username, logoutChan); err != nil {
			log.Println(internal.ErrorPrefix, "failed to start logout monitor:", err)
		}
	}()

	processStatus := process.NewNorduserGRPCProcessManager(uint32(uid)).ProcessStatus()
	log.Printf("%s Process status: %v", internal.InfoPrefix, processStatus)
	if processStatus == childprocess.Running {
		os.Exit(int(childprocess.CodeAlreadyRunning))
	}

	socketPath := internal.GetNorduserSocketSnap(uid)
	log.Printf("%s Opening UNIX socket at %q", internal.InfoPrefix, socketPath)

	if err := os.Remove(socketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println(internal.ErrorPrefix, "Failed to remove old socket file:", err)
	}

	listener, err := internal.ManualListener(socketPath, internal.PermUserRWGroupRWOthersRW)()
	if err != nil {
		log.Printf("%s Failed to open unix socket: %s", internal.ErrorPrefix, err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}
	limitedListener := netutil.LimitListener(listener, 100)

	fileshareManagementChan, fileshareShutdownChan := startFileshare(uint32(uid))

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshareManagementChan, stopChan)

	log.Println(internal.InfoPrefix, "Starting gRPC server (snap)")
	grpcServer := grpc.NewServer(
		grpc.Creds(internal.NewUnixSocketCredentials(internal.NewFileshareAuthenticator(uint32(uid)))))
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(limitedListener); err != nil {
			log.Println(internal.ErrorPrefix, "failed to start accept on grpc server:", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray(stopChan)

	log.Println(internal.InfoPrefix, "Daemon has started")

	waitForShutdown(stopChan, fileshareManagementChan, fileshareShutdownChan, logoutChan, grpcServer,
		func(disable bool) {
			if !disable {
				return
			}

			if err := os.Remove(autostartFile); err != nil {
				log.Println(internal.ErrorPrefix, "Failed to remove autostart file:", err)
			} else {
				log.Printf("%s Autostart file %q removed", internal.InfoPrefix, autostartFile)
			}
		})
}

func start() {
	listenerFunction := internal.SystemDListener

	setupLog()
	log.Println(internal.InfoPrefix, "Starting in non-snap environment")

	connURL := internal.GetNorduserSocketFork(os.Geteuid())
	log.Printf("%s Fork socket path: %q", internal.InfoPrefix, connURL)
	if err := os.Remove(connURL); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println(internal.ErrorPrefix, "Failed to remove old socket file:", err)
	}
	listenerFunction = internal.ManualListener(connURL, internal.PermUserRWX)

	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf("%s Error on listening to UNIX domain socket: %s\n", internal.ErrorPrefix, err)
	}
	listener = netutil.LimitListener(listener, 100)

	usr, err := user.Current()
	if err != nil {
		log.Println(internal.ErrorPrefix, "Unable to retrieve current user:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Println(internal.ErrorPrefix, "Failed to parse user id:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	fileshareManagementChan, fileshareShutdownChan := startFileshare(uint32(uid))

	stopChan := make(chan norduser.StopRequest)
	server := norduser.NewServer(fileshareManagementChan, stopChan)

	log.Println(internal.InfoPrefix, "Starting gRPC server (non-snap)")
	grpcServer := grpc.NewServer()
	pb.RegisterNorduserServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Println(internal.ErrorPrefix, "Failed to start accept on grpc server:", err)
			os.Exit(int(childprocess.CodeFailedToEnable))
		}
	}()

	go startTray(stopChan)

	log.Println(internal.InfoPrefix, "Norduser daemon has started")

	// logoutChan is not needed in non-snap environment, as startup/shutdown on login/logout is managed by the main daemon
	waitForShutdown(stopChan, fileshareManagementChan, fileshareShutdownChan, make(<-chan interface{}),
		grpcServer,
		func(disable bool) {})
}

func main() {
	if snapconf.IsUnderSnap() {
		startSnap()
	} else {
		start()
	}
}
