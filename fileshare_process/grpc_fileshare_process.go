package fileshare_process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/pkg/errors"
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

var (
	ErrAlreadyRunning             = errors.New("already running")
	ErrAlreadyRunningForOtherUser = errors.New("already running for other user")
	ErrFailedToCreateUnixScoket   = errors.New("failed to create unix socket")
	ErrMeshnetNotEnabled          = errors.New("meshnet not enabled")
	ErrAddressAlreadyInUse        = errors.New("address already in use")
	ErrFailedToEnable             = errors.New("failed to enable")
)

var errorCodeToError = map[StartupErrorCode]error{
	CodeAlreadyRunning:             ErrAlreadyRunning,
	CodeAlreadyRunningForOtherUser: ErrAlreadyRunningForOtherUser,
	CodeFailedToCreateUnixScoket:   ErrFailedToCreateUnixScoket,
	CodeMeshnetNotEnabled:          ErrMeshnetNotEnabled,
	CodeAddressAlreadyInUse:        ErrAddressAlreadyInUse,
	CodeFailedToEnable:             ErrFailedToEnable,
}

const (
	FileshareSocket     = "/tmp/fileshare.sock"
	FileshareBinaryName = "nordfileshare_process"
)

func getUserDataDir() string {
	// USER_DATA is set in case of snap.
	dir := os.Getenv("USER_DATA")
	if dir != "" {
		return dir
	}

	return ""
}

var FileshareURL = fmt.Sprintf("%s://%s", internal.Proto, FileshareSocket)
var FileshareDataPath = getUserDataDir()
var FileshareLogPath = getUserDataDir()

// GRPCFileshareProcess is an implementation of FileshareProcess manager, where process is started by an exec call,
// later managed by GRPCs.
type GRPCFileshareProcess struct {
}

// NewGRPCFileshareProcess creates GRPCFileshareProcess manager
func NewGRPCFileshareProcess() GRPCFileshareProcess {
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

func (g GRPCFileshareProcess) StartProcess() error {
	errChan := make(chan error)
	go func() {
		out, err := exec.Command(internal.AppDataPathStatic + "/" + FileshareBinaryName).Output()
		fmt.Println(string(out))
		// fmt.Println("execing: ", err)
		errChan <- err
	}()

	pingChan := make(chan error)
	// Start another goroutine where we ping the WaitForReady option, so that server has time to start up before we run
	// the acctuall command.
	go func() {
		fileshareClient, err := getFileshareClient()
		if err != nil {
			pingChan <- err
		}
		_, err = fileshareClient.Ping(context.Background(), &pb.Empty{}, grpc.WaitForReady(true))
		pingChan <- err
	}()

	select {
	case err := <-errChan:
		// fmt.Println("err chan")
		if exiterr, ok := err.(*exec.ExitError); ok {
			return errorCodeToError[StartupErrorCode(exiterr.ExitCode())]
		} else {
			return ErrFailedToEnable
		}
	case err := <-pingChan:
		// fmt.Println("ping chan")
		if err != nil {
			// fmt.Println("ping err")
			return ErrFailedToEnable
		}
	}

	return nil
}

func (g GRPCFileshareProcess) StopProcess() error {
	client, err := getFileshareClient()
	if err != nil {
		return fmt.Errorf("getting fileshare client: %w", err)
	}

	_, err = client.Stop(context.Background(), &pb.Empty{})
	if err != nil {
		return fmt.Errorf("stopping fileshare client: %w", err)
	}

	return nil
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

// Enable is a stub function, implemented to satisfy the Service interface. Enable is called from within the daemon, but
// in case of the orphan process we want to always start it from the client side. To start this implementation, use
// StartProcess function.
func (g GRPCFileshareProcess) Enable(_, _ uint32) error {
	return nil
}

func (g GRPCFileshareProcess) Disable(_, _ uint32) error {
	return g.Stop(0, 0)
}

func (g GRPCFileshareProcess) Stop(_, _ uint32) error {
	return g.StopProcess()
}
