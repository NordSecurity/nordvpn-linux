package dns

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Executables
const (
	// execResolvectl defines resolvectl executable
	execNMCli                    = "nmcli"
	networkManagerConfigFilePath = "/etc/NetworkManager/conf.d/nordvpn-dns.conf"
)

type nmCliCommandFunc func(...string) ([]byte, error)

func execNMCliCommand(args ...string) ([]byte, error) {
	return exec.Command(execNMCli, args...).CombinedOutput()
}

// Systemd-resolved and resolvectl based DNS handling method
type NMCli struct {
	runNMCliCommandFunc nmCliCommandFunc
	filesystemHandle    internal.FileSystemHandle
}

func newNMCli() *NMCli {
	return &NMCli{
		runNMCliCommandFunc: execNMCliCommand,
		filesystemHandle:    internal.StdFilesystemHandle{},
	}
}

func (n *NMCli) removeConfigFile() error {
	if err := n.filesystemHandle.Remove(networkManagerConfigFilePath); err != nil {
		return fmt.Errorf("removing DNS config file: %w", err)
	}

	return nil
}

func (n *NMCli) Set(iface string, nameservers []string) error {
	configContents := fmt.Sprintf(`[global-dns-domain-*]

servers=%s`, strings.Join(nameservers, ","))

	if err := n.filesystemHandle.WriteFile(networkManagerConfigFilePath,
		[]byte(configContents),
		internal.PermUserRWGroupROthersR); err != nil {
		return fmt.Errorf("creating DNS config file: %w", err)
	}

	if out, err := n.runNMCliCommandFunc("general", "reload"); err != nil {
		log.Println(internal.ErrorPrefix,
			dnsPrefix,
			"failed to reaload after adding a config file NetworkManager, command output: %s",
			strings.TrimSpace(string(out)))

		if err := n.removeConfigFile(); err != nil {
			log.Println(internal.ErrorPrefix, dnsPrefix, "removing config file after a failed reload operation:", err)
		}
		return fmt.Errorf("failed to reload NetworkManager config: %w", err)
	}

	return nil
}

func (n *NMCli) Unset(iface string) error {
	if err := n.removeConfigFile(); err != nil {
		return fmt.Errorf("removing config file: %w", err)
	}

	if out, err := n.runNMCliCommandFunc("general", "reload"); err != nil {
		log.Println(internal.ErrorPrefix,
			dnsPrefix,
			"failed to reaload after removing a config file NetworkManager, command output: %s",
			strings.TrimSpace(string(out)))
		return fmt.Errorf("failed to reload NetworkManager config: %w", err)
	}

	return nil
}

func (n *NMCli) Name() string {
	return "nmcli"
}
