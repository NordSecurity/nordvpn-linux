package exitnode

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	mock "github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

type firewallTables map[string][]string

type IptablesMock struct {
	tables firewallTables
}

func newIptablesMock(t *testing.T) IptablesMock {
	t.Helper()

	tables := make(firewallTables)
	tables["nat"] = []string{}

	tables["filter"] = []string{}

	return IptablesMock{
		tables: tables,
	}
}

func (im IptablesMock) handleCommand(args []string) string {
	tableNameIndex := 0

	for argIndex, arg := range args {
		if arg == "-t" {
			tableNameIndex = argIndex + 1
			break
		}
	}

	tableName := "filter"

	if tableNameIndex != 0 {
		tableName = args[tableNameIndex]
	}

	chain := im.tables[tableName]

	combinedArgs := strings.Join(args, " ")

	if strings.Contains(combinedArgs, "-D") {
		ruleToDelete := strings.ReplaceAll(combinedArgs, "-D", "-A")

		ruleIndex := 0
		for index, rule := range chain {
			if rule == ruleToDelete {
				ruleIndex = index
				break
			}
		}

		im.tables[tableName] = append(chain[:ruleIndex], chain[ruleIndex+1:]...)

		return ""
	}

	if strings.Contains(combinedArgs, "-A") || strings.Contains(combinedArgs, "-I") {
		im.tables[tableName] = append(chain, strings.ReplaceAll(combinedArgs, "-I", "-A"))
		return ""
	}

	// We cheat here by returning list of commands when -L is provided. Normally, iptables returns a
	// list of rules in this case, but in our case returning commands is usually enough(rules are
	// searched for by "nordvpn" comment, contained both in rules and in commands).
	if strings.Contains(combinedArgs, "-S") || strings.Contains(combinedArgs, "-L") {
		return strings.Join(chain, "\n")
	}

	return ""
}

type CommandExecutorMock struct {
	executedCommands []string
	iptables         IptablesMock
}

func newCommandExecutorMock(t *testing.T) *CommandExecutorMock {
	t.Helper()

	return &CommandExecutorMock{
		iptables: newIptablesMock(t),
	}
}

func (cm *CommandExecutorMock) popCommands(t *testing.T) []string {
	t.Helper()

	commands := cm.executedCommands
	cm.executedCommands = []string{}

	return commands
}

func (c *CommandExecutorMock) Execute(command string, arg ...string) ([]byte, error) {
	output := ""

	if command == "iptables" {
		output = c.iptables.handleCommand(arg)
	}

	args := strings.Join(arg, " ")
	cmd := command + " " + args
	c.executedCommands = append(c.executedCommands, cmd)

	return []byte(output), nil
}

type mockMasqueradeSetter struct {
}

func (mm *mockMasqueradeSetter) EnableMasquerading(intfNames []string) error {
	return nil
}

func (mm *mockMasqueradeSetter) ClearMasquerading(intfNames []string) error {
	return nil
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

	commandExecutor := newCommandExecutorMock(t)

	server := NewServer(interfaces, NewNordlynxMasqueradeSetter(), commandExecutor.Execute)

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

	commandExecutor := newCommandExecutorMock(t)

	server := NewServer(interfaces, NewNordlynxMasqueradeSetter(), commandExecutor.Execute)

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

func getNordlynxMasqueradeSetter(t *testing.T, cmdFunc runCommandFunc) MasqueradeSetter {
	t.Helper()

	return &NordlynxMasqueradeSetter{cmdFunc: cmdFunc}
}

func getOpenVPNMasqueradeSetter(t *testing.T, cmdFunc runCommandFunc) MasqueradeSetter {
	t.Helper()

	return &OpenVPNMasqueradeSetter{cmdFunc: cmdFunc}
}

// iptablesCommandsToFirewall converts commands to firewall table, i.e commands as they would be returned
// by iptables -S. Mainly, -I is replaced by -A.
func iptablesCommandsToFirewall(t *testing.T, commands []string) []string {
	t.Helper()

	iptables := []string{}

	for _, command := range commands {
		entry := strings.ReplaceAll(command, "-I", "-A")
		iptables = append(iptables, entry)
	}

	return iptables
}

func TestEnable(t *testing.T) {
	interfaces := []string{"eth0", "eth1"}

	type masqueradeSetterConstructor func(t *testing.T, cmdFunc runCommandFunc) MasqueradeSetter

	expectedFilterTable := []string{
		"-t filter -A FORWARD 1 -s 100.64.0.0/10 -j DROP -m comment --comment nordvpn",
		"-t filter -A FORWARD 1 -d 100.64.0.0/10 -j DROP -m comment --comment nordvpn",
		"-t filter -A FORWARD 1 -d 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment nordvpn",
	}

	expectedNordlynxNATTable := []string{
		"-t nat -A POSTROUTING -s 100.64.0.0/10 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
		"-t nat -A POSTROUTING -s 100.64.0.0/10 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
	}

	expectedOpenVPNNATTable := []string{
		"-t nat -A POSTROUTING -s 100.64.0.0/10 -o nordtun -j MASQUERADE -m comment --comment nordvpn",
		"-t nat -A POSTROUTING -s 100.64.0.0/10 -o eth0 -j MASQUERADE -m comment --comment nordvpn",
		"-t nat -A POSTROUTING -s 100.64.0.0/10 -o eth1 -j MASQUERADE -m comment --comment nordvpn",
	}

	tests := []struct {
		name                                  string
		masqueradeSetterCtorFunc              masqueradeSetterConstructor
		nextMasqueradeSetterCtorFunc          masqueradeSetterConstructor
		expectedNATTable                      []string
		expectedNATTableAfterMasqueradeSwitch []string
	}{
		{
			name:                                  "enable nordlynx",
			masqueradeSetterCtorFunc:              getNordlynxMasqueradeSetter,
			nextMasqueradeSetterCtorFunc:          getOpenVPNMasqueradeSetter,
			expectedNATTable:                      expectedNordlynxNATTable,
			expectedNATTableAfterMasqueradeSwitch: expectedOpenVPNNATTable,
		},
		{
			name:                                  "enable nordlynx",
			masqueradeSetterCtorFunc:              getOpenVPNMasqueradeSetter,
			nextMasqueradeSetterCtorFunc:          getNordlynxMasqueradeSetter,
			expectedNATTable:                      expectedOpenVPNNATTable,
			expectedNATTableAfterMasqueradeSwitch: expectedNordlynxNATTable,
		},
	}

	for _, test2 := range tests {
		t.Run(test2.name, func(t *testing.T) {
			commandExecutor := newCommandExecutorMock(t)
			sysctlSetterMock := mock.SysctlSetterMock{}

			server := Server{
				interfaceNames: interfaces,
				masquerade:     test2.masqueradeSetterCtorFunc(t, commandExecutor.Execute),
				runCommandFunc: commandExecutor.Execute,
				sysctlSetter:   &sysctlSetterMock,
			}

			server.Enable()

			assert.Equal(t, expectedFilterTable, commandExecutor.iptables.tables["filter"],
				"Filter table was configured incorrectly after exit node has been enabled.")

			assert.Equal(t, test2.expectedNATTable, commandExecutor.iptables.tables["nat"],
				"NAT table was configured incorrectly after exit node has been enabled.")

			assert.True(t, sysctlSetterMock.IsSet,
				"Kernel parameter was not set after enabling exit node.")

			server.Disable()

			assert.False(t, sysctlSetterMock.IsSet,
				"Kernel parameter was not unset after enabling exit node.")

			assert.Zero(t, len(commandExecutor.iptables.tables["filter"]),
				"Filter table was not cleaned up after exit node has been disabled.")

			assert.Zero(t, len(commandExecutor.iptables.tables["nat"]),
				"NAT table was not cleaned up after exit node has been disabled.")

			server.Enable()
			server.SetMasquerade(test2.nextMasqueradeSetterCtorFunc(t, commandExecutor.Execute))

			assert.Equal(t, test2.expectedNATTableAfterMasqueradeSwitch,
				commandExecutor.iptables.tables["nat"],
				"NAT table was configured incorrectly after masquerade switch in exit node.")

			assert.True(t, sysctlSetterMock.IsSet,
				"Kernel parameter was not set after enabling exit node.")
		})
	}
}
