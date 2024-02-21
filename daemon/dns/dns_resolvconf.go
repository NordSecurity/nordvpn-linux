package dns

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Executables
const (
	// execResolvconf defines resolvconf executable
	execResolvconf = "resolvconf"
)

// Files
const (
	// resolvconfInterfaceFilePath defines path of the file used to control
	// the order in which resolvconf nameserver information records are processed.
	resolconfInterfaceFilePath = "/etc/resolvconf/interface-order"
)

// Resolvconf based DNS handling method
type Resolvconf struct{}

func (m *Resolvconf) Set(iface string, nameservers []string) error {
	return setDNSWithResolvconf(iface, nameservers)
}

func (m *Resolvconf) Unset(iface string) error {
	return unsetDNSWithResolvconf(iface)
}

func (m *Resolvconf) IsAvailable() bool {
	return internal.IsCommandAvailable(execResolvconf)
}

func (m *Resolvconf) Name() string {
	return "resolvconf"
}

func resolvconfIfacePrefix(filePath string) (string, error) {
	if internal.FileExists(filePath) {
		// #nosec G304 - file path/name is constant
		file, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("opening %s: %w", filePath, err)
		}
		// #nosec G307 -- no writes are made
		defer file.Close()

		return checkForEntry(file), nil
	}
	return "", nil
}

func checkForEntry(data io.Reader) string {
	fscanner := bufio.NewScanner(data)
	re := regexp.MustCompile(`^([A-Za-z0-9-]+)\*$`)
	for fscanner.Scan() {
		match := re.FindStringSubmatch(fscanner.Text())
		if len(match) > 1 {
			return fmt.Sprintf("%s.", match[1])
		}
	}
	return ""
}

func setDNSWithResolvconf(iface string, addresses []string) error {
	var addrs = make([]string, len(addresses))
	for idx, address := range addresses {
		addrs[idx] = "nameserver " + address
	}
	content := strings.Join(addrs, "\n")
	prefix, err := resolvconfIfacePrefix(resolconfInterfaceFilePath)
	if err != nil {
		return fmt.Errorf("determining interface prefix: %w", err)
	}

	// #nosec G204 -- the code would have failed already if iface did not belong
	// to an actual network interface on the system
	cmd := exec.Command(execResolvconf, "-a", prefix+iface, "-m", "0", "-x")
	cmd.Stdin = strings.NewReader(content)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("setting dns with resolvconf: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func unsetDNSWithResolvconf(iface string) error {
	prefix, err := resolvconfIfacePrefix(resolconfInterfaceFilePath)
	if err != nil {
		return fmt.Errorf("determining interface prefix: %w", err)
	}

	// #nosec G204 -- the code would have failed already if iface did not belong
	// to an actual network interface on the system
	cmd := exec.Command(execResolvconf, "-d", prefix+iface, "-f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("calling %s: %s", cmd.String(), strings.Trim(string(out), "\n"))
	}
	return nil
}
