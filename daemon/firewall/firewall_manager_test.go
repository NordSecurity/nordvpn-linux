package firewall

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

var ErrIptablesFailure = errors.New("iptables failure")
var ErrGetDevicesFailed error = errors.New("get devices has failed")

const (
	OUTPUT_CHAIN_NAME = "OUTPUT"
	INPUT_CHAIN_NAME  = "INPUT"
)

type iptablesOutput struct {
	tableData []string
	rules     []string
}

func newIptablesOutput(chain string) iptablesOutput {
	tableData := []string{}
	tableData = append(tableData, fmt.Sprintf("Chain %s (policy ACCEPT)", chain))
	tableData = append(tableData, "target     prot opt source               destination")

	return iptablesOutput{
		tableData: tableData,
	}
}

func (i *iptablesOutput) addRules(rules ...string) {
	newRules := []string{}
	newRules = append(newRules, rules...)
	i.rules = append(newRules, i.rules...)
}

func (i *iptablesOutput) get() string {
	iptables := append(i.tableData, i.rules...)
	return strings.Join(iptables, "\n")
}

type commandRunnerMock struct {
	ipv4Commands []string
	ipv6Commands []string
	outputs      map[string]string
	errCommand   string
}

func newCommandRunnerMock() commandRunnerMock {
	return commandRunnerMock{
		outputs: make(map[string]string),
	}
}

func (i *commandRunnerMock) popIPv4Commands() []string {
	commands := i.ipv4Commands
	i.ipv4Commands = nil
	return commands
}

func (i *commandRunnerMock) popIPv6Commands() []string {
	commands := i.ipv6Commands
	i.ipv6Commands = nil
	return commands
}

func (i *commandRunnerMock) addIptablesListOutput(chain string, output string) {
	listCommand := fmt.Sprintf("-L %s --numeric", chain)
	i.outputs[listCommand] = output
}

func (i *commandRunnerMock) runCommand(command string, args string) (string, error) {
	if args == i.errCommand {
		return "", ErrIptablesFailure
	}

	// We do not want to track querying commands(mainly iptables -S/iptables -L) as they do not affect the state.
	// Implementation can achieve the same state with different querying commands, so testing them makes the code
	// unnecessarily complicated and make any changes harder to make.
	if strings.Contains(args, "-L") {
		if output, ok := i.outputs[args]; ok {
			return output, nil
		}
		return "", nil
	}

	if command == iptablesCommand {
		i.ipv4Commands = append(i.ipv4Commands, args)
	} else {
		i.ipv6Commands = append(i.ipv6Commands, args)
	}

	return "", nil
}

// transformCommandsForPriting takes list of commands and combines them into a single string, where commands are
// separated by newline.
func transformCommandsForPrinting(t *testing.T, commands []string) string {
	t.Helper()
	return strings.Join(commands, "\n")
}

func getDeviceFunc(fails bool, ifaces ...net.Interface) func() ([]net.Interface, error) {
	errFunc := func() ([]net.Interface, error) {
		return []net.Interface{}, ErrGetDevicesFailed
	}

	successFunc := func() ([]net.Interface, error) {
		return ifaces, nil
	}

	if fails {
		return errFunc
	}

	return successFunc
}

// transformCommandsToDelete changes the -I flag in all provided commands to -D flag
func transformCommandsToDelte(t *testing.T, oldCommands []string) []string {
	t.Helper()

	expr := regexp.MustCompile(`(-I) ([A-Z]+) (\d)`)

	newCommands := []string{}
	for _, command := range oldCommands {
		newCommands = append(newCommands, expr.ReplaceAllString(command, "-D $2"))
	}
	return newCommands
}

const connmark uint32 = 0x55

func TestTrafficBlocking(t *testing.T) {
	category.Set(t, category.Unit)

	iface0InsertInputCommand := fmt.Sprintf("-I INPUT 1 -i %s -j DROP -m comment --comment nordvpn-0", mock.En0Interface.Name)
	iface0CommandsAfterBlocking := []string{
		iface0InsertInputCommand,
		fmt.Sprintf("-I OUTPUT 1 -o %s -j DROP -m comment --comment nordvpn-0", mock.En0Interface.Name),
	}

	iface1CommandsAfterBlocking := []string{
		fmt.Sprintf("-I INPUT 1 -i %s -j DROP -m comment --comment nordvpn-0", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -j DROP -m comment --comment nordvpn-0", mock.En1Interface.Name),
	}

	iface0DeleteInputCommand := fmt.Sprintf("-D INPUT -i %s -j DROP -m comment --comment nordvpn-0", mock.En0Interface.Name)
	iface0CommandsAfterUnblocking := []string{
		iface0DeleteInputCommand,
		fmt.Sprintf("-D OUTPUT -o %s -j DROP -m comment --comment nordvpn-0", mock.En0Interface.Name),
	}

	iface1CommandsAfterUnblocking := []string{
		fmt.Sprintf("-D INPUT -i %s -j DROP -m comment --comment nordvpn-0", mock.En1Interface.Name),
		fmt.Sprintf("-D OUTPUT -o %s -j DROP -m comment --comment nordvpn-0", mock.En1Interface.Name),
	}

	tests := []struct {
		name                            string
		devicesFunc                     device.ListFunc
		devicesFuncUnblock              device.ListFunc
		errCommand                      string
		expectedErrBlock                error
		expectedErrUnblock              error
		expectedCommandsAfterBlocking   []string
		expectedCommandsAfterUnblocking []string
	}{
		{
			name:                            "success single interface",
			devicesFunc:                     getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterBlocking:   iface0CommandsAfterBlocking,
			expectedCommandsAfterUnblocking: iface0CommandsAfterUnblocking,
		},
		{
			name:                            "success two interfaces",
			devicesFunc:                     getDeviceFunc(false, mock.En0Interface, mock.En1Interface),
			expectedCommandsAfterBlocking:   append(iface0CommandsAfterBlocking, iface1CommandsAfterBlocking...),
			expectedCommandsAfterUnblocking: append(iface0CommandsAfterUnblocking, iface1CommandsAfterUnblocking...),
		},
		{
			name:             "get devices failure",
			devicesFunc:      getDeviceFunc(true),
			expectedErrBlock: ErrGetDevicesFailed,
		},
		{
			name:             "block failure",
			devicesFunc:      getDeviceFunc(false, mock.En0Interface),
			errCommand:       iface0InsertInputCommand,
			expectedErrBlock: ErrIptablesFailure,
		},
		{
			name:                          "unblock failure",
			devicesFunc:                   getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterBlocking: iface0CommandsAfterBlocking,
			errCommand:                    iface0DeleteInputCommand,
			expectedErrUnblock:            ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := newCommandRunnerMock()

			// When testing a single type rule, we do not care so much about ordering/respecting priority, so it is
			// enough to set output to empty iptables(necessary because output processing would crash the tests otherwise).
			outputChain := newIptablesOutput(OUTPUT_CHAIN_NAME)
			commandRunnerMock.addIptablesListOutput(OUTPUT_CHAIN_NAME, outputChain.get())

			inputChain := newIptablesOutput(INPUT_CHAIN_NAME)
			commandRunnerMock.addIptablesListOutput(INPUT_CHAIN_NAME, inputChain.get())

			if test.errCommand != "" {
				commandRunnerMock.errCommand = test.errCommand
			}

			firewallManager := NewFirewallManager(test.devicesFunc, &commandRunnerMock, connmark, true, true)

			err := firewallManager.BlockTraffic()

			if test.expectedErrBlock != nil {
				assert.ErrorIs(t, err, test.expectedErrBlock, "Unexpected error returned after block has failed.")
				return
			}

			commandsIPv4 := commandRunnerMock.popIPv4Commands()
			commandsIPv6 := commandRunnerMock.popIPv6Commands()
			expectedNumberOfCommands := len(test.expectedCommandsAfterBlocking)
			assert.Len(t,
				commandsIPv4,
				expectedNumberOfCommands,
				"Invalid number of IPv4 commands when blocking traffic. Commands:\n%s",
				transformCommandsForPrinting(t, commandsIPv4))
			assert.Len(t,
				commandsIPv6,
				expectedNumberOfCommands,
				"Invalid number of IPv6 commands when blocking traffic. Commands:\n%s",
				transformCommandsForPrinting(t, commandsIPv6))

			// rules are added to two different chains, so ordering doesn't matter in this case and we can use Contains
			for _, expectedCommand := range test.expectedCommandsAfterBlocking {
				assert.Contains(t,
					commandsIPv4,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")

				assert.Contains(t,
					commandsIPv6,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")
			}

			err = firewallManager.UnblockTraffic()

			if test.expectedErrUnblock != nil {
				assert.ErrorIs(t, err, test.expectedErrUnblock, "Unexpected error returned after unblock has failed.")
				return
			}

			commandsIPv4 = commandRunnerMock.popIPv4Commands()
			commandsIPv6 = commandRunnerMock.popIPv6Commands()
			expectedNumberOfCommands = len(test.expectedCommandsAfterUnblocking)
			assert.Len(t, commandsIPv4, expectedNumberOfCommands, "Invalid number of commands when unblocking traffic.")
			assert.Len(t, commandsIPv6, expectedNumberOfCommands, "Invalid number of commands when unblocking traffic.")

			for _, expectedCommand := range test.expectedCommandsAfterUnblocking {
				assert.Contains(t,
					commandsIPv4,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")

				assert.Contains(t,
					commandsIPv6,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")
			}
		})
	}
}

func TestBlockTraffic_AlreadyBlocked(t *testing.T) {
	category.Set(t, category.Unit)

	iface0CommandsAfterBlocking := []string{
		fmt.Sprintf("-I INPUT 1 -i %s -j DROP -m comment --comment nordvpn-0", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -j DROP -m comment --comment nordvpn-0", mock.En0Interface.Name),
	}

	commandRunnerMock := newCommandRunnerMock()
	firewallManager := NewFirewallManager(getDeviceFunc(false, mock.En0Interface), &commandRunnerMock, connmark, true, true)

	// When testing a single type rule, we do not care so much about ordering/respecting priority, so it is
	// enough to set output to empty iptables(necessary because output processing would crash the tests otherwise).
	outputChain := newIptablesOutput(OUTPUT_CHAIN_NAME)
	commandRunnerMock.addIptablesListOutput(OUTPUT_CHAIN_NAME, outputChain.get())

	inputChain := newIptablesOutput(INPUT_CHAIN_NAME)
	commandRunnerMock.addIptablesListOutput(INPUT_CHAIN_NAME, inputChain.get())

	err := firewallManager.BlockTraffic()
	assert.Nil(t, err, "Received unexpected error when blocking traffic.")

	commands := commandRunnerMock.popIPv4Commands()
	assert.Equal(t, iface0CommandsAfterBlocking, commands, "Invalid commands executed when blocking traffic.")

	err = firewallManager.BlockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received after blocking traffic a second time.")

	commands = commandRunnerMock.popIPv4Commands()
	assert.Empty(t, commands, "Commands were executed after blocking traffic for a second time.")
}

func TestUnblockTraffic_TrafficNotBlocked(t *testing.T) {
	category.Set(t, category.Unit)

	commandRunnerMock := newCommandRunnerMock()
	firewallManager := NewFirewallManager(getDeviceFunc(false), &commandRunnerMock, connmark, true, true)

	err := firewallManager.UnblockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received when unblocking traffic when it was not blocked.")

	commands := commandRunnerMock.popIPv4Commands()
	assert.Empty(t, commands, "Commands were executed when ublocking traffic when it was not blocked.")
}

func TestSetAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	udpPorts := []int{
		30000,
		30001,
		30002,
		40000,
	}

	tcpPorts := []int{
		50002,
		50003,
		50004,
		60000,
	}

	subnets := []netip.Prefix{
		netip.MustParsePrefix("102.56.52.223/22"),
	}

	expectedCommandsIface0 := []string{
		fmt.Sprintf("-I INPUT 1 -s 102.56.52.223/22 -i %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -d 102.56.52.223/22 -o %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --dport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --sport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --dport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --sport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --dport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --sport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --dport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --sport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --dport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --sport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --dport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --sport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --dport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --sport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --dport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --sport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
	}

	expectedCommandsIface1 := []string{
		fmt.Sprintf("-I INPUT 1 -s 102.56.52.223/22 -i %s -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -d 102.56.52.223/22 -o %s -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --dport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --sport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --dport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --sport 30000:30002 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --dport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --sport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --dport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --sport 40000:40000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --dport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --sport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --dport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --sport 50002:50004 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --dport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p tcp -m tcp --sport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --dport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p tcp -m tcp --sport 60000:60000 -j ACCEPT -m comment --comment nordvpn-3", mock.En1Interface.Name),
	}

	tests := []struct {
		name                     string
		deviceFunc               device.ListFunc
		firewallDisabled         bool
		expectedCommandsAfterSet []string
		invalidCommand           string
		expectedErrSet           error
		expectedErrUnset         error
	}{
		{
			name:                     "success",
			deviceFunc:               getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterSet: expectedCommandsIface0,
		},
		{
			name:                     "success two interfaces",
			deviceFunc:               getDeviceFunc(false, mock.En0Interface, mock.En1Interface),
			expectedCommandsAfterSet: append(expectedCommandsIface0, expectedCommandsIface1...),
		},
		{
			name:           "failure to get devices",
			deviceFunc:     getDeviceFunc(true),
			expectedErrSet: ErrGetDevicesFailed,
		},
		{
			name:           "iptables failure when setting",
			invalidCommand: fmt.Sprintf("-I INPUT 1 -s 102.56.52.223/22 -i %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
			deviceFunc:     getDeviceFunc(false, mock.En0Interface),
			expectedErrSet: ErrIptablesFailure,
		},
		{
			name:                     "iptables failure when unsetting",
			invalidCommand:           fmt.Sprintf("-D INPUT -s 102.56.52.223/22 -i %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
			deviceFunc:               getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterSet: expectedCommandsIface0,
			expectedErrUnset:         ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := newCommandRunnerMock()
			if test.invalidCommand != "" {
				commandRunnerMock.errCommand = test.invalidCommand
			}

			// When testing a single type rule, we do not care so much about ordering/respecting priority, so it is
			// enough to set output to empty iptables(necessary because output processing would crash the tests otherwise).
			outputChain := newIptablesOutput(OUTPUT_CHAIN_NAME)
			commandRunnerMock.addIptablesListOutput(OUTPUT_CHAIN_NAME, outputChain.get())

			inputChain := newIptablesOutput(INPUT_CHAIN_NAME)
			commandRunnerMock.addIptablesListOutput(INPUT_CHAIN_NAME, inputChain.get())

			firewallManager := NewFirewallManager(test.deviceFunc, &commandRunnerMock, connmark, true, true)

			err := firewallManager.SetAllowlist(udpPorts, tcpPorts, subnets)
			if test.expectedErrSet != nil {
				assert.ErrorIs(t, err, test.expectedErrSet, "Invalid error returned by SetAllowlist.")
				return
			}

			commandsAfterSet := commandRunnerMock.popIPv4Commands()
			assert.Len(t,
				commandsAfterSet,
				len(test.expectedCommandsAfterSet),
				"Invalid commands executed after setting allowlist.\nExpected:\n%s,\nGot:\n%s",
				transformCommandsForPrinting(t, test.expectedCommandsAfterSet),
				transformCommandsForPrinting(t, commandsAfterSet))
			for _, expectedCommand := range test.expectedCommandsAfterSet {
				assert.Contains(t,
					commandsAfterSet,
					expectedCommand,
					"Expected command not executed after setting allowlist.\nExpected command:\n%s\nExecuted commands:\n%s",
					expectedCommand,
					transformCommandsForPrinting(t, commandsAfterSet))
			}

			err = firewallManager.UnsetAllowlist()
			if test.expectedErrUnset != nil {
				assert.ErrorIs(t, err, test.expectedErrUnset, "Invalid error returned by UnsetAllowlist.")
				return
			}

			// same commands should be performed, just with -D flag instead of -I flag
			expectedCommandsAfterUnset := transformCommandsToDelte(t, test.expectedCommandsAfterSet)
			commandsAfterUnset := commandRunnerMock.popIPv4Commands()
			assert.Len(t,
				commandsAfterUnset,
				len(expectedCommandsAfterUnset),
				"Invalid commands executed after unseting allowlist.\nExpected:\n%s\nGot:\n%s",
				transformCommandsForPrinting(t, expectedCommandsAfterUnset),
				transformCommandsForPrinting(t, commandsAfterSet))
			for _, expectedCommand := range transformCommandsToDelte(t, test.expectedCommandsAfterSet) {
				assert.Contains(t,
					commandsAfterUnset,
					expectedCommand,
					"Expected command not executed after unsetting the allowlist.\nExpected command:\n%s\nExecuted commands:\n%s",
					expectedCommand,
					transformCommandsForPrinting(t, commandsAfterUnset))
			}
		})
	}
}

func TestSetAllowlist_IPv6(t *testing.T) {
	category.Set(t, category.Unit)

	udpPorts := []int{
		30000,
	}

	// Both IPv4 and IPv6 commands should be executed for ports.
	expectedCommandsAfterSet := []string{
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --dport 30000:30000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT 1 -i %s -p udp -m udp --sport 30000:30000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --dport 30000:30000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -o %s -p udp -m udp --sport 30000:30000 -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
	}

	subnets := []netip.Prefix{
		netip.MustParsePrefix("7628:c55b:3450:b739:bb1f:6112:a544:9226/30"),
	}

	subnetCommands := []string{
		fmt.Sprintf("-I INPUT 1 -s 7628:c55b:3450:b739:bb1f:6112:a544:9226/30 -i %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT 1 -d 7628:c55b:3450:b739:bb1f:6112:a544:9226/30 -o %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
	}

	// Only IPv6 commands should be executed for subnets.
	expectedIPv6CommandsAfterSet := append(subnetCommands, expectedCommandsAfterSet...)

	commandRunnerMock := newCommandRunnerMock()

	// When testing a single type rule, we do not care so much about ordering/respecting priority, so it is
	// enough to set output to empty iptables(necessary because output processing would crash the tests otherwise).
	outputChain := newIptablesOutput(OUTPUT_CHAIN_NAME)
	commandRunnerMock.addIptablesListOutput(OUTPUT_CHAIN_NAME, outputChain.get())

	inputChain := newIptablesOutput(INPUT_CHAIN_NAME)
	commandRunnerMock.addIptablesListOutput(INPUT_CHAIN_NAME, inputChain.get())

	firewallManager := NewFirewallManager(getDeviceFunc(false, mock.En0Interface), &commandRunnerMock, connmark, true, true)

	firewallManager.SetAllowlist(udpPorts, nil, subnets)

	commands := commandRunnerMock.popIPv4Commands()
	assert.Equal(t,
		expectedCommandsAfterSet,
		commands,
		"Invalid ipv4 commands after setting allowlist.\nExpected commands:\n%s\nExecuted commands:\n%s",
		transformCommandsForPrinting(t, expectedCommandsAfterSet),
		transformCommandsForPrinting(t, commands))

	commands = commandRunnerMock.popIPv6Commands()
	assert.Equal(t,
		expectedIPv6CommandsAfterSet,
		commands,
		"Invalid ipv6 commands after unsetting allowlist.\nExpected commands:\n%s\nExecuted commands:\n%s",
		transformCommandsForPrinting(t, expectedIPv6CommandsAfterSet),
		transformCommandsForPrinting(t, commands))

	expectedCommandsAfterUnset := transformCommandsToDelte(t, expectedCommandsAfterSet)
	expectedIPv6CommandsAfterUnset := transformCommandsToDelte(t, expectedIPv6CommandsAfterSet)

	firewallManager.UnsetAllowlist()

	commands = commandRunnerMock.popIPv4Commands()
	assert.Equal(t, expectedCommandsAfterUnset, commands)

	commands = commandRunnerMock.popIPv6Commands()
	assert.Equal(t, expectedIPv6CommandsAfterUnset, commands)
}

func TestApiAllowlist(t *testing.T) {
	category.Set(t, category.Unit)

	allowlistCommand := fmt.Sprintf("-I INPUT 1 -i %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-1", mock.En0Interface.Name, connmark)
	expectedAllowlistCommandsIf0 := []string{
		allowlistCommand,
		fmt.Sprintf("-I OUTPUT 1 -o %s -m mark --mark %d -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff -m comment --comment nordvpn-1", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-I OUTPUT 1 -o %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-2", mock.En0Interface.Name, connmark),
	}

	denylistCommand := fmt.Sprintf("-D INPUT -i %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-1", mock.En0Interface.Name, connmark)
	expectedDenylistCommandsIf0 := []string{
		denylistCommand,
		fmt.Sprintf("-D OUTPUT -o %s -m mark --mark %d -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff -m comment --comment nordvpn-1", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-D OUTPUT -o %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-2", mock.En0Interface.Name, connmark),
	}

	expectedAllowlistCommandsIf1 := []string{
		fmt.Sprintf("-I INPUT 1 -i %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-1", mock.En1Interface.Name, connmark),
		fmt.Sprintf("-I OUTPUT 1 -o %s -m mark --mark %d -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff -m comment --comment nordvpn-1", mock.En1Interface.Name, connmark),
		fmt.Sprintf("-I OUTPUT 1 -o %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-2", mock.En1Interface.Name, connmark),
	}

	expectedDenylistCommandsIf1 := []string{
		fmt.Sprintf("-D INPUT -i %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-1", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-D OUTPUT -o %s -m mark --mark %d -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff -m comment --comment nordvpn-1", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-D OUTPUT -o %s -m connmark --mark %d -j ACCEPT -m comment --comment nordvpn-2", mock.En0Interface.Name, connmark),
	}

	tests := []struct {
		name                      string
		deviceFunc                device.ListFunc
		firewallDisabled          bool
		invalidCommand            string
		expectedAllowlistCommands []string
		expectedDenylistCommands  []string
		expectedAllowlistError    error
		expectedDenylistError     error
	}{
		{
			name:                      "success",
			deviceFunc:                getDeviceFunc(false, mock.En0Interface),
			expectedAllowlistCommands: expectedAllowlistCommandsIf0,
			expectedDenylistCommands:  expectedDenylistCommandsIf0,
		},
		{
			name:                      "success two interfaces",
			deviceFunc:                getDeviceFunc(false, mock.En0Interface, mock.En1Interface),
			expectedAllowlistCommands: append(expectedAllowlistCommandsIf0, expectedAllowlistCommandsIf1...),
			expectedDenylistCommands:  append(expectedDenylistCommandsIf0, expectedDenylistCommandsIf1...),
		},
		{
			name:                   "device list failure",
			deviceFunc:             getDeviceFunc(true),
			expectedAllowlistError: ErrGetDevicesFailed,
		},
		{
			name:                   "iptables failure when allowlisting",
			deviceFunc:             getDeviceFunc(false, mock.En0Interface),
			invalidCommand:         allowlistCommand,
			expectedAllowlistError: ErrIptablesFailure,
		},
		{
			name:                      "iptables failure when denylisting",
			deviceFunc:                getDeviceFunc(false, mock.En0Interface),
			invalidCommand:            denylistCommand,
			expectedAllowlistCommands: expectedAllowlistCommandsIf0,
			expectedDenylistError:     ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := newCommandRunnerMock()
			if test.invalidCommand != "" {
				commandRunnerMock.errCommand = test.invalidCommand
			}

			// When testing a single type rule, we do not care so much about ordering/respecting priority, so it is
			// enough to set output to empty iptables(necessary because output processing would crash the tests otherwise).
			outputChain := newIptablesOutput(OUTPUT_CHAIN_NAME)
			commandRunnerMock.addIptablesListOutput(OUTPUT_CHAIN_NAME, outputChain.get())

			inputChain := newIptablesOutput(INPUT_CHAIN_NAME)
			commandRunnerMock.addIptablesListOutput(INPUT_CHAIN_NAME, inputChain.get())

			firewallManager := NewFirewallManager(test.deviceFunc, &commandRunnerMock, connmark, true, true)

			err := firewallManager.APIAllowlist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiAllowlist.")
				return
			}

			commandsIPv4AfterApiAllowlist := commandRunnerMock.popIPv4Commands()
			commandsIPv6AfterApiAllowlist := commandRunnerMock.popIPv6Commands()
			assert.Len(t, commandsIPv4AfterApiAllowlist, len(test.expectedAllowlistCommands),
				"Invalid IPv4 commands executed after api allowlist.")
			assert.Len(t, commandsIPv6AfterApiAllowlist, len(test.expectedAllowlistCommands),
				"Invalid IPv6 commands executed after api allowlist.")
			for _, expectedCommand := range test.expectedAllowlistCommands {
				assert.Contains(t,
					commandsIPv4AfterApiAllowlist,
					expectedCommand,
					"Expected IPv4 command not found after api allowlist.\nExpected command:\n%s\nExecuted commands:\n%s",
					expectedCommand,
					transformCommandsForPrinting(t, commandsIPv4AfterApiAllowlist))
				assert.Contains(t, commandsIPv6AfterApiAllowlist, expectedCommand,
					"Expected IPv6 command not found after api allowlist.\nExpected command:\n%s\nExecuted commands:\n%s",
					expectedCommand,
					transformCommandsForPrinting(t, commandsIPv6AfterApiAllowlist))
			}

			err = firewallManager.APIDenylist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiDenylist.")
				return
			}

			commandsIPv4AfterApiDenylist := commandRunnerMock.popIPv4Commands()
			assert.Len(t, commandsIPv4AfterApiDenylist, len(test.expectedDenylistCommands),
				"Invalid IPv4 commands executed after api denylist.")
			commandsIPv6AfterApiDenylist := commandRunnerMock.popIPv6Commands()
			assert.Len(t, commandsIPv6AfterApiDenylist, len(test.expectedDenylistCommands),
				"Invalid IPv6 commands executed after api denylist.")
			for _, expectedCommand := range test.expectedDenylistCommands {
				assert.Contains(t, commandsIPv4AfterApiDenylist, expectedCommand,
					"Expected IPv4 command not found after api denylist.")
				assert.Contains(t, commandsIPv6AfterApiDenylist, expectedCommand,
					"Expected IPv6 command not found after api denylist.")
			}
		})
	}
}

func TestIptablesManager(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name            string
		rules           []string
		newRulePriority rulePriority
		expectedCommand string
	}{
		{
			name: "insert rule with lowest priority",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
			},
			newRulePriority: 0,
			expectedCommand: "-I INPUT 4 -j DROP -m comment --comment nordvpn-0",
		},
		{
			name: "insert rule with highest priority",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
			},
			newRulePriority: 4,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-4",
		},
		{
			name: "insert rule inbetween",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-4 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
			},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 2 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name:            "insert rule in empty iptables",
			rules:           []string{},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name: "insert rule no nordvpn rules",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name: "insert with highest priority non-nordvpn rules at the bottom",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 1 -j DROP -m comment --comment nordvpn-3",
		},
		{
			name: "insert with lowest priority non-nordvpn rules at the bottom",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 0,
			expectedCommand: "-I INPUT 4 -j DROP -m comment --comment nordvpn-0",
		},
		{
			name: "insert inbetween non-nordvpn rules at the bottom",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */",
				"DROP       all  --  anywhere             anywhere             /* nordvpn-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-0 */",
				"DROP       all  --  anywhere             anywhere             /* other-1 */",
				"DROP       all  --  anywhere             anywhere             /* other-2 */"},
			newRulePriority: 2,
			expectedCommand: "-I INPUT 2 -j DROP -m comment --comment nordvpn-2",
		},
		{
			name: "insert with highest priority non-nordvpn inbetween",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",   // (1)
				"DROP       all  --  anywhere             anywhere             /* other-1 */",   // (2)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */", // nordvpn (3)
				"DROP       all  --  anywhere             anywhere             /* other-2 */",   // (4)
				"DROP       all  --  anywhere             anywhere             /* other-3 */",   // (5)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */", // nordvpn (6)
				"DROP       all  --  anywhere             anywhere             /* other-4 */",   // (7)
				"DROP       all  --  anywhere             anywhere             /* other-5 */",   // (8)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-0 */", // nordvpn (9)
			},
			newRulePriority: 4,
			expectedCommand: "-I INPUT 3 -j DROP -m comment --comment nordvpn-4",
		},
		{
			name: "insert with highest priority non-nordvpn inbetween",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",   // (1)
				"DROP       all  --  anywhere             anywhere             /* other-1 */",   // (2)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-3 */", // nordvpn (3)
				"DROP       all  --  anywhere             anywhere             /* other-2 */",   // (4)
				"DROP       all  --  anywhere             anywhere             /* other-3 */",   // (5)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */", // nordvpn (6)
				"DROP       all  --  anywhere             anywhere             /* other-4 */",   // (7)
				"DROP       all  --  anywhere             anywhere             /* other-5 */",   // (8)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */", // nordvpn (9)
			},
			newRulePriority: 0,
			expectedCommand: "-I INPUT 10 -j DROP -m comment --comment nordvpn-0",
		},
		{
			name: "insert inbetween non-nordvpn inbetween",
			rules: []string{
				"DROP       all  --  anywhere             anywhere             /* other-0 */",   // (1)
				"DROP       all  --  anywhere             anywhere             /* other-1 */",   // (2)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-4 */", // nordvpn (3)
				"DROP       all  --  anywhere             anywhere             /* other-2 */",   // (4)
				"DROP       all  --  anywhere             anywhere             /* other-3 */",   // (5)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-2 */", // nordvpn (6)
				"DROP       all  --  anywhere             anywhere             /* other-4 */",   // (7)
				"DROP       all  --  anywhere             anywhere             /* other-5 */",   // (8)
				"DROP       all  --  anywhere             anywhere             /* nordvpn-1 */", // nordvpn (9)
			},
			newRulePriority: 3,
			expectedCommand: "-I INPUT 6 -j DROP -m comment --comment nordvpn-3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chain := newIptablesOutput(INPUT_CHAIN_NAME)
			chain.addRules(test.rules...)

			commandRunnerMock := newCommandRunnerMock()
			commandRunnerMock.addIptablesListOutput(INPUT_CHAIN_NAME, chain.get())

			iptablesManager := newIptablesManager(&commandRunnerMock, true, true)
			// nolint:errcheck // Tested in other uts
			iptablesManager.insertRule(NewFwRule(INPUT, IPv4, "-j DROP", test.newRulePriority))

			commands := commandRunnerMock.popIPv4Commands()
			assert.Len(t, commands, 1, "Only one command per rule insertion should be executed.")
			assert.Equal(t, test.expectedCommand, commands[0], "Invalid command executed when inserting a rule.")
		})
	}
}
