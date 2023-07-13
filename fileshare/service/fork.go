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
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// ErrNotStarted when disabling fileshare
var ErrNotStarted = errors.New("fileshare wasn't started")

// ForkFileshare manages fileshare service through exec.Command
type ForkFileshare struct {
	cmd     *exec.Cmd
	logFile io.Closer
}

// Enable starts fileshare process
func (f *ForkFileshare) Enable(uid, gid uint32) (err error) {
	// Set up log file
	fileFlags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	logFilePath := internal.GetFilesharedLogPath(strconv.Itoa(int(uid)))
	// #nosec G304 -- logFilePath is properly validated
	logFile, err := os.OpenFile(logFilePath, fileFlags, internal.PermUserRW)
	if err != nil {
		return fmt.Errorf("opening fileshare log file %s: %w", logFilePath, err)
	}
	defer func() {
		if err != nil {
			if err := logFile.Close(); err != nil {
				log.Printf("closing fileshare log file: %s", err)
			}
		}
	}()
	if err = logFile.Chown(int(uid), int(gid)); err != nil {
		return fmt.Errorf("changing file %s ownership: %w", logFilePath, err)
	}

	// Set up socket dir
	socketDir := filepath.Dir(internal.GetFilesharedSocket(int(uid)))
	if err := os.Mkdir(socketDir, internal.PermUserRWX); err != nil {
		return fmt.Errorf("creating fileshare socket dir %s: %w", socketDir, err)
	}
	defer func() {
		if err != nil {
			if err := os.RemoveAll(socketDir); err != nil {
				log.Printf("removing fileshare socket dir: %s", err)
			}
		}
	}()
	if err := os.Chown(socketDir, int(uid), int(gid)); err != nil {
		return fmt.Errorf("changing fileshare socket dir %s ownership: %w", socketDir, err)
	}

	nordvpnGid, err := internal.GetNordvpnGid()
	if err != nil {
		return fmt.Errorf("determining nordvpn gid: %w", err)
	}

	// #nosec G204 -- no input comes from user
	cmd := exec.Command("/usr/bin/" + internal.Fileshared)
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

// Disable terminates fileshare process
func (f *ForkFileshare) Disable(uid, gid uint32) error {
	// Since this is not a service, disabling is the same as stopping
	return f.Stop(uid, gid)
}

// Stop teminates fileshare process
func (f *ForkFileshare) Stop(uid, _ uint32) error {
	if f.cmd == nil || f.cmd.Process == nil || f.logFile == nil {
		return ErrNotStarted
	}

	if err := os.RemoveAll(filepath.Dir(internal.GetFilesharedSocket(int(uid)))); err != nil {
		log.Println("deleting fileshare socket dir: " + err.Error())
	}
	if err := f.logFile.Close(); err != nil {
		log.Println("closing fileshare process log file: " + err.Error())
	}

	err := f.cmd.Process.Signal(unix.SIGTERM)
	if err != nil {
		return err
	}
	return f.cmd.Wait()
}
