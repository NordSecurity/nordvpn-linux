package dns

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	execNMCli                   = "nmcli"
	networkManagerConfigDirPath = "/etc/NetworkManager/conf.d/"
	// NetworkManager will load the config files in lexicographical order. Prefixing the file name with zz- ensures a
	// high priority.
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

type getNetworkManagerConfigFileFunc func() (names []string, err error)

// getNetworkManagerConfigFiles opens /etc/NetworkManager/conf.d/ and returns all of the filenames in that directory.
func getNetworkManagerConfigFiles() (names []string, err error) {
	file, err := os.Open(networkManagerConfigDirPath)
	if err != nil {
		return []string{}, fmt.Errorf("opening file: %w", err)
	}

	return file.Readdirnames(0)
}

// NMCli based detection method
type NMCli struct {
	runNMCliCommandFunc nmCliCommandFunc
	getConfigFilesFunc  getNetworkManagerConfigFileFunc
	filesystemHandle    internal.FileSystemHandle
}

func newNMCli() *NMCli {
	return &NMCli{
		runNMCliCommandFunc: execNMCliCommand,
		getConfigFilesFunc:  getNetworkManagerConfigFiles,
		filesystemHandle:    internal.StdFilesystemHandle{},
	}
}

func (n *NMCli) removeConfigFile() error {
	return n.filesystemHandle.Remove(networkManagerConfigFilePath)
}

// higherPriorityFileExists returns true if a config file with higher priority(i.e lexicographically greater) exists
// within the NetworkManager config directory.
func (n *NMCli) higherPriorityFileExists() (bool, error) {
	configFileNames, err := n.getConfigFilesFunc()
	if err != nil {
		return false, fmt.Errorf("reading directory names: %w", err)
	}

	configFileNames = slices.DeleteFunc(configFileNames, func(filename string) bool {
		return filename == networkManagerConfigFilename
	})

	if len(configFileNames) == 0 {
		return false, nil
	}

	largestFilename := slices.Max(configFileNames)

	return strings.Compare(largestFilename, networkManagerConfigFilename) == 1, nil
}

// Set attempts to configure DNS by creating a NetworkManager configuration file with a global DNS config and reloading
// the general NetworkManager config.
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

	configContents := fmt.Sprintf("[global-dns-domain-*]\nservers=%s\n", strings.Join(nameservers, ","))

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

		if revertErr := n.removeConfigFile(); revertErr != nil {
			err = errors.Join(err, revertErr)
		}
		return fmt.Errorf("failed to reload NetworkManager config: %w", err)
	}

	return nil
}

// Unset attempts to unset DNS configuration by removing the config file created by Set and reloading the general
// NewtorkManager config.
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
