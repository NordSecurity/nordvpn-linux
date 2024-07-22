package norduser

import (
	"log"
	"time"

	childprocess "github.com/NordSecurity/nordvpn-linux/child_process"
	"github.com/NordSecurity/nordvpn-linux/fileshare/fileshare_process"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type FileshareManagementMsg int

const (
	Start FileshareManagementMsg = iota
	Stop
	Shutdown
)

// StartFileshareManagementLoop starts the management loop in separate goroutine and returns a control channel and a
// shutdown channel. Management loop will try to disable fileshare and close the shutdown chan when Shutdown message is
// received.
func StartFileshareManagementLoop() (chan<- FileshareManagementMsg, <-chan interface{}) {
	managementChan := make(chan FileshareManagementMsg)
	shutdownChan := make(chan interface{})

	go fileshareManagementLoop(managementChan, shutdownChan)

	return managementChan, shutdownChan
}

func fileshareManagementLoop(managementChan <-chan FileshareManagementMsg, shutdownChan chan interface{}) {
	fileshareProcessManager := fileshare_process.NewFileshareGRPCProcessManager()
	for msg := range managementChan {
		switch msg {
		case Start:
			fileshareStartupLoop(fileshareProcessManager, managementChan, shutdownChan)
		case Stop:
			log.Println(internal.InfoPrefix, "stopping fileshare")
			if err := fileshareProcessManager.StopProcess(true); err != nil {
				log.Println(internal.ErrorPrefix, "failed to stop fileshare:", err)
			}
		case Shutdown:
			log.Println(internal.InfoPrefix, "stopping fileshare")
			if err := fileshareProcessManager.StopProcess(true); err != nil {
				log.Println(internal.ErrorPrefix, "failed to stop fileshare on shutdown:", err)
			}
			close(shutdownChan)
		}
	}
}

func startFileshare(fileshareProcessManager *childprocess.GRPCChildProcessManager) bool {
	result, err := fileshareProcessManager.StartProcess()
	if err != nil {
		log.Println(internal.ErrorPrefix, "error when starting fileshare:", err)
		return false
	}

	//exhaustive:ignore
	switch result {
	case 0:
		fallthrough
	case childprocess.CodeAddressAlreadyInUse:
		fallthrough
	case childprocess.CodeAlreadyRunningForOtherUser:
		fallthrough
	case childprocess.CodeFailedToCreateUnixScoket:
		fallthrough
	case childprocess.CodeAlreadyRunning:
		log.Println(internal.InfoPrefix, "fileshare started, final result:", result)
		return true
	}

	log.Println(internal.ErrorPrefix, "failed to start fileshare (will retry):", result)
	return false
}

func fileshareStartupLoop(fileshareProcessManager *childprocess.GRPCChildProcessManager,
	managementChan <-chan FileshareManagementMsg,
	shutdownChan chan interface{},
) {
	if startFileshare(fileshareProcessManager) {
		return
	}

	for {
		select {
		case msg := <-managementChan:
			switch msg {
			case Start:
			case Shutdown:
				close(shutdownChan)
				fallthrough
			case Stop:
				return
			}
		case <-time.After(10 * time.Second):
			if startFileshare(fileshareProcessManager) {
				return
			}
		}
	}
}
