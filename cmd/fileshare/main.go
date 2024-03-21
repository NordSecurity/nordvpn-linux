// NordVPN fileshare daemon.
package main

import (
	"errors"
	"log"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_startup"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/net/netutil"
)

// Values set when building the application
var Environment = ""

func openLogFile(path string) (*os.File, error) {
	// #nosec path is constant
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, internal.PermUserRW)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func main() {
	usr, err := user.Current()
	if err != nil {
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	configDirPath, err := internal.GetConfigDirPath(usr.HomeDir)
	if err == nil {
		if logFile, err := openLogFile(filepath.Join(configDirPath + "/" + internal.FileshareLogFileName)); err == nil {
			log.SetOutput(logFile)
			log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
		}
	}

	processStatus := fileshare_process.NewFileshareGRPCProcessManager().ProcessStatus()
	if processStatus == childprocess.Running {
		os.Exit(int(childprocess.CodeAlreadyRunning))
	} else if processStatus == childprocess.RunningForOtherUser {
		log.Println("Cannot start fileshare daemon, it is already running for another user.")
		os.Exit(int(childprocess.CodeAlreadyRunningForOtherUser))
	}

	storagePath := filepath.Join(
		configDirPath,
		internal.FileshareHistoryFile,
	)

	// Before storage handling was implemented in libdrop, we had our own json implementation. It is possible that user
	// still has this history file, so we need to account for that by joining new transfer history with old transfer
	// history. Fileshare process was implemented after this change, so we do not need legacy storage.
	legacyStoragePath := ""

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Printf("Invalid unix user id, failed to convert from string: %s", usr.Uid)
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	if err := os.Remove(internal.FileshareSocket); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Println("Failed to remove old socket file: ", err)
	}

	listener, err := internal.ManualListener(internal.FileshareSocket, internal.PermUserRWGroupRWOthersRW)()
	if err != nil {
		log.Printf("Failed to open unix socket: %s", err)
		os.Exit(int(childprocess.CodeFailedToCreateUnixScoket))
	}
	limitedListener := netutil.LimitListener(listener, 100)

	defer func() {
		if err != nil {
			if err := listener.Close(); err != nil {
				log.Println("Failed to close socket listener on failure: ", err)
			}
		}
	}()

	eventsDBPath := filepath.Join(configDirPath, "moose.db")

	fileshareHandle, err := fileshare_startup.Startup(storagePath,
		legacyStoragePath,
		eventsDBPath,
		limitedListener,
		Environment,
		internal.NewFileshareAuthenticator(uint32(uid)))
	if err != nil {
		log.Println("Failed to start the service: ", err.Error())
		if errors.Is(err, fileshare_startup.ErrMeshNotEnabled) {
			os.Exit(int(childprocess.CodeMeshnetNotEnabled))
		}
		if errors.Is(err, fileshare_startup.ErrMeshAddressAlreadyInUse) {
			os.Exit(int(childprocess.CodeAddressAlreadyInUse))
		}
		os.Exit(int(childprocess.CodeFailedToEnable))
	}

	signals := internal.GetSignalChan()

	log.Println(internal.InfoPrefix, "Daemon has started")
	select {
	case sig := <-signals:
		log.Println("Received signal: ", sig)
	case <-fileshareHandle.GetShutdownChan():
	}

	// Teardown
	log.Println("Stopping fileshare process.")
	fileshareHandle.Shutdown()
}
