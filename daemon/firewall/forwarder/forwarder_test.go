package forwarder

import (
	"errors"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

type CommandExecutorMock struct {
	executedCommands []string
	err              error
	mockedOutputs    map[string]string
}

func newCommandExecutorMock(t *testing.T) *CommandExecutorMock {
	t.Helper()

	return &CommandExecutorMock{
		mockedOutputs: make(map[string]string),
	}
}

func (c *CommandExecutorMock) Execute(command string, arg ...string) ([]byte, error) {
	if c.err != nil {
		return nil, c.err
	}
	args := strings.Join(arg, " ")
	cmd := command + " " + args
	c.executedCommands = append(c.executedCommands, cmd)
	if output, ok := c.mockedOutputs[cmd]; ok {
		return []byte(output), nil
	}
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
		{
			Address:              netip.MustParseAddr("202.242.38.68"),
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: true,
		},
		{
			Address:              netip.Addr{},
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: false,
		},
	}
}

func TestResetPeersExitnode(t *testing.T) {
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
		{
			Address:              netip.MustParseAddr("202.242.38.68"),
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: true,
		},
		{
			Address:              netip.Addr{},
			DoIAllowRouting:      false,
			DoIAllowLocalNetwork: false,
		},
	}

	interfaces := []string{"eth0", "eth1"}

	commandExecutor := newCommandExecutorMock(t)
	commandExecutor.mockedOutputs["iptables -t nat -S POSTROUTING"] = strings.Join(
		[]string{
			"-A POSTROUTING -s 202.242.38.68/32 -o nordtun -j MASQUERADE -m comment --comment nordvpn",
			"-A POSTROUTING -s 202.242.38.68/32 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
			"-A POSTROUTING -s 202.242.38.68/32 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
		}, "\n",
	)
	commandExecutor.mockedOutputs["iptables -t filter -S FORWARD"] = strings.Join(
		[]string{
			// transient rules should be deleted
			"-A FORWARD -s 70.19.250.31 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 202.242.38.68/32 -d 169.254.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 202.242.38.68/32 -d 192.168.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 202.242.38.68/32 -d 172.16.0.0/12 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 202.242.38.68/32 -d 10.0.0.0/8 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			// permanent rules, should not be deleted
			"-A FORWARD -d 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment nordvpn-exitnode-permanent -j ACCEPT",
			"-A FORWARD -d 100.64.0.0/10 -m comment --comment nordvpn-exitnode-permanent -j DROP",
			"-A FORWARD -s 100.64.0.0/10 -m comment --comment nordvpn-exitnode-permanent -j DROP",
			"-A FORWARD -d 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment nordvpn-exitnode-permanent -j ACCEPT",
			// unrelated rule, should not be deleted
			"-A FORWARD -s 150.155.225.161/20 -j DROP",
		}, "\n",
	)

	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{}, &mock.SysctlSetterMock{})

	server.ResetPeers(peers, true, false, false, config.Allowlist{})

	expectedCommands := []string{
		// List nat table so that old nat rules can be deleted. All existing rules should be deleted.
		"iptables -t nat -S POSTROUTING",
		"iptables -t nat -D POSTROUTING -s 202.242.38.68/32 -o nordtun -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -D POSTROUTING -s 202.242.38.68/32 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -D POSTROUTING -s 202.242.38.68/32 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
		// list FORWARD rules in order to delete them
		"iptables -t filter -S FORWARD",
		// Delete old FORWARD rules
		"iptables -t filter -D FORWARD -s 70.19.250.31 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 202.242.38.68/32 -d 169.254.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 202.242.38.68/32 -d 192.168.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 202.242.38.68/32 -d 172.16.0.0/12 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 202.242.38.68/32 -d 10.0.0.0/8 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		// Add new FORWARD rules
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 70.19.250.31/32 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		// Add nat rules for 70.19.250.31/32. 202.242.38.68/32 will not be touched, as routing is not enabled for that peer.
		"iptables -t nat -A POSTROUTING -s 70.19.250.31/32 ! -d 100.64.0.0/10 -j MASQUERADE -m comment --comment nordvpn",
	}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart:\n\nEXPECTED:\n%s\n\nGOT:\n%s",
		strings.Join(expectedCommands, "\n"), strings.Join(commandExecutor.executedCommands, "\n"))
}

func TestResetPeers_LANDiscoveryEnabled(t *testing.T) {
	category.Set(t, category.Unit)

	peers := getPeers()
	interfaces := []string{"eth0", "eth1"}
	commandExecutor := CommandExecutorMock{}
	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{
		Subnets: config.Subnets{"192.168.0.1/32": true},
		Ports:   config.Ports{TCP: map[int64]bool{1000: true}, UDP: map[int64]bool{2000: true, 2001: true}},
	}, &mock.SysctlSetterMock{})

	err := server.ResetPeers(peers, true, false, false, config.Allowlist{})
	assert.NoError(t, err)

	expectedCommands := []string{
		// nat POSTROUTING rules are listed so that they can be deleted. There are no such rules, so
		// no further commands related to nat POSTROUTING are issued.
		"iptables -t nat -S POSTROUTING",
		// FORWARD rules are listed in order to clean old firewall rules. Any rule that contains
		// nordvpn-exitnode-transient should be deleted, but since no such rules is returned,
		// nothing is deleted. Deletion is tested in TestResetPeersExitnode.
		"iptables -t filter -S FORWARD",
		"iptables -t filter -I FORWARD -s 192.168.0.2/32 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 192.168.0.1/32 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 192.168.0.3/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 10.0.0.0/8 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 172.16.0.0/12 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 192.168.0.0/16 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 202.242.38.68/32 -d 169.254.0.0/16 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		// 192.168.0.1 and 192.168.0.2 are the only valid peers that can route to this machine
		// (see getPeers function). MASQUERADE rules are added for those peers.
		"iptables -t nat -A POSTROUTING -s 192.168.0.1/32 ! -d 100.64.0.0/10 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -A POSTROUTING -s 192.168.0.2/32 ! -d 100.64.0.0/10 -j MASQUERADE -m comment --comment nordvpn",
	}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart:\n\nEXPECTED:\n%s\n\nGOT:\n%s",
		strings.Join(expectedCommands, "\n"), strings.Join(commandExecutor.executedCommands, "\n"))
}

func TestResetPeers_LANDiscoveryDisabled(t *testing.T) {
	category.Set(t, category.Unit)

	peers := getPeers()
	interfaces := []string{"eth0", "eth1"}
	commandExecutor := CommandExecutorMock{}
	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{
		Subnets: config.Subnets{"192.168.0.1/32": true},
		Ports:   config.Ports{TCP: map[int64]bool{1000: true}, UDP: map[int64]bool{2000: true, 2001: true}},
	}, &mock.SysctlSetterMock{})

	err := server.ResetPeers(peers, false, false, false, config.Allowlist{})
	assert.NoError(t, err)

	expectedCommands := []string{
		// nat POSTROUTING rules are listed so that they can be deleted. There are no such rules, so
		// no further commands related to nat POSTROUTING are issued.
		"iptables -t nat -S POSTROUTING",
		// FORWARD rules are listed in order to clean old firewall rules. Any rule that contains
		// nordvpn-exitnode-transient should be deleted, but since no such rules is returned,
		// nothing is deleted. Deletion is tested in TestResetPeersExitnode.
		"iptables -t filter -S FORWARD",
		"iptables -t filter -I FORWARD -s 192.168.0.1/32 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 192.168.0.2/32 -j ACCEPT -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t nat -A POSTROUTING -s 192.168.0.1/32 ! -d 100.64.0.0/10 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -A POSTROUTING -s 192.168.0.2/32 ! -d 100.64.0.0/10 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.1/32",
		"iptables -I FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.1/32",
		"iptables -I FORWARD -s 202.242.38.68 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.1/32",
		"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -p tcp -m tcp --dport 1000",
		"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -p udp -m udp --dport 2000:2001",
	}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart.\n\nEXPECTED:\n%s\n\nGOT:\n%s",
		strings.Join(expectedCommands, "\n"), strings.Join(commandExecutor.executedCommands, "\n"))
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
				"iptables -D FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.0/16",
				"iptables -D FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.0/16",
				"iptables -D FORWARD -s 202.242.38.68 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.0/16",
				"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.1/32",
				"iptables -I FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.1/32",
				"iptables -I FORWARD -s 202.242.38.68 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.1/32",
				"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -p tcp -m tcp --dport 1000",
				"iptables -I FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -p udp -m udp --dport 2000:2001",
			},
		},
		{
			name:      "server disabled",
			isEnabled: false,
			expectedCommands: []string{
				"iptables -D FORWARD -s 192.168.0.1 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.0/16",
				"iptables -D FORWARD -s 192.168.0.3 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.0/16",
				"iptables -D FORWARD -s 202.242.38.68 -j ACCEPT -m comment --comment nordvpn-exitnode-allowlist -d 192.168.0.0/16",
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
			server := Forwarder{
				interfaceNames: interfaces,
				runCommandFunc: commandExecutor.Execute,
				allowlistManager: newAllowlist(commandExecutor.Execute, config.Allowlist{
					Subnets: config.Subnets{initialNetwork: true},
				}),
				enabled: test.isEnabled,
			}

			commandExecutor.err = nil

			err := server.ResetPeers(peers, false, false, false, config.Allowlist{})
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
				"Firewall was configured incorrectly after meshnet peer restart.\n\nEXPECTED:\n%s\n\nGOT:\n%s",
				strings.Join(test.expectedCommands, "\n"), strings.Join(commandExecutor.executedCommands, "\n"))
		})
	}
}

func TestDisable(t *testing.T) {
	category.Set(t, category.Firewall)

	commandExecutor := newCommandExecutorMock(t)
	// The contents of the rules is not so important. Basically, we want to make sure that Disable
	// will remove every rule from filter FORWARD chain and nat POSTROUTING returned by querying commands.
	commandExecutor.mockedOutputs["iptables -t filter -S FORWARD"] = strings.Join(
		[]string{
			"-A FORWARD -s 100.77.148.168/32 -m comment --comment nordvpn -j ACCEPT",
			"-A FORWARD -s 22.232.81.241/32 -d 169.254.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 22.232.81.241/32 -d 192.168.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 22.232.81.241/32 -d 172.16.0.0/12 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 22.232.81.241/32 -d 10.0.0.0/8 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -m comment --comment nordvpn-exitnode-transient -j DROP",
			"-A FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -m comment --comment nordvpn-exitnode-transient -j DROP",
			"-A FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -m comment --comment nordvpn-exitnode-transient -j DROP",
			"-A FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -m comment --comment nordvpn-exitnode-transient -j DROP",
			"-A FORWARD -s 230.191.4.88/32 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
			"-A FORWARD -d 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment nordvpn-exitnode-permanent -j ACCEPT",
			"-A FORWARD -d 100.64.0.0/10 -m comment --comment nordvpn-exitnode-permanent -j DROP",
			"-A FORWARD -s 100.64.0.0/10 -m comment --comment nordvpn-exitnode-permanent -j DROP",
			"-A FORWARD -j DOCKER-USER",
			"-A FORWARD -j DOCKER-ISOLATION-STAGE-1",
			"-A FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT",
			"-A FORWARD -o docker0 -j DOCKER",
			"-A FORWARD -i docker0 ! -o docker0 -j ACCEPT",
			"-A FORWARD -i docker0 -o docker0 -j ACCEPT",
		}, "\n",
	)
	commandExecutor.mockedOutputs["iptables -t nat -S POSTROUTING"] = strings.Join(
		[]string{
			"-D POSTROUTING -s 202.242.38.68/32 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
			"-D POSTROUTING -s 202.242.38.68/32 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
			"-D POSTROUTING -s 230.191.4.88/32 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
			"-D POSTROUTING -s 230.191.4.88/32 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
			// this is some random rule, it doesn't contain 'nordvpn' comment so it should not be deleted
			"-D POSTROUTING -s 155.91.117.151/32 -o eth1 -j MASQUERADE",
		}, "\n",
	)

	interfaces := []string{"eth0", "eth1"}

	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{}, &mock.SysctlSetterMock{})
	server.Disable()

	expectedCommands := []string{
		"iptables -t filter -S FORWARD",
		"iptables -t filter -D FORWARD -s 22.232.81.241/32 -d 169.254.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 22.232.81.241/32 -d 192.168.0.0/16 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 22.232.81.241/32 -d 172.16.0.0/12 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 22.232.81.241/32 -d 10.0.0.0/8 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -m comment --comment nordvpn-exitnode-transient -j DROP",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -m comment --comment nordvpn-exitnode-transient -j DROP",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -m comment --comment nordvpn-exitnode-transient -j DROP",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -m comment --comment nordvpn-exitnode-transient -j DROP",
		"iptables -t filter -D FORWARD -s 230.191.4.88/32 -m comment --comment nordvpn-exitnode-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -d 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment nordvpn-exitnode-permanent -j ACCEPT",
		"iptables -t filter -D FORWARD -d 100.64.0.0/10 -m comment --comment nordvpn-exitnode-permanent -j DROP",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -m comment --comment nordvpn-exitnode-permanent -j DROP",
		"iptables -t nat -S POSTROUTING",
		"iptables -t nat -D POSTROUTING -s 202.242.38.68/32 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -D POSTROUTING -s 202.242.38.68/32 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -D POSTROUTING -s 230.191.4.88/32 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
		"iptables -t nat -D POSTROUTING -s 230.191.4.88/32 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
	}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after exit node was disabled: \n EXPECTED: \n%s\n GOT: \n%s",
		strings.Join(expectedCommands, "\n"), strings.Join(commandExecutor.executedCommands, "\n"))
}

func TestFirewall_AllowlistOrdering(t *testing.T) {
	interfaces := []string{"eth0", "eth1"}

	commandExecutor := newCommandExecutorMock(t)
	commandExecutor.mockedOutputs["iptables -t nat -S POSTROUTING"] = strings.Join(
		[]string{}, "\n",
	)
	commandExecutor.mockedOutputs["iptables -t filter -S FORWARD"] = strings.Join(
		[]string{
			"-A FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
			"-A FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
			"-A FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
			"-A FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
			"-A FORWARD -d 1.1.1.1/32 -o eth0 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
			"-A FORWARD -d 1.1.1.1/32 -o eth1 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
			"-A FORWARD -d 2.2.2.2/32 -o eth0 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
			"-A FORWARD -d 2.2.2.1/32 -o eth1 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		}, "\n",
	)

	server := NewServer(interfaces, commandExecutor.Execute, config.Allowlist{}, &mock.SysctlSetterMock{})
	server.enabled = true

	server.ResetFirewall(true, false, true, config.Allowlist{Subnets: config.Subnets{"1.1.1.1/32": true}})

	expectedCommands := []string{
		// list nat rules, no rules are currently present so furhter actions will be taken on the nat table
		"iptables -t nat -S POSTROUTING",
		// list filter rules for the FORWARD chain
		"iptables -t filter -S FORWARD",
		// delete all allowlist rules from the FORWARD table
		"iptables -t filter -D FORWARD -d 1.1.1.1/32 -o eth0 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -d 1.1.1.1/32 -o eth1 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -d 2.2.2.2/32 -o eth0 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		"iptables -t filter -D FORWARD -d 2.2.2.1/32 -o eth1 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		// delete old rules for blocking meshnet traffic
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		// add FORWARDING rules for subnets present in the current allowlist
		// 2.2.2.2/32 is not added as it is not present in the allowlist
		"iptables -I FORWARD -d 1.1.1.1/32 -o eth0 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		"iptables -I FORWARD -d 1.1.1.1/32 -o eth1 -m comment --comment nordvpn-allowlist-transient -j ACCEPT",
		// add rules to drop traffic to local subnets from meshnet peers
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -D FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
		"iptables -t filter -I FORWARD -s 100.64.0.0/10 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-exitnode-transient",
	}

	assert.Equal(t, expectedCommands, commandExecutor.executedCommands,
		"Firewall was configured incorrectly after meshnet peer restart:\n\nEXPECTED:\n%s\n\nGOT:\n%s",
		strings.Join(expectedCommands, "\n"), strings.Join(commandExecutor.executedCommands, "\n"))
}
