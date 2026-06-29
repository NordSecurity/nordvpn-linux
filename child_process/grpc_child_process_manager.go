package childprocess

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"google.golang.org/grpc/status"
)

const processStopTimeout = 5 * time.Second

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
			log.Error("failed to start:", err)
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
		log.Warnf("gRPC stop for %s: %s",
			filepath.Base(g.processBinaryPath), err)
	}

	// ensure process is terminated
	g.confirmProcessDeath()

	return err
}

func (g *GRPCChildProcessManager) confirmProcessDeath() {
	binaryName := filepath.Base(g.processBinaryPath)

	if pids := internal.FindProcessPIDsByName(binaryName); len(pids) == 0 {
		return
	}

	deadline := time.Now().Add(processStopTimeout)
	for time.Now().Before(deadline) {
		time.Sleep(1 * time.Second)
		if pids := internal.FindProcessPIDsByName(binaryName); len(pids) == 0 {
			return
		}
	}

	if pids := internal.FindProcessPIDsByName(binaryName); len(pids) > 0 {
		log.Warnf("%s still running after %s, force killing PIDs: %v",
			binaryName, processStopTimeout, pids)
		internal.KillStaleProcesses(binaryName, pids)
	}
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
