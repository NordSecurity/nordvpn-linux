// NordVPN fileshare daemon.
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"net/netip"
	"os"
	"path/filepath"
	"runtime/debug"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	daemonpb "github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx"
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_startup"
	"github.com/NordSecurity/nordvpn-linux/fileshare/storage"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"golang.org/x/net/netutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// Value set when building the application
	Environment = ""
	daemonURL   = fmt.Sprintf("%s://%s", internal.Proto, internal.DaemonSocket)
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic:", r)
			if internal.IsDevEnv(Environment) {
				log.Error(string(debug.Stack()))
			}
			panic(r)
		}
	}()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	logFile := internal.UserLogOutput(internal.FileshareLogFileName)
	redirectNativeOutput(logFile)
	stopLevelWatcher := log.SetupLogger(
		logFile,
		internal.LogLevelFile,
		log.DefaultLevel(internal.IsDevEnv(Environment)),
	)
	defer stopLevelWatcher()

	processStatus := fileshare_process.NewFileshareGRPCProcessManager().ProcessStatus()
	switch processStatus {
	case childprocess.Running:
		os.Exit(int(childprocess.CodeAlreadyRunning))
	case childprocess.RunningForOtherUser:
		log.Error("Cannot start fileshare daemon, it is already running for another user.")
		os.Exit(int(childprocess.CodeAlreadyRunningForOtherUser))
	case childprocess.NotRunning:
		// Continue with normal startup
	}

	eventsDBPath := filepath.Join(internal.DatFilesPath, "moose.db")

	if snapconf.IsUnderSnap() {
		eventsDBPath = filepath.Join(os.Getenv("SNAP_USER_COMMON"), "moose.db")
		// In case of snap, if default directory is determined to be under $HOME and that
		// is translated to $SNAP_USER_DATA, during the first execution Downloads directory
		// will not be created yet
		downloadsDir, err := fileshare.GetDefaultDownloadDirectory()
		if err != nil {
			log.Warn("failed to get the default downloads directory:", err)
		} else {
			if err := internal.EnsureDir(filepath.Join(downloadsDir, "a")); err != nil {
				log.Warn("failed to ensure default downloads directory:", err)
			}
		}
		drainStart(eventsDBPath)
	}

	if err := os.Remove(internal.FileshareSocket); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Warn("failed to remove old socket file:", err)
	}

	listener, err := internal.ManualListener(internal.FileshareSocket, internal.PermUserRWX)()
	if err != nil {
		log.Error("failed to open unix socket:", err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}
	limitedListener := netutil.LimitListener(listener, 100)

	defer func() {
		if err != nil {
			if err := listener.Close(); err != nil {
				log.Defer("failed to close socket listener on failure:", err)
			}
		}
	}()

	//nolint:staticcheck
	grpcConn, err := grpc.Dial(
		daemonURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("cannot start fileshare daemon:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	defer func() {
		if err != nil {
			if err := grpcConn.Close(); err != nil {
				log.Error("failed to close grpc connection on failure")
			}
		}
	}()

	daemonClient := daemonpb.NewDaemonClient(grpcConn)
	meshClient := meshpb.NewMeshnetClient(grpcConn)

	resp, err := meshClient.IsEnabled(context.Background(), &meshpb.Empty{})
	if err != nil || !resp.GetStatus().GetValue() {
		log.Error("meshnet not enabled:", err)
		os.Exit(int(childprocess.CodeMeshnetNotEnabled))
	}

	defaultDownloadDirectory, err := fileshare.GetDefaultDownloadDirectory()
	if err != nil {
		log.Warn("failed to find default download directory:", err)
		defaultDownloadDirectory = ""
	}

	eventManager := fileshare.NewEventManager(
		internal.IsProdEnv(Environment),
		meshClient,
		fileshare.StdOsInfo{},
		fileshare.NewStdFilesystem("/"),
		defaultDownloadDirectory,
	)

	privKeyResponse, err := meshClient.GetPrivateKey(context.Background(), &meshpb.Empty{})
	if err != nil || privKeyResponse.GetPrivateKey() == "" {
		log.Errorf(" retrieving mesh private key: error: %s, response: %s", err, privKeyResponse.GetPrivateKey())
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	meshPrivKey, err := base64.StdEncoding.DecodeString(privKeyResponse.GetPrivateKey())
	if err != nil || len(meshPrivKey) != 32 {
		log.Error("failed to decode mesh private key")
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	configDirPath, err := internal.GetConfigDirPath(homeDir)
	if err != nil {
		log.Error("failed to get path to OS configuration directory:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	storagePath := filepath.Join(configDirPath, internal.FileshareHistoryFileName)

	if err := internal.EnsureDir(storagePath); err != nil {
		log.Error("failed to ensure dir for transfer history:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	fileshareImplementation, fileshareStorage, err := newFileshare(
		eventManager,
		eventsDBPath,
		internal.IsProdEnv(Environment),
		fileshare.NewPubkeyProvider(meshClient).PubkeyFunc,
		string(meshPrivKey),
		storagePath,
	)
	if err != nil {
		log.Error("can't create fileshare implementation:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	// Before storage handling was implemented in libdrop, we had our own json implementation. It is possible that user
	// still has this history file, so we need to account for that by joining new transfer history with old transfer
	// history. Fileshare process was implemented after this change, so we do not need legacy storage.
	legacyStoragePath := ""

	eventManager.SetFileshare(fileshareImplementation)
	if legacyStoragePath != "" {
		eventManager.SetStorage(storage.NewCombined(legacyStoragePath, fileshareStorage))
	} else {
		eventManager.SetStorage(storage.NewLibdrop(fileshareStorage))
	}

	settings, err := daemonClient.Settings(context.Background(), &daemonpb.Empty{})
	if err != nil {
		log.Error("failed to retrieve daemon setting:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	if settings != nil && settings.Data.UserSettings.Notify {
		err = eventManager.EnableNotifications(fileshareImplementation)
		if err != nil {
			log.Error("failed to enable notifications:", err)
		}
	}

	meshnetIP, err := firstAddressByInterfaceName(nordlynx.InterfaceName)
	if err != nil {
		log.Error("failed to look up meshnet ip:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	err = fileshareImplementation.Enable(meshnetIP)
	if err != nil {
		if errors.Is(err, fileshare.ErrAddressAlreadyInUse) {
			log.Error("mesh already in use:", err)
			os.Exit(int(childprocess.CodeAddressAlreadyInUse))
		}
		log.Error("failed to enable libdrop:", err)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	fileshareHandle := fileshare_startup.Startup(storagePath,
		limitedListener,
		internal.NewFileshareAuthenticator(uint32(os.Getuid())), // #nosec G115
		fileshareImplementation,
		eventManager,
		meshClient,
		grpcConn,
	)

	signals := internal.GetSignalChan()

	log.Info("Daemon has started")
	select {
	case sig := <-signals:
		log.Info("Received signal: ", sig)
	case <-fileshareHandle.GetShutdownChan():
	}

	// Teardown
	log.Info("Stopping fileshare process.")
	fileshareHandle.Shutdown()
}

func firstAddressByInterfaceName(name string) (netip.Addr, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return netip.Addr{}, fmt.Errorf("interface not found: %w", err)
	}

	ips, err := iface.Addrs()
	if err != nil || len(ips) == 0 {
		return netip.Addr{}, fmt.Errorf("interface is missing ips: %w", err)
	}

	ip, err := netip.ParsePrefix(ips[0].String())
	if err != nil {
		return netip.Addr{}, fmt.Errorf("invalid ip format: %w", err)
	}

	return ip.Addr(), nil
}
