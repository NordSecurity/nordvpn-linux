package service

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// ErrNotStarted when disabling norduser
var ErrNotStarted = errors.New("norduserd wasn't started")

// ChildProcessNorduser manages norduser service through exec.Command
type ChildProcessNorduser struct {
	mu sync.Mutex
	wg sync.WaitGroup
}

func NewChildProcessNorduser() *ChildProcessNorduser {
	return &ChildProcessNorduser{}
}

// handlePsError returns nil if err is nil or if there is no output. It returns unmodified err in any other
// case.
func handlePsError(out []byte, err error) error {
	if err == nil {
		return nil
	}

	var exiterr *exec.ExitError
	if errors.As(err, &exiterr) {
		// ps returns error when no processes are shown. We do not treat such cases as errors.
		if len(out) == 0 {
			return nil
		}
	}

	return err
}

func parseNorduserPIDs(psOutput string) []int {
	pids := []int{}
	for _, pidStr := range strings.Split(psOutput, "\n") {
		pidStr = strings.TrimSpace(pidStr)
		if pidStr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to parse pid string:", pidStr, "; err:", err)
			continue
		}

		pids = append(pids, pid)
	}

	return pids
}

func getRunningNorduserPIDs() ([]int, error) {
	// #nosec G204 -- arguments are constant
	output, err := exec.Command("ps", "-C", internal.Norduserd, "-o", "pid=").CombinedOutput()
	if err := handlePsError(output, err); err != nil {
		return []int{}, fmt.Errorf("listing norduserd pids: %w", err)
	}

	return parseNorduserPIDs(string(output)), nil
}

func findPIDOfUID(uids string, desiredUID uint32) int {
	for _, uidPid := range strings.Split(uids, "\n") {
		var pid int
		var uid int
		n, err := fmt.Sscanf(uidPid, "%d%d", &uid, &pid)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to parse uid pid line:", uidPid, "; err:", err)
			continue
		}
		if n != 2 {
			log.Println(internal.ErrorPrefix, "invalid input line, expected <uid> <pid> format:", uidPid)
		}
		if uid == int(desiredUID) {
			return pid
		}
	}

	return -1
}

func getPIDForNorduserUID(uid uint32) (int, error) {
	// #nosec G204 -- arguments are constant
	output, err := exec.Command("ps", "-C", internal.Norduserd, "-o", "uid=", "-o", "pid=").CombinedOutput()
	if err := handlePsError(output, err); err != nil {
		return -1, fmt.Errorf("listing norduser uids/pids: %w", err)
	}
	return findPIDOfUID(string(output), uid), nil
}

// Enable starts norduser process
func (c *ChildProcessNorduser) Enable(uid uint32, gid uint32, home string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pid, err := getPIDForNorduserUID(uid)
	if err != nil {
		return fmt.Errorf("failed to determine if the process is already running: %w", err)
	}

	if pid != -1 {
		return nil
	}

	nordvpnGid, err := internal.GetNordvpnGid()
	if err != nil {
		return fmt.Errorf("determining nordvpn gid: %w", err)
	}

	// #nosec G204 -- no input comes from user
	cmd := exec.Command(internal.NorduserdBinaryPath)
	credential := &syscall.Credential{
		Uid:    uid,
		Gid:    gid,
		Groups: []uint32{uint32(nordvpnGid)},
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: credential}
	// os.UserHomeDir always returns value of $HOME and spawning child process copies
	// environment variables from a parent process, therefore value of $HOME will be root home
	// dir, where user usually does not have access.
	cmd.Env = append(cmd.Env, "HOME="+home)

	err = mergeUserSessionEnv(uid, gid, &cmd.Env, NewSystemEnvConfigurator())
	if err != nil {
		log.Println(internal.WarningPrefix, "failed to retrieve user session's environment: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting the process: %w", err)
	}

	c.wg.Add(1)
	go func() {
		cmd.Wait()
		c.wg.Done()
	}()

	return nil
}

// Stop teminates norduser process
func (c *ChildProcessNorduser) Stop(uid uint32, wait bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	pid, err := getPIDForNorduserUID(uid)
	if err != nil {
		return fmt.Errorf("looking up norduserd pid: %w", err)
	}

	if pid == -1 {
		return nil
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			if errno == syscall.ESRCH {
				return nil
			}
		}
		return fmt.Errorf("sending SIGTERM to norduserd: %w", err)
	}

	if wait {
		proc, err := os.FindProcess(pid)
		if err == nil {
			_, _ = proc.Wait()
		}
	}

	return nil
}

func (c *ChildProcessNorduser) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	pids, err := getRunningNorduserPIDs()
	if err != nil {
		return
	}

	for _, pid := range pids {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			log.Println(internal.ErrorPrefix, "failed to send a signal to norduserd:", err)
		}
	}

	doneChan := make(chan interface{})
	go func() {
		c.wg.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-doneChan:
	case <-time.After(10 * time.Second):
	}
}

func (c *ChildProcessNorduser) Restart(uid uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	pid, err := getPIDForNorduserUID(uid)
	if err != nil {
		return fmt.Errorf("looking up norduserd pid: %w", err)
	}

	if pid == -1 {
		return nil
	}

	if err := syscall.Kill(pid, syscall.SIGHUP); err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			if errno == syscall.ESRCH {
				return nil
			}
		}
		return fmt.Errorf("sending SIGHUP to norduserd: %w", err)
	}

	return nil
}

// mergeUserSessionEnv gathers environment variables from user session
// and merges them with the provided environment list.
func mergeUserSessionEnv(uid, gid uint32, currentEnv *[]string, conf EnvConfigurator) error {
	if currentEnv == nil {
		return fmt.Errorf("environemnt slice does not exist")
	}

	// Use the injected configurator to get the prepared command
	cmd, err := conf.ConfigureEnv(uid, gid)
	if err != nil {
		return fmt.Errorf("failed to configure environment: %w", err)
	}

	// Execute the command and retrieve output
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not fetch user environment: %w", err)
	}

	// Scan output and append to currentEnv
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		*currentEnv = append(*currentEnv, scanner.Text())
	}

	return nil
}

type EnvConfigurator interface {
	ConfigureEnv(uid, gid uint32) (*exec.Cmd, error)
}

type systemGIDProvider struct{}

// GetNordvpnGid fetches the actual group ID of NordVPN.
func (s *systemGIDProvider) GetNordvpnGid() (uint32, error) {
	gid, err := internal.GetNordvpnGid()
	if err != nil {
		return 0, err
	}

	if gid < 0 {
		return 0, errors.New("negative gid cannot be converted to uint32")
	}

	if gid > math.MaxUint32 {
		return 0, errors.New("gid exceeds uint32 maximum value")
	}

	return uint32(gid), nil
}

type gidProvider interface {
	GetNordvpnGid() (uint32, error)
}

type SystemEnvConfigurator struct {
	provider gidProvider
}

func NewSystemEnvConfigurator() EnvConfigurator {
	return &SystemEnvConfigurator{provider: &systemGIDProvider{}}
}

func (s *SystemEnvConfigurator) ConfigureEnv(uid, gid uint32) (*exec.Cmd, error) {
	userRtDir := fmt.Sprintf("/run/user/%d", uid)

	cmd := exec.Command("systemctl", "--user", "show-environment")
	cmd.Env = []string{"XDG_RUNTIME_DIR=" + userRtDir}

	nordvpnGid, err := s.provider.GetNordvpnGid()
	if err != nil {
		return nil, fmt.Errorf("determining nordvpn gid: %w", err)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:    uid,
			Gid:    gid,
			Groups: []uint32{nordvpnGid},
		},
	}

	return cmd, nil
}
