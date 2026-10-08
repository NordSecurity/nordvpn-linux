package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	_ "embed"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/crypto/ssh"
)

//go:embed kickstart.cfg.tmpl
var kickstartTemplate string

func main() {
	version := flag.Int("version", 44, "Fedora release")
	out := flag.String("out", "fedora-nordvpn.qcow2", "image to create")
	diskSize := flag.String("disk", "30G", "disk size")
	memory := flag.Int("memory", 4096, "installer VM memory in MB")
	cpus := flag.Int("cpus", 4, "installer VM CPUs")
	mirror := flag.String("mirror", "https://download.fedoraproject.org/pub/fedora/linux", "Fedora mirror")
	user := flag.String("user", "tester", "user account created in the image")
	timeout := flag.Duration("timeout", 2*time.Hour, "installation timeout")
	flag.Parse()

	if err := build(*version, *out, *diskSize, *memory, *cpus, *mirror, *user, *timeout); err != nil {
		log.Fatal(err)
	}
}

func build(version int, out, diskSize string, memory, cpus int, mirror, user string, timeout time.Duration) error {
	out, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if _, err := os.Stat(out); err == nil {
		return fmt.Errorf("%s already exists", out)
	}

	keyPath := out + ".id_ed25519"
	publicKey, err := ensureSSHKey(keyPath)
	if err != nil {
		return err
	}

	osURL := fmt.Sprintf("%s/releases/%d/Everything/x86_64/os", mirror, version)
	cacheDir := filepath.Join(filepath.Dir(out), fmt.Sprintf(".fedora-%d-pxeboot", version))
	kernel := filepath.Join(cacheDir, "vmlinuz")
	initrd := filepath.Join(cacheDir, "initrd.img")
	for path, url := range map[string]string{
		kernel: osURL + "/images/pxeboot/vmlinuz",
		initrd: osURL + "/images/pxeboot/initrd.img",
	} {
		if err := download(url, path); err != nil {
			return err
		}
	}

	var kickstart bytes.Buffer
	err = template.Must(template.New("ks").Parse(kickstartTemplate)).Execute(&kickstart, map[string]any{
		"Version":   version,
		"User":      user,
		"PublicKey": strings.TrimSpace(publicKey),
	})
	if err != nil {
		return fmt.Errorf("rendering kickstart: %w", err)
	}
	port, stopServer, err := serveKickstart(kickstart.Bytes())
	if err != nil {
		return err
	}
	defer stopServer()

	partial := out + ".part"
	_ = os.Remove(partial)
	if output, err := exec.Command("qemu-img", "create", "-q", "-f", "qcow2", partial, diskSize).CombinedOutput(); err != nil {
		return fmt.Errorf("creating disk: %v: %s", err, output)
	}

	// In QEMU's user networking the host is reachable from the guest at
	// 10.0.2.2, so the installer can fetch the kickstart from our server.
	kernelArgs := strings.Join([]string{
		fmt.Sprintf("inst.ks=http://10.0.2.2:%d/ks.cfg", port),
		"inst.stage2=" + osURL,
		"inst.text",
		"console=ttyS0",
	}, " ")
	cmd := exec.Command("qemu-system-x86_64",
		"-enable-kvm", "-cpu", "host", "-machine", "q35",
		"-m", strconv.Itoa(memory), "-smp", strconv.Itoa(cpus),
		"-drive", "file="+partial+",if=virtio,format=qcow2",
		"-netdev", "user,id=net0", "-device", "virtio-net-pci,netdev=net0",
		"-kernel", kernel, "-initrd", initrd, "-append", kernelArgs,
		"-display", "none", "-serial", "stdio", "-no-reboot",
	)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	log.Printf("installing Fedora %d into %s, this takes a while; installer output follows", version, partial)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting QEMU: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("installer VM failed: %w", err)
		}
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		return fmt.Errorf("installation didn't finish within %s, see the installer output", timeout)
	}

	if err := os.Rename(partial, out); err != nil {
		return err
	}
	log.Printf("image ready, run the tests with:\n\n"+
		"  NORDVPN_E2E_IMAGE=%s NORDVPN_E2E_SSH_KEY=%s NORDVPN_TOKEN=... go test -v -count=1 ./vmtests/\n", out, keyPath)
	return nil
}

func ensureSSHKey(path string) (string, error) {
	if data, err := os.ReadFile(path); err == nil {
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return "", fmt.Errorf("parsing %s: %w", path, err)
		}
		return string(ssh.MarshalAuthorizedKey(signer.PublicKey())), nil
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	block, err := ssh.MarshalPrivateKey(priv, "nordvpn-e2e")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
		return "", err
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		return "", err
	}
	log.Printf("created SSH key %s", path)
	return string(ssh.MarshalAuthorizedKey(sshPub)), nil
}

func download(url, path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	log.Printf("downloading %s", url)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: %s", url, resp.Status)
	}
	tmp := path + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func serveKickstart(kickstart []byte) (int, func(), error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ks.cfg", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("installer fetched the kickstart")
		_, _ = w.Write(kickstart)
	})
	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(l) }()
	return l.Addr().(*net.TCPAddr).Port, func() { server.Close() }, nil
}
