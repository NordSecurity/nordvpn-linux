package internal

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sys/unix"
)

const (
	// listenFdsStart corresponds to `SD_LISTEN_FDS_START`.
	listenFdsStart = 3
)

// systemDFile returns a `os.systemDFile` object for
// systemDFile descriptor passed to this process via systemd fd-passing protocol.
//
// The order of the systemDFile descriptors is preserved in the returned slice.
// `unsetEnv` is typically set to `true` in order to avoid clashes in
// fd usage and to avoid leaking environment flags to child processes.
func systemDFile(unsetEnv bool) *os.File {
	defer func() {
		if unsetEnv {
			if err := os.Unsetenv(ListenPID); err != nil {
				log.Println(DeferPrefix, err)
			}
			if err := os.Unsetenv(ListenFDS); err != nil {
				log.Println(DeferPrefix, err)
			}
			if err := os.Unsetenv(ListenFDNames); err != nil {
				log.Println(DeferPrefix, err)
			}
		}
	}()

	pid, err := strconv.Atoi(os.Getenv(ListenPID))
	if err != nil || pid != os.Getpid() {
		return nil
	}

	nfds, err := strconv.Atoi(os.Getenv(ListenFDS))
	if err != nil || nfds != 1 {
		return nil
	}

	unix.CloseOnExec(listenFdsStart)
	name := os.Getenv(ListenFDNames)

	return os.NewFile(listenFdsStart, name)
}

// SystemDListener returns systemd defined, socket activated listener
func SystemDListener() (net.Listener, error) {
	file := systemDFile(true)
	defer func() {
		if err := file.Close(); err != nil {
			log.Println(DeferPrefix, err)
		}
	}()
	return net.FileListener(file)
}

type runDirListener struct {
	listener net.Listener
	socket   string
	pidfile  string
}

func (rdl *runDirListener) Accept() (net.Conn, error) {
	return rdl.listener.Accept()
}

func (rdl *runDirListener) Close() error {
	listenerCloseErr := rdl.listener.Close()

	cleanPidFile(rdl.pidfile)

	var protoRemoveErr error
	var dirRemoveErr error
	if protoRemoveErr = os.Remove(rdl.socket); protoRemoveErr == nil {
		// It's safe to assume that socket dir was created by us so it can be removed.
		dirRemoveErr = os.Remove(path.Dir(rdl.socket))
		// In case any other files were added to the dir by other, `os.Remove` will fail.
		// Such error is expected and can be ignored
		if dirRemoveErr != nil && errors.Is(dirRemoveErr, os.ErrExist) {
			dirRemoveErr = nil
		}
	}

	return errors.Join(protoRemoveErr, dirRemoveErr, listenerCloseErr)
}

func (rdl *runDirListener) Addr() net.Addr {
	return rdl.listener.Addr()
}

// ManualListenerIfNotInUse returns manually created listener with provided permissions, it also detects if this socket
// is in use by another process, and returns an appropriate error if it is.
func ManualListenerIfNotInUse(socket string, perm fs.FileMode, pidfile string) func() (net.Listener, error) {
	return func() (net.Listener, error) {
		// check if daemon already is running
		if err := checkPidFile(pidfile); err != nil {
			return nil, err
		}

		// we checked if daemon is running, if daemon is not running, then socket file exsits as garbage
		// and should be removed otherwise new socket listener will fail to start
		if FileExists(socket) {
			if err := FileDelete(socket); err != nil {
				log.Println(WarningPrefix, "cleaning socket file:", err)
			}
		}

		if err := os.MkdirAll(
			path.Dir(socket), PermUserRWXGroupRXOthersRX,
		); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating run dir: %w", err)
		}

		listener, err := net.Listen(Proto, socket)
		if err != nil {
			return nil, err
		}

		err = os.Chmod(socket, perm)
		if err != nil {
			return nil, err
		}

		// write PID to file
		pidstring := fmt.Sprintf("%d", os.Getpid())
		if err := FileWrite(pidfile, []byte(pidstring), PermUserRWGroupROthersR); err != nil {
			return nil, err
		}

		return &runDirListener{
			listener: listener,
			socket:   socket,
			pidfile:  pidfile,
		}, nil
	}
}

// ManualListener returns manually created listener with provided permissions
func ManualListener(socket string, perm fs.FileMode) func() (net.Listener, error) {
	return func() (net.Listener, error) {
		if err := os.MkdirAll(
			path.Dir(socket), PermUserRWXGroupRXOthersRX,
		); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating run dir: %w", err)
		}

		listener, err := net.Listen(Proto, socket)
		if err != nil {
			return nil, err
		}

		err = os.Chmod(socket, perm)
		if err != nil {
			return nil, err
		}

		return &runDirListener{
			listener: listener,
			socket:   socket,
		}, nil
	}
}

func checkPidFile(pidfile string) error {
	// check and read pid file
	if FileExists(pidfile) {
		out, err := FileRead(pidfile)
		if err != nil {
			// pid file exists, but is not readable,
			// some garbage from previous run/failure? maybe we can cleanup and proceed?
			return fmt.Errorf("daemon pid file is not readable: %w", err)
		}
		pidFromFile, err := strconv.Atoi(strings.TrimSpace(string(out)))
		if err != nil {
			// pid value is not valid integer, some garbage from previous run/failure?
			log.Println(WarningPrefix, fmt.Errorf("daemon pid file does not contain valid integer value: %w", err))
		} else {
			procFile := fmt.Sprintf("/proc/%d/cmdline", pidFromFile)
			out, err := FileRead(procFile)
			if err == nil && strings.Contains(string(out), "nord") {
				// found process in the process list - not going to start another process, exiting
				return fmt.Errorf("daemon is already running with pid: %d", pidFromFile)
			}
		}
		// invalid pid or process not found - remove pid file
		if err := FileDelete(pidfile); err != nil {
			log.Println(WarningPrefix, "cleaning pid file:", err)
		}
	}
	return nil
}

func cleanPidFile(pidFile string) {
	if pidFile != "" && FileExists(pidFile) {
		if err := FileDelete(pidFile); err != nil {
			log.Println(ErrorPrefix, "removing pid file:", err)
		}
	}
}

// EnsureDir creates all directories along the path excluding the last element.
func EnsureDir(path string) error {
	return EnsureDirFull(filepath.Dir(path))
}

// EnsureDirAll creates all directories along the path.
func EnsureDirFull(path string) error {
	dir, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}
	err = os.MkdirAll(dir, PermUserRWX)
	if err != nil {
		return fmt.Errorf("making directories: %w", err)
	}

	return nil
}

// FileWrite writes the given string array to file, previously flushing it clean
func FileWrite(path string, contents []byte, permissions os.FileMode) error {
	err := EnsureDir(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, contents, permissions)
}

// FileCreate with the given permissions, but leave the closing to the caller.
func FileCreate(path string, permissions os.FileMode) (*os.File, error) {
	if err := EnsureDir(path); err != nil {
		return nil, err
	}

	// #nosec G304 -- no input comes from the user
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	if err := file.Chmod(permissions); err != nil {
		// #nosec G104 -- no writes were made
		file.Close()
		return nil, err
	}
	return file, nil
}

// FileRead reads all file
func FileRead(file string) ([]byte, error) {
	// #nosec G304 -- no input comes from the user
	return os.ReadFile(file)
}

// FileExists checks if the given file exists or not
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// FileWritable checks if the given file exists and is writable by its owner
func FileWritable(path string) bool {
	info, err := os.Stat(path)
	if err == nil && info.Mode().Perm()&0200 == 0200 {
		return true
	} else {
		return false
	}
}

func IsFile(fileName string) bool {
	fileInfo, err := os.Lstat(fileName)
	if err != nil {
		return false
	}

	return fileInfo.Mode().IsRegular()
}

// IsSymLink check if file name is sym link
func IsSymLink(fileName string) bool {
	fileInfo, err := os.Lstat(fileName)
	if err != nil {
		return false
	}

	return fileInfo.Mode().Type() == fs.ModeSymlink
}

// FileDelete deletes file from system
func FileDelete(path string) error {
	return os.Remove(path)
}

// FileUnlock removes ia attributes from a file
func FileUnlock(filepath string) error {
	_, err := exec.Command(ChattrExec, "-ia", filepath).CombinedOutput()
	return err
}

// FileLock adds i attribute from a file
func FileLock(filepath string) error {
	_, err := exec.Command(ChattrExec, "+i", filepath).CombinedOutput()
	return err
}

// IsFileLocked checks if file is immutable
func IsFileLocked(filepath string) bool {
	out, err := exec.Command(LsattrExec, "-l", filepath).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "immutable")
}

// FileCopy copies a file in path src to path dst
func FileCopy(src, dst string) error {
	// #nosec G304 -- no input comes from the user
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	// #nosec G307 -- no writes are made
	defer in.Close()

	// #nosec G304 -- no input comes from the user
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		// #nosec G104 -- errors.Join would be useful here
		out.Close()
		return err
	}
	return out.Close()
}

// FileTemp creates temp file, writes given content to it
// and returns path to temp file
func FileTemp(name string, content []byte) (*os.File, error) {
	file, err := os.CreateTemp("", name)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(content)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func FileSha256(filepath string) (sum []byte, err error) {
	// #nosec G304 -- no input comes from the user
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		// #nosec G104 -- errors.Join would be useful here
		f.Close()
		return nil, err
	}

	return h.Sum(nil), f.Close()
}

// checks if command is in PATH
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// MachineID return unique machine identification id
func MachineID() uuid.UUID {
	// systemd machine id
	out, _ := exec.Command("sh", "-c", "cat /etc/machine-id").Output()
	id := strings.Trim(string(out), "\n")
	if id == "" {
		// failsafe, this is generated without systemd
		out, _ = exec.Command("sh", "-c", "cat /var/lib/dbus/machine-id").Output()
		id = strings.Trim(string(out), "\n")
	}
	if id == "" {
		// this might fail because it requires sudo
		out, _ = exec.Command("sh", "-c", "cat /sys/class/dmi/id/product_uuid").Output()
		id = strings.Trim(string(out), "\n")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return uuid.New()
	}

	machineUUID, err := uuid.Parse(id)
	if err != nil {
		id, _ := uuid.NewRandom()
		return id
	}
	return uuid.NewSHA1(machineUUID, []byte(hostname))
}

type NetLink struct {
	Name    string
	Address string
	Index   int
}

func NetworkLinks() ([]NetLink, error) {
	var res []NetLink
	ifaces, err := net.Interfaces()
	if err != nil {
		return res, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return res, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ipv4 := v.IP.To4()
				if ipv4 != nil && i.Flags&net.FlagUp != 0 {
					mask, _ := v.Mask.Size()
					ip := fmt.Sprintf("%s/%d", ipv4, mask)
					res = append(res, NetLink{
						Name:    i.Name,
						Address: ip,
						Index:   i.Index,
					})
				}
			}
		}
	}
	return res, nil
}

func IsNetworkLinkUnmanaged(link string) bool {
	out, err := exec.Command(NetworkctlExec, "status", link).CombinedOutput()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.Trim(string(out), "\n"), "\n") {
		if strings.Contains(line, "State") && strings.Contains(line, "unmanaged") {
			return true
		}
	}
	return false
}

// PrefixCommonPath is supposed to be used for files which are version specific and not persistent
func PrefixCommonPath(p string) string {
	return prefixPath(p, "PREFIX_COMMON")
}

// PrefixDataPath is supposed to be used for files which are non version specific and persistent
func PrefixDataPath(p string) string {
	return prefixPath(p, "PREFIX_DATA")
}

// PrefixStaticPath is supposed to be used for files which are version specific and persistent
func PrefixStaticPath(p string) string {
	return prefixPath(p, "PREFIX_STATIC")
}

func prefixPath(p string, envKey string) string {
	dir := os.Getenv(envKey)
	if dir == "" {
		dir = "/"
	}
	return filepath.Clean(dir + p)
}

// Will open or create the given file.
// If a file already exists with the given name but is not a regular file,
// e.g. a symlink, it will be deleted and a regular file re-created instead
func OpenOrCreateRegularFile(fileName string, flags int, permission fs.FileMode) (*os.File, error) {
	fileName = filepath.Clean(fileName)
	// check if it is a file before opening, because if it is a pipe will block on open
	if !IsFile(fileName) {
		if err := os.Remove(fileName); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot delete file %s: %w", fileName, err)
		}
	}
	// #nosec G304 -- fileName was cleaned before
	file, err := os.OpenFile(fileName, flags, permission)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", fileName, err)
	}

	// check again that the file is a regular file after it is open to be sure it was not changed between checking and opening it
	if !IsFile(fileName) {
		return nil, fmt.Errorf("not a regular file %s", fileName)
	}

	return file, nil
}

type (
	readfileFunc func(name string) ([]byte, error)
	readdirFunc  func(name string) ([]os.DirEntry, error)
)

var (
	defaultReadfile readfileFunc = os.ReadFile
	defaultReaddir  readdirFunc  = os.ReadDir
)

// IsProcessRunning returns `true` if the executable specified as an argument is being executed, `false` otherwise.
func IsProcessRunning(execPath string) bool {
	isRunning, err := isProcessRunning(execPath, defaultReaddir, defaultReadfile)
	if err != nil {
		log.Println(WarningPrefix, "failed to check if process is running, returning false:", err)
		return false
	}
	return isRunning
}

func isProcessRunning(executablePath string, readdir readdirFunc, readfile readfileFunc) (bool, error) {
	procDirs, err := readdir("/proc")
	if err != nil {
		return false, fmt.Errorf("error while reading /proc directories: %w", err)
	}

	for _, dir := range procDirs {
		cmdlinePath := filepath.Join("/proc", dir.Name(), "cmdline")

		cmdline, err := readfile(cmdlinePath)
		if err != nil {
			continue
		}
		args := strings.Split(string(cmdline), "\x00")
		if len(args) > 0 && args[0] == filepath.Clean(executablePath) {
			return true, nil
		}
	}

	return false, nil
}
