package meshnet

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

type (
	readfileFunc func(name string) ([]byte, error)
	readdirFunc  func(name string) ([]os.DirEntry, error)
)

var (
	defaultReadfile readfileFunc = os.ReadFile
	defaultReaddir  readdirFunc  = os.ReadDir
)

type DefaultProcChecker struct {
	readfile readfileFunc
	readdir  readdirFunc
}

func NewProcessChecker() DefaultProcChecker {
	return DefaultProcChecker{
		readfile: defaultReadfile,
		readdir:  defaultReaddir,
	}
}

// IsFileshareProcess returns true if the process specified by PID
// is nordfileshare process, false otherwise.
func (DefaultProcChecker) IsFileshareProcess(pid PID) bool {
	execPath, err := readExecutablePath(pid)
	if err != nil {
		log.Println(internal.ErrorPrefix, "failed to read process path from /proc", err)
		return false
	}

	return execPath == internal.FileshareBinaryPath
}

func readExecutablePath(pid PID) (string, error) {
	pidStr := strconv.FormatUint(uint64(pid), 10)
	cmdlinePath := filepath.Join("/proc", pidStr, "cmdline")

	cmdline, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return "", err
	}
	args := strings.Split(string(cmdline), "\x00")
	if len(args) == 0 {
		return "", ErrIncorrectCmdlineContent
	}
	return args[0], nil
}

// GiveProcessPID returns process PID if the executable specified
// by `path` argument is being executed, `nil` otherwise.
func (pc DefaultProcChecker) GiveProcessPID(path string) *PID {
	PID, err := giveProcessPID(path, pc.readdir, pc.readfile)
	if err != nil {
		log.Println(internal.WarningPrefix, "failed to check if process is running:", err)
		return nil
	}
	return PID
}

func giveProcessPID(executablePath string, readdir readdirFunc, readfile readfileFunc) (*PID, error) {
	procDirs, err := readdir("/proc")
	if err != nil {
		return nil, fmt.Errorf("error while reading /proc directories: %w", err)
	}

	for _, dir := range procDirs {
		cmdlinePath := filepath.Join("/proc", dir.Name(), "cmdline")

		cmdline, err := readfile(cmdlinePath)
		if err != nil {
			continue
		}
		args := strings.Split(string(cmdline), "\x00")
		if len(args) > 0 && args[0] == filepath.Clean(executablePath) {
			result, err := strconv.ParseUint(dir.Name(), 10, 64)
			if err != nil {
				continue
			}
			PID := PID(result)
			return &PID, nil
		}
	}

	return nil, fmt.Errorf("process with path '%s' not found", executablePath)
}

// CurrentPID returns current process PID
func (pc DefaultProcChecker) CurrentPID() PID {
	return PID(os.Getpid())
}
