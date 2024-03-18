package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// ErrNotStarted when disabling norduser
var ErrNotStarted = errors.New("norduserd wasn't started")

// ChildProcessNorduser manages norduser service through exec.Command
type ChildProcessNorduser struct {
	cmd     *exec.Cmd
	logFile io.Closer
}

func isRunning(uid uint32) (bool, error) {
	// list all norduserd processes, restrict output to uid of the owner
	// #nosec G204 -- arguments are constatn
	output, err := exec.Command("ps", "-C", internal.Norduserd, "-o", "uid=").CombinedOutput()
	if err != nil {
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
func (f *ChildProcessNorduser) Enable(uid uint32, gid uint32) (err error) {
	running, err := isRunning(uid)
	if err != nil {
		return fmt.Errorf("failed to determine if the process is already running: %w", err)
	}

	if running {
		return nil
	}

	// Set up log file
	fileFlags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	logFilePath := internal.GetNorduserdLogPath(strconv.Itoa(int(uid)))
	// #nosec G304 -- logFilePath is properly validated
	logFile, err := os.OpenFile(logFilePath, fileFlags, internal.PermUserRW)
	if err != nil {
		return fmt.Errorf("opening norduser log file %s: %w", logFilePath, err)
	}
	defer func() {
		if err != nil {
			if err := logFile.Close(); err != nil {
				log.Printf("closing norduser log file: %s", err)
			}
		}
	}()
	if err = logFile.Chown(int(uid), int(gid)); err != nil {
		return fmt.Errorf("changing file %s ownership: %w", logFilePath, err)
	}

	// Set up socket dir
	socketDir := filepath.Dir(internal.GetNorduserdSocket(int(uid)))
	if err := os.Mkdir(socketDir, internal.PermUserRWX); err != nil {
		return fmt.Errorf("creating norduser socket dir %s: %w", socketDir, err)
	}
	defer func() {
		if err != nil {
			if err := os.RemoveAll(socketDir); err != nil {
				log.Printf("removing norduser socket dir: %s", err)
			}
		}
	}()
	if err := os.Chown(socketDir, int(uid), int(gid)); err != nil {
		return fmt.Errorf("changing norduser socket dir %s ownership: %w", socketDir, err)
	}

	nordvpnGid, err := internal.GetNordvpnGid()
	if err != nil {
		return fmt.Errorf("determining nordvpn gid: %w", err)
	}

	// #nosec G204 -- no input comes from user
	cmd := exec.Command("/usr/bin/" + internal.Norduserd)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil
	credential := &syscall.Credential{
		Uid:    uid,
		Gid:    gid,
		Groups: []uint32{uint32(nordvpnGid)},
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: credential}

	f.cmd = cmd
	f.logFile = logFile
	return cmd.Start()
}

// Stop teminates norduser process
func (f *ChildProcessNorduser) Stop(uid uint32) error {
	if f.cmd == nil || f.cmd.Process == nil || f.logFile == nil {
		return ErrNotStarted
	}

	if err := os.RemoveAll(filepath.Dir(internal.GetNorduserdSocket(int(uid)))); err != nil {
		log.Println("deleting norduserd socket dir: " + err.Error())
	}
	if err := f.logFile.Close(); err != nil {
		log.Println("closing norduserd process log file: " + err.Error())
	}

	err := f.cmd.Process.Signal(unix.SIGTERM)
	if err != nil {
		return err
	}
	return f.cmd.Wait()
}
