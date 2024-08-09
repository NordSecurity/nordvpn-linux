package childprocess

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"google.golang.org/grpc/status"
)

type ProcessClient interface {
	Ping(nowait bool) error
	Stop(disable bool) error
	Restart() error
}

type GRPCChildProcessManager struct {
	processClient     ProcessClient
	processBinaryPath string
}

func NewGRPCChildProcessManager(processClient ProcessClient, processBinaryPath string) *GRPCChildProcessManager {
	return &GRPCChildProcessManager{
		processClient:     processClient,
		processBinaryPath: processBinaryPath,
	}
}

func (g *GRPCChildProcessManager) StartProcess() (StartupErrorCode, error) {
	errChan := make(chan error)
	go func() {
		// #nosec G204 -- arg values are known before even running the program
		err := exec.Command(g.processBinaryPath).Run()
		errChan <- err
	}()

	pingChan := make(chan error)
	// Start another goroutine where we ping the WaitForReady option, so that server has time to start up before we run
	// the acctual command.
	go func() {
		err := g.processClient.Ping(false)
		pingChan <- err
	}()

	select {
	case err := <-errChan:
		if err == nil {
			return 0, fmt.Errorf("process finished unexpectedly")
		}
		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			exitCode := StartupErrorCode(exiterr.ExitCode())
			log.Println(internal.ErrorPrefix, "failed to start:", err)
			return exitCode, nil
		}
		return 0, fmt.Errorf("failed to start the process: %w", err)
	case err := <-pingChan:
		if err != nil {
			return 0, fmt.Errorf("failed to ping the process after starting: %w", err)
		}

		// Process was started and pinged successfully.
		return 0, nil
	}
}

func (g *GRPCChildProcessManager) StopProcess(disable bool) error {
	err := g.processClient.Stop(disable)
	if err != nil {
		return fmt.Errorf("stopping process: %w", err)
	}

	return nil
}

func (g *GRPCChildProcessManager) RestartProcess() error {
	err := g.processClient.Restart()
	if err != nil {
		return fmt.Errorf("restarting process: %w", err)
	}

	return nil
}

func (g *GRPCChildProcessManager) ProcessStatus() ProcessStatus {
	err := g.processClient.Ping(true)
	if err != nil {
		if strings.Contains(status.Convert(err).Message(), "permission denied") {
			return RunningForOtherUser
		}
		return NotRunning
	}

	return Running
}
