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
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	iptablesmock "github.com/NordSecurity/nordvpn-linux/test/mock/firewall/iptables_manager"
	"github.com/stretchr/testify/assert"
)

var ErrGetDevicesFailed error = errors.New("get devices has failed")

const (
	peerPublicKey = "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet="
	peerIPAddress = "48.242.30.25"
)

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
			expectedErrBlock: iptablesmock.ErrIptablesFailure,
		},
		{
			name:                          "unblock failure",
			devicesFunc:                   getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterBlocking: iface0CommandsAfterBlocking,
			errCommand:                    iface0DeleteInputCommand,
			expectedErrUnblock:            iptablesmock.ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()

			if test.errCommand != "" {
				commandRunnerMock.ErrCommand = test.errCommand
			}

			firewallManager := NewFirewallManager(test.devicesFunc, &commandRunnerMock, connmark, true)

			err := firewallManager.BlockTraffic()

			if test.expectedErrBlock != nil {
				assert.ErrorIs(t, err, test.expectedErrBlock, "Unexpected error returned after block has failed.")
				return
			}

			commandsIPv4 := commandRunnerMock.PopIPv4Commands()
			expectedNumberOfCommands := len(test.expectedCommandsAfterBlocking)
			assert.Len(t,
				commandsIPv4,
				expectedNumberOfCommands,
				"Invalid number of IPv4 commands when blocking traffic. Commands:\n%s",
				transformCommandsForPrinting(t, commandsIPv4))

			// rules are added to two different chains, so ordering doesn't matter in this case and we can use Contains
			for _, expectedCommand := range test.expectedCommandsAfterBlocking {
				assert.Contains(t,
					commandsIPv4,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")
			}

			err = firewallManager.UnblockTraffic()

			if test.expectedErrUnblock != nil {
				assert.ErrorIs(t, err, test.expectedErrUnblock, "Unexpected error returned after unblock has failed.")
				return
			}

			commandsIPv4 = commandRunnerMock.PopIPv4Commands()
			expectedNumberOfCommands = len(test.expectedCommandsAfterUnblocking)
			assert.Len(t, commandsIPv4, expectedNumberOfCommands, "Invalid number of commands when unblocking traffic.")

			for _, expectedCommand := range test.expectedCommandsAfterUnblocking {
				assert.Contains(t,
					commandsIPv4,
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

	commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
	firewallManager := NewFirewallManager(getDeviceFunc(false, mock.En0Interface), &commandRunnerMock, connmark, true)

	err := firewallManager.BlockTraffic()
	assert.Nil(t, err, "Received unexpected error when blocking traffic.")

	commands := commandRunnerMock.PopIPv4Commands()
	assert.Equal(t, iface0CommandsAfterBlocking, commands, "Invalid commands executed when blocking traffic.")

	err = firewallManager.BlockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received after blocking traffic a second time.")

	commands = commandRunnerMock.PopIPv4Commands()
	assert.Empty(t, commands, "Commands were executed after blocking traffic for a second time.")
}

func TestUnblockTraffic_TrafficNotBlocked(t *testing.T) {
	category.Set(t, category.Unit)

	commandRunnerMock := iptablesmock.NewCommandRunnerMock()
	firewallManager := NewFirewallManager(getDeviceFunc(false), &commandRunnerMock, connmark, true)

	err := firewallManager.UnblockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received when unblocking traffic when it was not blocked.")

	commands := commandRunnerMock.PopIPv4Commands()
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
			expectedErrSet: iptablesmock.ErrIptablesFailure,
		},
		{
			name:                     "iptables failure when unsetting",
			invalidCommand:           fmt.Sprintf("-D INPUT -s 102.56.52.223/22 -i %s -j ACCEPT -m comment --comment nordvpn-3", mock.En0Interface.Name),
			deviceFunc:               getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterSet: expectedCommandsIface0,
			expectedErrUnset:         iptablesmock.ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
			if test.invalidCommand != "" {
				commandRunnerMock.ErrCommand = test.invalidCommand
			}

			firewallManager := NewFirewallManager(test.deviceFunc, &commandRunnerMock, connmark, true)

			err := firewallManager.SetAllowlist(udpPorts, tcpPorts, subnets)
			assert.ErrorIs(t, err, test.expectedErrSet, "Invalid error returned by SetAllowlist.")

			if test.expectedErrSet != nil {
				return
			}

			commandsAfterSet := commandRunnerMock.PopIPv4Commands()
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
			commandsAfterUnset := commandRunnerMock.PopIPv4Commands()
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
			expectedAllowlistError: iptablesmock.ErrIptablesFailure,
		},
		{
			name:                      "iptables failure when denylisting",
			deviceFunc:                getDeviceFunc(false, mock.En0Interface),
			invalidCommand:            denylistCommand,
			expectedAllowlistCommands: expectedAllowlistCommandsIf0,
			expectedDenylistError:     iptablesmock.ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
			if test.invalidCommand != "" {
				commandRunnerMock.ErrCommand = test.invalidCommand
			}

			// When testing a single type rule, we do not care so much about ordering/respecting priority, so it is
			// enough to set output to empty iptables(necessary because output processing would crash the tests otherwise).
			outputChain := iptablesmock.NewIptablesOutput(iptablesmock.OutputChainName)
			commandRunnerMock.AddIptablesListOutput(iptablesmock.OutputChainName, outputChain.Get())

			inputChain := iptablesmock.NewIptablesOutput(iptablesmock.InputChainName)
			commandRunnerMock.AddIptablesListOutput(iptablesmock.InputChainName, inputChain.Get())

			firewallManager := NewFirewallManager(test.deviceFunc, &commandRunnerMock, connmark, true)

			err := firewallManager.APIAllowlist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiAllowlist.")
				return
			}

			commandsIPv4AfterApiAllowlist := commandRunnerMock.PopIPv4Commands()
			assert.Len(t, commandsIPv4AfterApiAllowlist, len(test.expectedAllowlistCommands),
				"Invalid IPv4 commands executed after api allowlist.")
			for _, expectedCommand := range test.expectedAllowlistCommands {
				assert.Contains(t,
					commandsIPv4AfterApiAllowlist,
					expectedCommand,
					"Expected IPv4 command not found after api allowlist.\nExpected command:\n%s\nExecuted commands:\n%s",
					expectedCommand,
					transformCommandsForPrinting(t, commandsIPv4AfterApiAllowlist))
			}

			err = firewallManager.APIDenylist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiDenylist.")
				return
			}

			commandsIPv4AfterApiDenylist := commandRunnerMock.PopIPv4Commands()
			assert.Len(t, commandsIPv4AfterApiDenylist, len(test.expectedDenylistCommands),
				"Invalid IPv4 commands executed after api denylist.")
			for _, expectedCommand := range test.expectedDenylistCommands {
				assert.Contains(t, commandsIPv4AfterApiDenylist, expectedCommand,
					"Expected IPv4 command not found after api denylist.")
			}
		})
	}
}

func TestAllowDenyFileshare(t *testing.T) {
	peerPublicKey := peerPublicKey
	peerIPAddress := "48.242.30.25"
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	allowFileshareCommand := fmt.Sprintf(
		"-I INPUT 1 -s %s/32 -p tcp -m tcp --dport 49111 -j ACCEPT -m comment --comment nordvpn-4",
		peerIPAddress)
	denyFileshareCommand := fmt.Sprintf(
		"-D INPUT -s %s/32 -p tcp -m tcp --dport 49111 -j ACCEPT -m comment --comment nordvpn-4",
		peerIPAddress)

	tests := []struct {
		name                       string
		invalidCommand             string
		firewallDisabled           bool
		expectedCommandsAfterAllow []string
		expectedCommandsAfterDeny  []string
		expectedAllowErr           error
		expectedDenyErr            error
	}{
		{
			name:                       "success",
			expectedCommandsAfterAllow: []string{allowFileshareCommand},
			expectedCommandsAfterDeny:  []string{denyFileshareCommand},
			expectedAllowErr:           nil,
		},
		{
			name:             "allow fails",
			invalidCommand:   allowFileshareCommand,
			expectedAllowErr: iptablesmock.ErrIptablesFailure,
		},
		{
			name:                       "deny fileshare",
			invalidCommand:             denyFileshareCommand,
			expectedCommandsAfterAllow: []string{allowFileshareCommand},
			expectedDenyErr:            iptablesmock.ErrIptablesFailure,
		},
		{
			name:             "firewall disabled",
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
			if test.invalidCommand != "" {
				commandRunnerMock.ErrCommand = test.invalidCommand
			}
			firewallManager := NewFirewallManager(nil, &commandRunnerMock, connmark, !test.firewallDisabled)

			err := firewallManager.AllowFileshare(peerAddress)
			if test.expectedAllowErr != nil {
				assert.ErrorIs(t,
					err,
					test.expectedAllowErr,
					"Unexpected error returned by FirewallManager: %w", err)
				return
			}

			commandsAfterAllow := commandRunnerMock.PopIPv4Commands()
			assert.Len(t,
				commandsAfterAllow,
				len(test.expectedCommandsAfterAllow),
				"Unexpected commands executed after allowing fileshare")
			assert.Equal(t,
				test.expectedCommandsAfterAllow,
				commandsAfterAllow,
				"Fileshare allow rule was not added to iptables.")

			err = firewallManager.DenyFileshare(peerPublicKey)
			if test.expectedDenyErr != nil {
				assert.ErrorIs(t,
					err,
					test.expectedDenyErr,
					"Unexpected error returned by FirewallManager: %w", err)
				return
			}

			commandsAfterDeny := commandRunnerMock.PopIPv4Commands()
			assert.Len(t,
				commandsAfterDeny,
				len(test.expectedCommandsAfterDeny),
				"Unexpected commands executed denying fileshare.")
			assert.Equal(t,
				test.expectedCommandsAfterDeny,
				commandsAfterDeny,
				"Fileshare deny rule was not added to iptables.")
		})
	}
}

func TestAllowDenyIncoming(t *testing.T) {
	peerPublicKey := peerPublicKey
	peerIPAddress := "48.242.30.25"
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	allowCommand := fmt.Sprintf("-I INPUT 1 -s %s/32 -j ACCEPT -m comment --comment nordvpn-5", peerIPAddress)
	blockLANCommands := []string{
		fmt.Sprintf("-I INPUT 1 -s %s/32 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
		fmt.Sprintf("-I INPUT 1 -s %s/32 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
		fmt.Sprintf("-I INPUT 1 -s %s/32 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
		fmt.Sprintf("-I INPUT 1 -s %s/32 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
	}

	denyCommand := fmt.Sprintf("-D INPUT -s %s/32 -j ACCEPT -m comment --comment nordvpn-5", peerIPAddress)
	unblockLANCommands := []string{
		fmt.Sprintf("-D INPUT -s %s/32 -d 169.254.0.0/16 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
		fmt.Sprintf("-D INPUT -s %s/32 -d 192.168.0.0/16 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
		fmt.Sprintf("-D INPUT -s %s/32 -d 172.16.0.0/12 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
		fmt.Sprintf("-D INPUT -s %s/32 -d 10.0.0.0/8 -j DROP -m comment --comment nordvpn-6", peerIPAddress),
	}

	tests := []struct {
		name                       string
		lanAllowed                 bool
		firewallDisabled           bool
		expectedCommandsAfterAllow []string
		expectedCommandsAfterDeny  []string
		invalidCommand             string
		expectedAllowErr           error
		expectedDenyErr            error
	}{
		{
			name:                       "success",
			lanAllowed:                 true,
			expectedCommandsAfterAllow: []string{allowCommand},
			expectedCommandsAfterDeny:  []string{denyCommand},
		},
		{
			name:       "success lan blocked",
			lanAllowed: false,
			// commands in order of execution, i.e allow incoming first, then block all LANs
			expectedCommandsAfterAllow: append(blockLANCommands, allowCommand),
			expectedCommandsAfterDeny:  append([]string{denyCommand}, unblockLANCommands...),
		},
		{
			name:             "failure when allowing",
			lanAllowed:       true,
			invalidCommand:   allowCommand,
			expectedAllowErr: iptablesmock.ErrIptablesFailure,
		},
		{
			name:                       "failure when denying",
			lanAllowed:                 true,
			invalidCommand:             denyCommand,
			expectedCommandsAfterAllow: []string{allowCommand},
			expectedDenyErr:            iptablesmock.ErrIptablesFailure,
		},
		{
			name:             "firewall disabled",
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
			if test.invalidCommand != "" {
				commandRunnerMock.ErrCommand = test.invalidCommand
			}

			firewallManager := NewFirewallManager(nil, &commandRunnerMock, connmark, !test.firewallDisabled)

			err := firewallManager.AllowIncoming(peerAddress, test.lanAllowed)
			if test.expectedAllowErr != nil {
				assert.ErrorIs(t, err, test.expectedAllowErr, "Invalid error returned by AllowIncoming.")
				return
			}

			commandsAfterAllow := commandRunnerMock.PopIPv4Commands()
			assert.Equal(t, test.expectedCommandsAfterAllow, commandsAfterAllow,
				"Invalid commands executed when allowing incoming mesh traffic.")

			err = firewallManager.DenyIncoming(peerPublicKey)
			if test.expectedDenyErr != nil {
				assert.ErrorIs(t, err, test.expectedDenyErr, "Invalid error returned by DenyIncoming.")
				return
			}

			commandsAfterDeny := commandRunnerMock.PopIPv4Commands()
			assert.Equal(t, test.expectedCommandsAfterDeny, commandsAfterDeny,
				"Invalid commands executed when denying incoming mesh traffic.")
		})
	}
}

func TestAllowIncoming_AleradyAllowed(t *testing.T) {
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
	firewallManager := NewFirewallManager(nil, &commandRunnerMock, connmark, true)

	err := firewallManager.AllowIncoming(peerAddress, true)
	assert.Nil(t, err, "AllowIncoming has returned an unexpected error.")

	// remove commands form initial call from the mock
	commandRunnerMock.PopIPv4Commands()

	err = firewallManager.AllowIncoming(peerAddress, true)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error returned on subsequent AllowIncoming.")
	assert.Empty(t, commandRunnerMock.PopIPv4Commands(), "Commands executed after allowing incoming traffic for a second time")

	// rule duplication should be based on peers public key, so it should be detected even if the address has changed
	peerAddress.Address = netip.MustParseAddr("128.236.166.204")
	err = firewallManager.AllowIncoming(peerAddress, true)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error returned on subsequent AllowIncoming.")
	assert.Empty(t, commandRunnerMock.PopIPv4Commands(), "Commands executed after allowing incoming traffic for a second time")
}

func TestDenyIncoming_NotDenied(t *testing.T) {
	commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()
	firewallManager := NewFirewallManager(nil, &commandRunnerMock, connmark, true)

	err := firewallManager.DenyIncoming(peerPublicKey)
	assert.ErrorIs(t, err, ErrRuleNotFound)
	assert.Empty(t, commandRunnerMock.PopIPv4Commands(), "Commands executed after denying mesh traffic that was not allowed.")
}

func TestAllowFileshare_AlreadyAllowed(t *testing.T) {
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()

	firewallManager := NewFirewallManager(getDeviceFunc(false), &commandRunnerMock, connmark, true)

	err := firewallManager.AllowFileshare(peerAddress)
	assert.Nil(t, err, "Unexpected error when allowing fileshare: %w", err)

	commands := commandRunnerMock.PopIPv4Commands()
	assert.Len(t, commands, 1, "Unexpected commands executed when allowing fileshare.")

	err = firewallManager.AllowFileshare(peerAddress)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error received when allowing fileshare when it was already allowed.")
}

func TestDenyFileshare_NotAllowed(t *testing.T) {
	commandRunnerMock := iptablesmock.NewCommandRunnerMockWithTables()

	firewallManager := NewFirewallManager(getDeviceFunc(false), &commandRunnerMock, connmark, true)

	err := firewallManager.DenyFileshare(peerPublicKey)
	assert.ErrorIs(t, err, ErrRuleNotActive,
		"Invalid error received when denying fileshare when it was not previously allowed.")

	commands := commandRunnerMock.PopIPv4Commands()
	assert.Empty(t, commands,
		"Commands were executed when denying fileshare when it was not previously allowed.")
}
