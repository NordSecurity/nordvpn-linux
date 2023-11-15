package firewall

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

var ErrIptablesFailure = errors.New("iptables failure")

type IptablesMock struct {
	isErr       bool
	Commands    []string
	errCommands map[string]bool
}

func NewIptablesMock(isErr bool) IptablesMock {
	return IptablesMock{
		isErr:       isErr,
		errCommands: make(map[string]bool),
	}
}

// AddErrCommand adds an error command. Subsequent calls to ExecuteCommand with the command will return
// ErrIptablesFailure.
func (i *IptablesMock) AddErrCommand(command string) {
	i.errCommands[command] = true
}

// popCommands returns recorded commands executed by the mock and clears the internal state.
func (i *IptablesMock) popCommands() []string {
	commands := i.Commands
	i.Commands = nil
	return commands
}

func (i *IptablesMock) executeCommand(command string) error {
	if i.isErr {
		return ErrIptablesFailure
	}

	if _, ok := i.errCommands[command]; ok {
		return ErrIptablesFailure
	}

	i.Commands = append(i.Commands, command)

	return nil
}

func (i *IptablesMock) InsertRule(command string) error {
	return i.executeCommand("-I " + command)
}

func (i *IptablesMock) DeleteRule(command string) error {
	return i.executeCommand("-D " + command)
}

func (i *IptablesMock) InsertRuleIPv6(command string) error {
	return nil
}

func (i *IptablesMock) DeleteRuleIPv6(command string) error {
	return nil
}

func (i *IptablesMock) Enable() {}

func (i *IptablesMock) Disable() {}

var ErrGetDevicesFailed error = errors.New("get devices has failed")

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

	newCommands := []string{}
	for _, command := range oldCommands {
		newCommands = append(newCommands, strings.Replace(command, "-I", "-D", 1))
	}
	return newCommands
}

const connmark uint32 = 0x55

func TestTrafficBlocking(t *testing.T) {
	category.Set(t, category.Unit)

	iface0InsertInputCommand := fmt.Sprintf("-I INPUT -i %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name)
	iface0CommandsAfterBlocking := []string{
		iface0InsertInputCommand,
		fmt.Sprintf("-I OUTPUT -o %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name),
	}

	iface1CommandsAfterBlocking := []string{
		fmt.Sprintf("-I INPUT -i %s -m comment --comment nordvpn -j DROP", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -m comment --comment nordvpn -j DROP", mock.En1Interface.Name),
	}

	iface0DeleteInputCommand := fmt.Sprintf("-D INPUT -i %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name)
	iface0CommandsAfterUnblocking := []string{
		iface0DeleteInputCommand,
		fmt.Sprintf("-D OUTPUT -o %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name),
	}

	iface1CommandsAfterUnblocking := []string{
		fmt.Sprintf("-D INPUT -i %s -m comment --comment nordvpn -j DROP", mock.En1Interface.Name),
		fmt.Sprintf("-D OUTPUT -o %s -m comment --comment nordvpn -j DROP", mock.En1Interface.Name),
	}

	tests := []struct {
		name                            string
		devicesFunc                     device.ListFunc
		devicesFuncUnblock              device.ListFunc
		failingCommand                  string
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
			failingCommand:   iface0InsertInputCommand,
			expectedErrBlock: ErrIptablesFailure,
		},
		{
			name:                          "unblock failure",
			devicesFunc:                   getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterBlocking: iface0CommandsAfterBlocking,
			failingCommand:                iface0DeleteInputCommand,
			expectedErrUnblock:            ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)

			if test.failingCommand != "" {
				iptablesMock.AddErrCommand(test.failingCommand)
			}

			firewallManager := NewFirewallManager(test.devicesFunc, &iptablesMock, connmark)

			err := firewallManager.BlockTraffic()

			if test.expectedErrBlock != nil {
				assert.ErrorIs(t, err, test.expectedErrBlock, "Unexpected error returned after block has failed.")
				return
			}

			commands := iptablesMock.popCommands()
			expectedNumberOfCommands := len(test.expectedCommandsAfterBlocking)
			assert.Len(t, commands, expectedNumberOfCommands, "Invalid number of commands when blocking traffic.")

			// rules are added to two different chains, so ordering doesn't matter in this case and we can use Contains
			for _, expectedCommand := range test.expectedCommandsAfterBlocking {
				assert.Contains(t,
					commands,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")

				assert.Contains(t,
					commands,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")
			}

			err = firewallManager.UnblockTraffic()

			if test.expectedErrUnblock != nil {
				assert.ErrorIs(t, err, test.expectedErrUnblock, "Unexpected error returned after unblock has failed.")
				return
			}

			commands = iptablesMock.popCommands()
			expectedNumberOfCommands = len(test.expectedCommandsAfterUnblocking)
			assert.Len(t, commands, expectedNumberOfCommands, "Invalid number of commands when unblocking traffic.")

			for _, expectedCommand := range test.expectedCommandsAfterUnblocking {
				assert.Contains(t,
					commands,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")

				assert.Contains(t,
					commands,
					expectedCommand,
					"Input block traffic rule was not added to the firewall.")
			}
		})
	}
}

func TestBlockTraffic_AlreadyBlocked(t *testing.T) {
	category.Set(t, category.Unit)

	iface0CommandsAfterBlocking := []string{
		fmt.Sprintf("-I INPUT -i %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name),
	}

	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false, mock.En0Interface), &iptablesMock, connmark)

	err := firewallManager.BlockTraffic()
	assert.Nil(t, err, "Received unexpected error when blocking traffic.")

	commands := iptablesMock.popCommands()
	assert.Equal(t, iface0CommandsAfterBlocking, commands, "Invalid commands executed when blocking traffic.")

	err = firewallManager.BlockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received after blocking traffic a second time.")

	commands = iptablesMock.popCommands()
	assert.Empty(t, commands, "Commands were executed after blocking traffic for a second time.")
}

func TestUnblockTraffic_TrafficNotBlocked(t *testing.T) {
	category.Set(t, category.Unit)

	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false), &iptablesMock, connmark)

	err := firewallManager.UnblockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received when unblocking traffic when it was not blocked.")

	commands := iptablesMock.popCommands()
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
		fmt.Sprintf("-I INPUT -s 102.56.52.223/22 -i %s -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -d 102.56.52.223/22 -o %s -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --dport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --sport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --dport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --sport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --dport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --sport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --dport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --sport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --dport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --sport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --dport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --sport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --dport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --sport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --dport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --sport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
	}

	expectedCommandsIface1 := []string{
		fmt.Sprintf("-I INPUT -s 102.56.52.223/22 -i %s -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -d 102.56.52.223/22 -o %s -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --dport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --sport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --dport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --sport 30000:30002 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --dport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p udp -m udp --sport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --dport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p udp -m udp --sport 40000:40000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --dport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --sport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --dport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --sport 50002:50004 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --dport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I INPUT -i %s -p tcp -m tcp --sport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --dport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -p tcp -m tcp --sport 60000:60000 -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name),
	}

	tests := []struct {
		name                     string
		deviceFunc               device.ListFunc
		firewallDisabled         bool
		expectedCommandsAfterSet []string
		invalidCommand           string
		failingCommand           string
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
			invalidCommand: fmt.Sprintf("-I INPUT -s 102.56.52.223/22 -i %s -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
			deviceFunc:     getDeviceFunc(false, mock.En0Interface),
			expectedErrSet: ErrIptablesFailure,
		},
		{
			name:                     "iptables failure when unsetting",
			invalidCommand:           fmt.Sprintf("-D INPUT -s 102.56.52.223/22 -i %s -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name),
			deviceFunc:               getDeviceFunc(false, mock.En0Interface),
			expectedCommandsAfterSet: expectedCommandsIface0,
			expectedErrUnset:         ErrIptablesFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(test.deviceFunc, &iptablesMock, connmark)

			err := firewallManager.SetAllowlist(udpPorts, tcpPorts, subnets)
			if test.expectedErrSet != nil {
				assert.ErrorIs(t, err, test.expectedErrSet, "Invalid error returned by SetAllowlist.")
				return
			}

			commandsAfterSet := iptablesMock.popCommands()
			assert.Len(t,
				commandsAfterSet,
				len(test.expectedCommandsAfterSet),
				"Invalid commands executed after setting allowlist.\nExpected:\n%s,\nGot:\n%s",
				transformCommandsForPrinting(t, test.expectedCommandsAfterSet),
				transformCommandsForPrinting(t, commandsAfterSet))
			for _, expectedCommands := range test.expectedCommandsAfterSet {
				assert.Contains(t, commandsAfterSet, expectedCommands,
					"Expected command not executed after setting allowlist.")
			}

			err = firewallManager.UnsetAllowlist()
			if test.expectedErrUnset != nil {
				assert.ErrorIs(t, err, test.expectedErrUnset, "Invalid error returned by UnsetAllowlist.")
				return
			}

			// same commands should be performed, just with -D flag instead of -I flag
			expectedCommandsAfterUnset := transformCommandsToDelte(t, test.expectedCommandsAfterSet)
			commandsAfterUnset := iptablesMock.popCommands()
			assert.Len(t,
				commandsAfterUnset,
				len(expectedCommandsAfterUnset),
				"Invalid commands executed after unseting allowlist.\nExpected:\n%s\nGot:\n%s",
				transformCommandsForPrinting(t, expectedCommandsAfterUnset),
				transformCommandsForPrinting(t, commandsAfterSet))
			for _, expectedCommand := range transformCommandsToDelte(t, test.expectedCommandsAfterSet) {
				assert.Contains(t, commandsAfterUnset, expectedCommand,
					"Expected command not executed after unsetting the allowlist.")
			}
		})
	}
}

func TestApiAllowlist(t *testing.T) {
	allowlistCommand := fmt.Sprintf("-I INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name, connmark)
	expectedAllowlistCommandsIf0 := []string{
		allowlistCommand,
		fmt.Sprintf("-I OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-I OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name, connmark),
	}

	denylistCommand := fmt.Sprintf("-D INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name, connmark)
	expectedDenylistCommandsIf0 := []string{
		denylistCommand,
		fmt.Sprintf("-D OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-D OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name, connmark),
	}

	expectedAllowlistCommandsIf1 := []string{
		fmt.Sprintf("-I INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name, connmark),
		fmt.Sprintf("-I OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff", mock.En1Interface.Name, connmark),
		fmt.Sprintf("-I OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En1Interface.Name, connmark),
	}

	expectedDenylistCommandsIf1 := []string{
		fmt.Sprintf("-D INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-D OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff", mock.En0Interface.Name, connmark),
		fmt.Sprintf("-D OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", mock.En0Interface.Name, connmark),
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
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(test.deviceFunc, &iptablesMock, connmark)

			err := firewallManager.APIAllowlist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiAllowlist.")
				return
			}

			commandsAfterApiAllowlist := iptablesMock.popCommands()
			assert.Len(t, commandsAfterApiAllowlist, len(test.expectedAllowlistCommands),
				"Invalid commands executed after api allowlist.")
			for _, expectedCommand := range test.expectedAllowlistCommands {
				assert.Contains(t, commandsAfterApiAllowlist, expectedCommand,
					"Expected command not found after api allowlist.")
			}

			err = firewallManager.APIDenylist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiDenylist.")
				return
			}

			commandsAfterApiDenylist := iptablesMock.popCommands()
			assert.Len(t, commandsAfterApiDenylist, len(test.expectedDenylistCommands),
				"Invalid commands executed after api denylist.")
			for _, expectedCommand := range test.expectedDenylistCommands {
				assert.Contains(t, commandsAfterApiDenylist, expectedCommand,
					"Expected command not found after api denylist.")
			}
		})
	}
}
