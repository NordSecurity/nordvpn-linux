package fileshare_process

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type StartupErrorCode int

const (
	CodeAlreadyRunning StartupErrorCode = iota + 1
	CodeAlreadyRunningForOtherUser
	CodeFailedToCreateUnixScoket
	CodeMeshnetNotEnabled
	CodeAddressAlreadyInUse
	CodeFailedToEnable
)

const FileshareSocket = "/tmp/fileshare.sock"

var FileshareURL = fmt.Sprintf("%s://%s", internal.Proto, FileshareSocket)

type ProcessStatus int

const (
	Running ProcessStatus = iota
	RunningForOtherUser
	NotRunning
)

type FileshareProcess interface {
	// Disable stops the fileshare process
	Disable()
	// ProcessStatus checks the status of fileshare process
	ProcessStatus() ProcessStatus
}

// GRPCFileshareProcess always tries to use main as primary method, and if it fails fallbacks to backup
type GRPCFileshareProcess struct {
}

// NewFileshareService creates CombinedFileshare
func NewFileshareService() GRPCFileshareProcess {
	return GRPCFileshareProcess{}
}

func getFileshareClient() (pb.FileshareClient, error) {
	fileshareConn, err := grpc.Dial(
		FileshareURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("connecting to fileshare daemon: %w", err)
	}

	client := pb.NewFileshareClient(fileshareConn)
	return client, nil
}

// Disable using method that was used to Enable
func (c GRPCFileshareProcess) Disable() {
	client, err := getFileshareClient()
	if err != nil {
		log.Println("failed to initialize fileshare client: ", err)
		return
	}

	_, err = client.Stop(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println("failed to stop fileshare client: ", err)
	}
}

// Disable using method that was used to Enable
func (c GRPCFileshareProcess) ProcessStatus() ProcessStatus {
	client, err := getFileshareClient()
	if err != nil {
		return NotRunning
	}

	_, err = client.Ping(context.Background(), &pb.Empty{})
	if err != nil {
		if strings.Contains(status.Convert(err).Message(), "permission denied") {
			return RunningForOtherUser
		}
		return NotRunning
	}

	return Running
}

type NoopFileshareProcess struct {
}

func (c NoopFileshareProcess) Disable() {
}

func (c NoopFileshareProcess) ProcessStatus() ProcessStatus {
	return NotRunning
}
