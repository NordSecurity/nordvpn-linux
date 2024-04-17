package service

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/sys/unix"
)

// ErrNotStarted when disabling norduser
var ErrNotStarted = errors.New("norduserd wasn't started")

// ChildProcessNorduser manages norduser service through exec.Command
type ChildProcessNorduser struct {
	mu             sync.Mutex
	commandHandles map[uint32]*exec.Cmd
}

func NewChildProcessNorduser() *ChildProcessNorduser {
	return &ChildProcessNorduser{
		commandHandles: make(map[uint32]*exec.Cmd),
	}
}

func isRunning(uid uint32) (bool, error) {
	// list all norduserd processes, restrict output to uid of the owner
	// #nosec G204 -- arguments are constant
	output, err := exec.Command("ps", "-C", internal.Norduserd, "-o", "uid=").CombinedOutput()
	if err != nil {
		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			// ps returns 1 when no processes are shown
			if exiterr.ExitCode() == 1 {
				return false, nil
			}
		}

		return false, fmt.Errorf("listing processes: %w", err)
	}

	desiredUID := fmt.Sprint(uid)
	uids := string(output)
	for _, uid := range strings.Split(uids, "\n") {
		if strings.Trim(uid, " ") == desiredUID {
			return true, nil
		}
	}

	return false, nil
}

// Enable starts norduser process
func (c *ChildProcessNorduser) Enable(uid uint32, gid uint32, home string) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	running, err := isRunning(uid)
	if err != nil {
		return fmt.Errorf("failed to determine if the process is already running: %w", err)
	}

	if running {
		return nil
	}

	nordvpnGid, err := internal.GetNordvpnGid()
	if err != nil {
		return fmt.Errorf("determining nordvpn gid: %w", err)
	}

	// #nosec G204 -- no input comes from user
	cmd := exec.Command("/usr/bin/" + internal.Norduserd)
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
	c.commandHandles[uid] = cmd

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting the process: %w", err)
	}

	go cmd.Wait()

	return nil
}

// Stop teminates norduser process
func (c *ChildProcessNorduser) Stop(uid uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	commandHandle, ok := c.commandHandles[uid]
	if !ok {
		return fmt.Errorf("command handle not found for given uid")
	}

	if err := commandHandle.Process.Signal(unix.SIGTERM); err != nil {
		return fmt.Errorf("sending SIGTERM to norduser process: %w", err)
	}

	return nil
}

func (c *ChildProcessNorduser) StopAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for uid, commandHandle := range c.commandHandles {
		if err := commandHandle.Process.Signal(unix.SIGTERM); err != nil {
			fmt.Println("sending SIGTERM to norduser process: ", err)
		}
		delete(c.commandHandles, uid)
	}
}
