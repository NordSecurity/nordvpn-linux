package exitnode

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/stretchr/testify/assert"
)

type CommandExecutorMock struct {
	executedCommands []string
}

func (c *CommandExecutorMock) Execute(command string, arg ...string) ([]byte, error) {
	args := strings.Join(arg, " ")
	cmd := command + " " + args
	c.executedCommands = append(c.executedCommands, cmd)
	return []byte{}, nil
}

func TestResetPeers(t *testing.T) {
	peers := mesh.MachinePeers{
		{
			Address:              netip.MustParseAddr("70.19.250.31"),
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
		},
		{
			Address:              netip.Addr{},
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
		},
	}

	interfaces := []string{"eth0", "eth1"}

	commandExecutor := CommandExecutorMock{}

	server := NewServer(interfaces, commandExecutor.Execute)

	server.ResetPeers(peers, true)

	expectedCommands := []string{"iptables -t filter -D FORWARD -s 70.19.250.31/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 70.19.250.31/32 -j ACCEPT -m comment --comment nordvpn"}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart.",
		expectedCommands, commandExecutor.executedCommands)
}

func TestResetPeers_LANDiscoveryDisabled(t *testing.T) {
	peers := mesh.MachinePeers{
		{
			Address:              netip.MustParseAddr("70.19.250.31"),
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: true,
		},
	}

	interfaces := []string{"eth0", "eth1"}

	commandExecutor := CommandExecutorMock{}

	server := NewServer(interfaces, commandExecutor.Execute)

	server.ResetPeers(peers, false)

	expectedCommands := []string{"iptables -t filter -D FORWARD -s 70.19.250.31/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 70.19.250.31/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn"}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart.",
		expectedCommands, commandExecutor.executedCommands)
}
