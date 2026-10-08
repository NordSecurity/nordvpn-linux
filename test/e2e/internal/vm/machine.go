package vm

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	qmpTimeout  = 30 * time.Second
	bootTimeout = 5 * time.Minute
	sshTimeout  = 2 * time.Minute
)

type Config struct {
	Image  string
	SSHKey string
	User   string

	MemoryMB int
	CPUs     int
	Width, Height int
	Display      string
	ArtifactsDir string
}

type Machine struct {
	cfg  Config
	dir  string
	cmd  *exec.Cmd
	qmp  *qmp
	ssh  *ssh.Client
	done chan error

	screen         image.Rectangle // bounds of the last screenshot
	lastScreenshot string
	lastOCR        string
}

func Start(t *testing.T, cfg Config) *Machine {
	t.Helper()
	m := &Machine{cfg: cfg, dir: t.TempDir(), done: make(chan error, 1)}

	image, err := filepath.Abs(cfg.Image)
	if err != nil {
		t.Fatalf("resolving image path: %v", err)
	}
	overlay := filepath.Join(m.dir, "disk.qcow2")
	if out, err := exec.Command("qemu-img", "create", "-q", "-f", "qcow2",
		"-b", image, "-F", "qcow2", overlay).CombinedOutput(); err != nil {
		t.Fatalf("creating disk overlay: %v: %s", err, out)
	}

	sshPort := freePort(t)
	display := cfg.Display
	if display == "" {
		display = "none"
	}
	qmpSocket := filepath.Join(m.dir, "qmp.sock")
	m.cmd = exec.Command("qemu-system-x86_64",
		"-enable-kvm", "-cpu", "host", "-machine", "q35",
		"-m", strconv.Itoa(cfg.MemoryMB), "-smp", strconv.Itoa(cfg.CPUs),
		"-drive", "file="+overlay+",if=virtio,format=qcow2",
		"-netdev", fmt.Sprintf("user,id=net0,hostfwd=tcp:127.0.0.1:%d-:22", sshPort),
		"-device", "virtio-net-pci,netdev=net0",
		"-device", fmt.Sprintf("virtio-vga,xres=%d,yres=%d", cfg.Width, cfg.Height),
		// An absolute pointing device, like a touchscreen without touch:
		// clicks land exactly where requested.
		"-device", "qemu-xhci", "-device", "usb-tablet",
		"-qmp", "unix:"+qmpSocket+",server=on,wait=off",
		"-serial", "file:"+filepath.Join(m.dir, "serial.log"),
		"-display", display,
	)
	var stderr bytes.Buffer
	m.cmd.Stderr = &stderr
	if err := m.cmd.Start(); err != nil {
		t.Fatalf("starting QEMU: %v", err)
	}
	go func() { m.done <- m.cmd.Wait() }()
	t.Cleanup(func() { m.stop(t) })

	if m.qmp, err = dialQMP(qmpSocket, qmpTimeout); err != nil {
		t.Fatalf("%v (QEMU stderr: %s)", err, stderr.String())
	}
	if m.ssh, err = dialSSH(cfg, sshPort, bootTimeout); err != nil {
		t.Fatalf("VM didn't accept SSH within %s: %v", bootTimeout, err)
	}
	return m
}

func (m *Machine) stop(t *testing.T) {
	if t.Failed() {
		m.saveArtifacts(t)
	}
	if m.ssh != nil {
		m.ssh.Close()
	}
	if m.qmp != nil {
		_ = m.qmp.quit()
		m.qmp.close()
	}
	select {
	case <-m.done:
	case <-time.After(10 * time.Second):
		_ = m.cmd.Process.Kill()
		<-m.done
	}
}

func (m *Machine) saveArtifacts(t *testing.T) {
	if m.cfg.ArtifactsDir == "" {
		return
	}
	dir := filepath.Join(m.cfg.ArtifactsDir, strings.ReplaceAll(t.Name(), "/", "_"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Logf("creating artifacts dir: %v", err)
		return
	}
	// Take a fresh screenshot, the screen may have changed since the last one.
	if m.qmp != nil {
		_ = m.qmp.screendump(filepath.Join(dir, "final-screen.png"))
	}
	if m.lastScreenshot != "" {
		copyFile(t, m.lastScreenshot, filepath.Join(dir, "last-checked-screen.png"))
	}
	if m.lastOCR != "" {
		_ = os.WriteFile(filepath.Join(dir, "last-ocr.txt"), []byte(m.lastOCR), 0o644)
	}
	copyFile(t, filepath.Join(m.dir, "serial.log"), filepath.Join(dir, "serial.log"))
	t.Logf("artifacts saved in %s", dir)
}

func copyFile(t *testing.T, from, to string) {
	data, err := os.ReadFile(from)
	if err != nil {
		t.Logf("saving %s: %v", filepath.Base(from), err)
		return
	}
	_ = os.WriteFile(to, data, 0o644)
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("finding a free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func dialSSH(cfg Config, port int, timeout time.Duration) (*ssh.Client, error) {
	key, err := os.ReadFile(cfg.SSHKey)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key: %w", err)
	}
	clientCfg := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		// The VM is a throwaway local guest, its host key is generated on
		// the first boot of each overlay.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(timeout)
	for {
		client, err := ssh.Dial("tcp", addr, clientCfg)
		if err == nil {
			return client, nil
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(2 * time.Second)
	}
}

func (m *Machine) Run(cmd string) (string, error) {
	return m.RunWithInput(cmd, nil)
}

func (m *Machine) RunWithInput(cmd string, stdin io.Reader) (string, error) {
	session, err := m.ssh.NewSession()
	if err != nil {
		return "", fmt.Errorf("opening SSH session: %w", err)
	}
	defer session.Close()
	session.Stdin = stdin
	var out bytes.Buffer
	session.Stdout = &out
	session.Stderr = &out

	done := make(chan error, 1)
	go func() { done <- session.Run(cmd) }()
	select {
	case err = <-done:
	case <-time.After(sshTimeout):
		err = fmt.Errorf("timed out after %s", sshTimeout)
	}
	return out.String(), err
}

func (m *Machine) MustRun(t *testing.T, cmd string) string {
	t.Helper()
	out, err := m.Run(cmd)
	if err != nil {
		t.Fatalf("running %q in VM: %v: %s", cmd, err, out)
	}
	return out
}

const SessionEnv = `env XDG_RUNTIME_DIR=/run/user/$(id -u) ` +
	`DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/$(id -u)/bus WAYLAND_DISPLAY=wayland-0 DISPLAY=:0 `
