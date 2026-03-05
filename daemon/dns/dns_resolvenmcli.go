package dns

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	execNMCli                   = "nmcli"
	networkManagerConfigDirPath = "/etc/NetworkManager/conf.d/"
	// prefix filename with zz- to guarantee high lexographical order/priority
	networkManagerConfigFilename = "zz-nordvpn-dns.conf"
	networkManagerConfigFilePath = networkManagerConfigDirPath + networkManagerConfigFilename
)

// errCannotGuaranteeConfig is returned if a config file with higher priority exists in NetworkManager config
// directory. In such cases our DNS config cannot be guaranteed to be applied.
var errCannotGuaranteeConfig = errors.New("cannot guarantee that config will be applied")

type nmCliCommandFunc func(...string) ([]byte, error)

func execNMCliCommand(args ...string) ([]byte, error) {
	return exec.Command(execNMCli, args...).CombinedOutput()
}

// NMCli based detection method
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
	return n.filesystemHandle.Remove(networkManagerConfigFilePath)
}

// higherPriorityFileExists returns true if a config file with higher priority(i.e lexicographically greater) exists
// within the NetworkManager config directory.
func (n *NMCli) higherPriorityFileExists() (bool, error) {
	directoryNames, err := n.filesystemHandle.Readdirnames(networkManagerConfigDirPath)
	if err != nil {
		return false, fmt.Errorf("reading directory names: %w", err)
	}

	if len(directoryNames) == 0 {
		return false, nil
	}

	largestFilename := slices.Max(directoryNames)

	return strings.Compare(largestFilename, networkManagerConfigFilename) == 1, nil
}

// Set attempts to configure DNS by creating a NetworkManager configuration file with a global DNS config and reloading
// the general config.
func (n *NMCli) Set(iface string, nameservers []string) error {
	if len(nameservers) == 0 {
		return errors.New("empty nameservers slice was provided")
	}

	higherPriorityFileExists, err := n.higherPriorityFileExists()
	if err != nil {
		log.Println(internal.ErrorPrefix, dnsPrefix, "failed to check if higher priority config file exists:", err)
		return errCannotGuaranteeConfig
	}

	if higherPriorityFileExists {
		return errCannotGuaranteeConfig
	}

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
			"failed to reload after adding a config file NetworkManager, command output:",
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
			"failed to reload after removing a config file NetworkManager, command output:",
			strings.TrimSpace(string(out)))
		return fmt.Errorf("failed to reload NetworkManager config: %w", err)
	}

	return nil
}

func (n *NMCli) Name() string {
	return "nmcli"
}
