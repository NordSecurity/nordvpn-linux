package dns

import (
	"bytes"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

const (
	HostsFilePath = "/etc/hosts"
	mark          = "# NordVPN"
)

type Hosts []Host

type Host struct {
	IP         netip.Addr
	FQDN       string
	DomainName string
}

// Add hostname interface
type HostnameSetter interface {
	SetHosts(Hosts) error
	UnsetHosts() error
}

// HostsFileSetter modifies the hosts file in order to add custom DNS
type HostsFileSetter struct {
	filePath string
}

func NewHostsFileSetter(filePath string) *HostsFileSetter {
	return &HostsFileSetter{
		filePath: filePath,
	}
}

func (h Host) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s", h.IP, h.FQDN, h.DomainName, mark)
}

func (s *HostsFileSetter) SetHosts(hosts Hosts) error {
	file, content, err := openReadAndTruncateFile(s.filePath)
	if err != nil {
		return err
	}
	content = setHostLines(content, hosts)

	if _, err := file.Write(content); err != nil {
		// #nosec G104 -- errors.Join would be useful here
		file.Close()
		return fmt.Errorf("writing to the hosts file: %w", err)
	}

	return file.Close()
}

func (s *HostsFileSetter) UnsetHosts() error {
	file, content, err := openReadAndTruncateFile(s.filePath)
	if err != nil {
		return err
	}

	content = removeHostLinesFrom(content)

	if _, err := file.Write(content); err != nil {
		// #nosec G104 -- errors.Join would be useful here
		file.Close()
		return fmt.Errorf("writing to the hosts file: %w", err)
	}

	return file.Close()
}

func openReadAndTruncateFile(filePath string) (*os.File, []byte, error) {
	// #nosec G304 -- no input comes from the user
	file, err := os.OpenFile(filePath, os.O_RDWR, internal.PermUserRWGroupROthersR)
	if err != nil {
		return nil, nil, fmt.Errorf("opening hosts file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return file, content, fmt.Errorf(
			"reading hosts file: %w",
			err,
		)
	}
	if err := file.Truncate(0); err != nil {
		return file, content, fmt.Errorf(
			"truncating hosts file: %w",
			err,
		)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return file, content, fmt.Errorf(
			"seeking hosts file: %w",
			err,
		)
	}
	return file, content, nil
}

// setHostLines removes all of our maintained lines from the hosts file
// and appends the new ones to it
func setHostLines(content []byte, hosts Hosts) []byte {
	return appendHostLines(removeHostLinesFrom(content), hosts)
}

// removeHostLinesFrom removes our maintained lines from the hosts file
// output
func removeHostLinesFrom(content []byte) []byte {
	lines := [][]byte{}
	// Remove the .nord lines
	for _, line := range bytes.Split(content, []byte{'\n'}) {
		if !bytes.HasSuffix(line, []byte(mark)) {
			lines = append(lines, line)
		}
	}
	return bytes.Join(lines, []byte{'\n'})
}

// appendHostLines appends our maintained hosts lines to the end of the
// content
func appendHostLines(content []byte, hosts Hosts) []byte {
	if len(hosts) == 0 {
		return content
	}
	lines := []string{}
	for _, host := range hosts {
		lines = append(lines, host.String())
	}
	return append(bytes.TrimSpace(content), []byte("\n\n"+strings.Join(lines, "\n")+"\n")...)
}
