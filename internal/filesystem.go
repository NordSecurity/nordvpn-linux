package internal

import (
	"crypto/sha256"
	"errors"
	"flag"
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
}

func (rdl *runDirListener) Accept() (net.Conn, error) {
	return rdl.listener.Accept()
}

func (rdl *runDirListener) Close() error {
	listenerCloseErr := rdl.listener.Close()

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

// ManualListener returns manually created listener with provided permissions
func ManualListener(socket string, perm fs.FileMode) func() (net.Listener, error) {
	return func() (net.Listener, error) {
		if err := os.MkdirAll(
			path.Dir(socket), PermUserRWGroupRW,
		); err != nil && !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("creating run dir: %w\n", err)
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

func EnsureDir(path string) error {
	dir, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, PermUserRWX)
		if err != nil {
			return err
		}
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

// FileCreateForUser but leave closing to the caller.
func FileCreateForUser(path string, permissions os.FileMode, uid int, gid int) (*os.File, error) {
	file, err := FileCreate(path, permissions)
	if err != nil {
		return nil, err
	}

	if err := file.Chown(uid, gid); err != nil {
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

// Columns formats a list of strings to a tidy column representation
func Columns(input []string) (string, error) {
	cliSize, err := CliDimensions()
	if err != nil {
		// workaround for tests: while running tests stty fails
		// TODO: find a better way
		if flag.Lookup("test.v") != nil {
			return strings.Join(input, " "), err
		}
		return "", err
	}

	// #nosec G204 -- input is properly sanitized
	cmd := exec.Command(ColumnExec, "-c", cliSize[1])
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	for _, arg := range input {
		_, err = stdin.Write([]byte(arg + "\n"))
		if err != nil {
			return "", err
		}
	}

	err = stdin.Close()
	if err != nil {
		return "", err
	}

	result, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(strings.Trim(string(result), "\n"))
	}

	return strings.Trim(string(result), "\n"), nil
}

// Gets the size of CLI window
func CliDimensions() ([]string, error) {
	cmd := exec.Command(SttyExec, "size")
	cmd.Stdin = os.Stdin

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.New(strings.Trim(string(out), "\n"))
	}

	return strings.Split(strings.Trim(string(out), "\n"), " "), nil
}

// IsServiceActive check if given service is active
func IsServiceActive(service string) bool {
	out, err := exec.Command(SystemctlExec, "is-active", service).Output()
	if err != nil {
		return false
	}
	return "active" == strings.Trim(strings.Trim(string(out), "\n"), " ")
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

// SystemUsers returns all non-root user names
func SystemUsers() ([]string, error) {
	// get list of 'human' users on the host system
	out, err := exec.Command("sh", "-c", "awk -F$':' 'BEGIN { ORS=\" \" }; { if ($3 >= 1000 && $3 < 2000) print $1; }' /etc/passwd").CombinedOutput()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.Trim(string(out), " \n"), " "), nil
}

// SystemUsersIDs returns all non-root user ids
func SystemUsersIDs() ([]int64, error) {
	users, err := SystemUsers()
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, u := range users {
		// #nosec G204 -- input is properly sanitized
		out, err := exec.Command(
			"awk",
			"-v", fmt.Sprintf("val=%s", u),
			"-F", ":",
			"$1==val{print $3}",
			"/etc/passwd",
		).CombinedOutput()
		if err != nil {
			continue
		}
		id, err := strconv.ParseInt(strings.Trim(string(out), "\n"), 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// DBUSSessionBusAddress finds user dbus session bus address
func DBUSSessionBusAddress(id int64) string {
	// #nosec G204 -- input is properly sanitized
	out, err := exec.Command("ps", "-u", fmt.Sprintf("%d", id), "-o", "pid=").CombinedOutput()
	if err != nil {
		return ""
	}
	for _, number := range strings.Split(strings.Trim(string(out), "\n"), "\n") {
		pid, err := strconv.ParseInt(strings.Trim(strings.Trim(number, "\n"), " "), 10, 64)
		if err != nil {
			continue
		}
		out, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
		for _, env := range strings.Split(string(out), "\000") {
			if strings.Contains(env, "DBUS_SESSION_BUS_ADDRESS") {
				return env
			}
		}
	}
	return ""
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
