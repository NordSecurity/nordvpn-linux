// NordVPN fileshare daemon.
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // #nosec G108 -- http server is not run in production builds
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_startup"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Values set when building the application
var (
	Environment = ""
	PprofPort   = 6961
)

func main() {
	// Pprof
	go func() {
		if internal.IsDevEnv(Environment) {
			// #nosec G114 -- not used in production
			if err := http.ListenAndServe(fmt.Sprintf(":%d", PprofPort), nil); err != nil {
				log.Println(internal.ErrorPrefix, err)
			}
		}
	}()

	// Logging

	log.SetOutput(os.Stdout)

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("can't retrieve current user info: %s", err)
	}
	// we have to hardcode config directory, using os.UserConfigDir is not viable as nordfileshared
	// is spawned by nordvpnd(owned by root) and inherits roots environment variables
	storagePath := filepath.Join(
		currentUser.HomeDir,
		internal.ConfigDirectory,
		internal.FileshareHistoryFile,
	)
	if err := internal.EnsureDir(storagePath); err != nil {
		log.Fatalf("ensuring dir for transfer history file: %s", err)
	}

	eventsDbPath := filepath.Join(internal.DatFilesPath, "moose.db")

	listenerFunction := internal.SystemDListener
	if os.Getenv(internal.ListenPID) != strconv.Itoa(os.Getpid()) {
		listenerFunction = internal.ManualListener(fileshare_startup.ConnURL, internal.PermUserRWX)
	}
	listener, err := listenerFunction()
	if err != nil {
		log.Fatalf("Error on listening to UNIX domain socket: %s\n", err)
	}

	// Before storage handling was implemented in libdrop, we had our own json implementation. It is possible that user
	// still has this history file, so we need to account for that by joining new transfer history with old transfer
	// history.
	legacyStoragePath := filepath.Join(currentUser.HomeDir, internal.ConfigDirectory)

	fileshareHandle, err := fileshare_startup.Startup(storagePath,
		legacyStoragePath,
		eventsDbPath,
		listener,
		Environment,
		nil)
	if err != nil {
		log.Fatalf("Fileshare daemon startup failed: %s", err.Error())
	}

	log.Println(internal.InfoPrefix, "Daemon has started")
	// Teardown
	internal.WaitSignal()
	log.Println("Stopping fileshare daemon.")
	fileshareHandle.Shutdown()
}
