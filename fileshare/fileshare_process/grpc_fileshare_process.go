package fileshare_process

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	norduserpb "github.com/NordSecurity/nordvpn-linux/norduser/pb"
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

var errorCodeToProtobuffError = map[StartupErrorCode]norduserpb.StartFileshareStatus{
	CodeAlreadyRunning:             norduserpb.StartFileshareStatus_ALREADY_RUNNING,
	CodeAlreadyRunningForOtherUser: norduserpb.StartFileshareStatus_ALREADY_RUNNING_FOR_OTHER_USER,
	CodeFailedToCreateUnixScoket:   norduserpb.StartFileshareStatus_FAILED_TO_CREATE_UNIX_SOCKET,
	CodeMeshnetNotEnabled:          norduserpb.StartFileshareStatus_MESHNET_NOT_ENABLED,
	CodeAddressAlreadyInUse:        norduserpb.StartFileshareStatus_ADDRESS_ALREADY_IN_USE,
	CodeFailedToEnable:             norduserpb.StartFileshareStatus_FAILED_TO_ENABLE,
}

const (
	FileshareSocket     = "/tmp/fileshare.sock"
	FileshareBinaryName = "nordfileshare"
)

func getUserDataDir() string {
	// USER_DATA is set in case of snap.
	dir := os.Getenv("USER_DATA")
	if dir != "" {
		return dir
	}

	usr, err := user.Current()
	if err != nil {
		log.Println("failed to lookup current user")
	}

	return filepath.Join(usr.HomeDir, ".config", "nordvpn")
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

func (g GRPCFileshareProcess) StartProcess() norduserpb.StartFileshareStatus {
	errChan := make(chan error)
	go func() {
		// #nosec G204 -- arg values are known before even running the program
		_, err := exec.Command(internal.AppDataPathStatic + "/" + FileshareBinaryName).Output()
		errChan <- err
	}()

	pingChan := make(chan error)
	// Start another goroutine where we ping the WaitForReady option, so that server has time to start up before we run
	// the acctual command.
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
		log.Println(err)
		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			if status, ok := errorCodeToProtobuffError[StartupErrorCode(exiterr.ExitCode())]; ok {
				return status
			}
		}
		return norduserpb.StartFileshareStatus_FAILED_TO_ENABLE
	case err := <-pingChan:
		if err != nil {
			return norduserpb.StartFileshareStatus_FAILED_TO_ENABLE
		}
	}

	return norduserpb.StartFileshareStatus_SUCCESS
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
