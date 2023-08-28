package exitnode

import (
	"errors"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

type CommandExecutorMock struct {
	executedCommands []string
	err              error
}

func (c *CommandExecutorMock) Execute(command string, arg ...string) ([]byte, error) {
	if c.err != nil {
		return nil, c.err
	}
	args := strings.Join(arg, " ")
	cmd := command + " " + args
	c.executedCommands = append(c.executedCommands, cmd)
	return []byte{}, nil
}

func getPeers() mesh.MachinePeers {
	return mesh.MachinePeers{
		{
			Address:              netip.MustParseAddr("192.168.0.1"),
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
		},
		{
			Address:              netip.MustParseAddr("192.168.0.2"),
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: false,
		},
		{
			Address:              netip.MustParseAddr("192.168.0.3"),
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: true,
		},
		{
			Address:              netip.MustParseAddr("192.168.0.4"),
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: false,
		},
		{
			Address:              netip.Addr{},
			DoIAllowRouting:      true,
			DoIAllowLocalNetwork: true,
		},
	}
}

func TestResetPeers_LANDiscoveryEnabled(t *testing.T) {
	category.Set(t, category.Unit)

	peers := getPeers()
	interfaces := []string{"eth0", "eth1"}
	commandExecutor := CommandExecutorMock{}
	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{
		Subnets: config.Subnets{"192.168.0.1/32": true},
		Ports:   config.Ports{TCP: map[int64]bool{1000: true}, UDP: map[int64]bool{2000: true, 2001: true}},
	})

	err := server.ResetPeers(peers, true)
	assert.NoError(t, err)

	expectedCommands := []string{
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.2/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.1/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
	}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart.")
}

func TestResetPeers_LANDiscoveryDisabled(t *testing.T) {
	category.Set(t, category.Unit)

	peers := getPeers()
	interfaces := []string{"eth0", "eth1"}
	commandExecutor := CommandExecutorMock{}
	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{
		Subnets: config.Subnets{"192.168.0.1/32": true},
		Ports:   config.Ports{TCP: map[int64]bool{1000: true}, UDP: map[int64]bool{2000: true, 2001: true}},
	})

	err := server.ResetPeers(peers, false)
	assert.NoError(t, err)

	expectedCommands := []string{
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.1/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.2/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.3/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 192.168.0.4/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.1/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 192.168.0.2/32 -j ACCEPT -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn",
		"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.1/32",
		"iptables -I FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.1/32",
		"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -p tcp -m tcp --dport 1000",
		"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -p udp -m udp --dport 2000:2001"}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart.",
		expectedCommands, commandExecutor.executedCommands)
}

func TestSetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	peers := getPeers()
	initialNetwork := "192.168.0.0/16"
	interfaces := []string{"eth0", "eth1"}
	commandExecutor := CommandExecutorMock{}

	tests := []struct {
		name             string
		isEnabled        bool
		expectedCommands []string
		err              error
	}{
		{
			name:      "server enabled",
			isEnabled: true,
			expectedCommands: []string{
				"iptables -D FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.0/16",
				"iptables -D FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.0/16",
				"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.1/32",
				"iptables -I FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.1/32",
				"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -p tcp -m tcp --dport 1000",
				"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -p udp -m udp --dport 2000:2001",
			},
		},
		{
			name:      "server disabled",
			isEnabled: false,
			expectedCommands: []string{
				"iptables -D FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.0/16",
				"iptables -D FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn -d 192.168.0.0/16",
			},
		},
		{
			name:             "error",
			isEnabled:        true,
			err:              errors.New("test error"),
			expectedCommands: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := Server{
				interfaceNames: interfaces,
				runCommandFunc: commandExecutor.Execute,
				allowlistManager: newAllowlist(commandExecutor.Execute, config.Allowlist{
					Subnets: config.Subnets{initialNetwork: true},
				}),
				enabled: test.isEnabled,
			}

			commandExecutor.err = nil

			err := server.ResetPeers(peers, false)
			assert.NoError(t, err)
			// clean expected commands as ResetPeers is covered by other tests
			commandExecutor.executedCommands = commandExecutor.executedCommands[:0]

			commandExecutor.err = test.err

			err = server.SetAllowlist(config.Allowlist{
				Subnets: config.Subnets{"192.168.0.1/32": true, "1.2.3.4/32": true},
				Ports:   config.Ports{TCP: map[int64]bool{1000: true}, UDP: map[int64]bool{2000: true, 2001: true}},
			}, false)
			assert.ErrorIs(t, err, test.err)

			assert.Equal(t, test.expectedCommands, commandExecutor.executedCommands,
				"Firewall was configured incorrectly after meshnet peer restart.",
				test.expectedCommands, commandExecutor.executedCommands)
		})
	}
}
