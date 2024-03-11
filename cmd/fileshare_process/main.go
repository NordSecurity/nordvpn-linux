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

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_startup"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Values set when building the application
var Environment = ""

func openLogFile(path string) (*os.File, error) {
	// #nosec path is constant
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func main() {
	if logFile, err := openLogFile(filepath.Join(fileshare_process.FileshareLogPath, "/nordfileshared.log")); err == nil {
		log.SetOutput(logFile)
		log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	}

	processStatus := fileshare_process.GRPCFileshareProcess{}.ProcessStatus()
	if processStatus == fileshare_process.Running {
		os.Exit(int(fileshare_process.CodeAlreadyRunning))
	} else if processStatus == fileshare_process.RunningForOtherUser {
		os.Exit(int(fileshare_process.CodeAlreadyRunningForOtherUser))
		log.Println("Cannot start fileshare daemon, it is already running for another user.")
	}

	usr, err := user.Current()
	if err != nil {
		log.Println("Failed to retrieve current user: ", err)
		os.Exit(int(fileshare_process.CodeFailedToEnable))
	}

	storagePath := filepath.Join(
		fileshare_process.FileshareDataPath,
		internal.FileshareHistoryFile,
	)

	// Before storage handling was implemented in libdrop, we had our own json implementation. It is possible that user
	// still has this history file, so we need to account for that by joining new transfer history with old transfer
	// history. Fileshare process was implemented after this change, so we do not need legacy storage.
	legacyStoragePath := ""

	uid, err := strconv.Atoi(usr.Uid)
	if err != nil {
		log.Printf("Invalid unix user id, failed to convert from string: %s", usr.Uid)
		os.Exit(int(fileshare_process.CodeFailedToEnable))
	}

	listener, err := internal.ManualListener(fileshare_process.FileshareSocket, internal.PermUserRWGroupRWOthersRW)()
	if err != nil {
		log.Printf("Failed to open unix socket: %s", err)
		os.Exit(int(fileshare_process.CodeFailedToCreateUnixScoket))
	}
	defer func() {
		if err != nil {
			listener.Close()
		}
	}()

	eventsDBPath := filepath.Join(fileshare_process.FileshareDataPath, "moose.db")

	fileshareHandle, err := fileshare_startup.Startup(storagePath,
		legacyStoragePath,
		eventsDBPath,
		listener,
		Environment,
		internal.NewFileshareAuthenticator(uint32(uid)))
	if err != nil {
		log.Println("Failed to start the service: %s", err.Error())
		if errors.Is(err, fileshare_startup.ErrMeshNotEnabled) {
			os.Exit(int(fileshare_process.CodeMeshnetNotEnabled))
		}
		if errors.Is(err, fileshare_startup.ErrMeshAddressAlreadyInUse) {
			os.Exit(int(fileshare_process.CodeAddressAlreadyInUse))
		}
		os.Exit(int(fileshare_process.CodeFailedToEnable))
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
