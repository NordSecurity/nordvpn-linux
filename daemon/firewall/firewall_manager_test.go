package firewall

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
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

// PopCommands returns recorded commands executed by the mock and clears the internal state.
func (i *IptablesMock) PopCommands() []string {
	commands := i.Commands
	i.Commands = nil
	return commands
}

func (i *IptablesMock) ExecuteCommand(command string) error {
	if i.isErr {
		return ErrIptablesFailure
	}

	if _, ok := i.errCommands[command]; ok {
		return ErrIptablesFailure
	}

	i.Commands = append(i.Commands, command)

	return nil
}

func (i *IptablesMock) ExecuteCommandIPv6(command string) error {
	return nil
}

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
		firewallDisabled                bool
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
		{
			name:             "firewall is disabled",
			devicesFunc:      getDeviceFunc(false, mock.En0Interface),
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)

			if test.failingCommand != "" {
				iptablesMock.AddErrCommand(test.failingCommand)
			}

			firewallManager := NewFirewallManager(test.devicesFunc, &iptablesMock, connmark, !test.firewallDisabled)

			err := firewallManager.BlockTraffic()

			if test.expectedErrBlock != nil {
				assert.ErrorIs(t, err, test.expectedErrBlock, "Unexpected error returned after block has failed.")
				return
			}

			commands := iptablesMock.PopCommands()
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

			commands = iptablesMock.PopCommands()
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
	iface0CommandsAfterBlocking := []string{
		fmt.Sprintf("-I INPUT -i %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name),
		fmt.Sprintf("-I OUTPUT -o %s -m comment --comment nordvpn -j DROP", mock.En0Interface.Name),
	}

	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false, mock.En0Interface), &iptablesMock, connmark, true)

	err := firewallManager.BlockTraffic()
	assert.Nil(t, err, "Received unexpected error when blocking traffic.")

	commands := iptablesMock.PopCommands()
	assert.Equal(t, iface0CommandsAfterBlocking, commands, "Invalid commands executed when blocking traffic.")

	err = firewallManager.BlockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received after blocking traffic a second time.")

	commands = iptablesMock.PopCommands()
	assert.Empty(t, commands, "Commands were executed after blocking traffic for a second time.")
}

func TestUnblockTraffic_TrafficNotBlocked(t *testing.T) {
	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false), &iptablesMock, connmark, true)

	err := firewallManager.UnblockTraffic()
	assert.ErrorIs(t, err, ErrRuleAlreadyActive, "Invalid error received when unblocking traffic when it was not blocked.")

	commands := iptablesMock.PopCommands()
	assert.Empty(t, commands, "Commands were executed when ublocking traffic when it was not blocked.")
}

func TestAllowDenyFileshare(t *testing.T) {
	peerPublicKey := "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet="
	peerIPAddress := "48.242.30.25"
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	allowFileshareCommand := fmt.Sprintf(
		"-I INPUT -s %s/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT",
		peerIPAddress)
	denyFileshareCommand := fmt.Sprintf(
		"-D INPUT -s %s/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT",
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
			expectedAllowErr: ErrIptablesFailure,
		},
		{
			name:                       "deny fileshare",
			invalidCommand:             denyFileshareCommand,
			expectedCommandsAfterAllow: []string{allowFileshareCommand},
			expectedDenyErr:            ErrIptablesFailure,
		},
		{
			name:             "firewall disabled",
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, !test.firewallDisabled)

			err := firewallManager.AllowFileshare(peerAddress)
			if test.expectedAllowErr != nil {
				assert.ErrorIs(t,
					err,
					test.expectedAllowErr,
					"Unexpected error returned by FirewallManager: %w", err)
				return
			}

			commandsAfterAllow := iptablesMock.PopCommands()
			assert.Len(t,
				commandsAfterAllow,
				len(test.expectedCommandsAfterAllow),
				"Unexpected executed after allowing fileshare")
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

			commandsAfterDeny := iptablesMock.PopCommands()
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

func TestAllowFileshare_AlreadyAllowed(t *testing.T) {
	peerPublicKey := "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet="
	peerIPAddress := "48.242.30.25"
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false), &iptablesMock, connmark, true)

	err := firewallManager.AllowFileshare(peerAddress)
	assert.Nil(t, err, "Unexpected error when allowing fileshare: %w", err)

	commands := iptablesMock.PopCommands()
	assert.Len(t, commands, 1, "Unexpected commands executed when allowing fileshare.")

	err = firewallManager.AllowFileshare(peerAddress)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error received when allowing fileshare when it was allready allowed.")
}

func TestDenyFileshare_NotAllowed(t *testing.T) {
	peerPublicKey := "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet="
	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false), &iptablesMock, connmark, true)

	err := firewallManager.DenyFileshare(peerPublicKey)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error received when denying fileshare when it was not previously allowed.")

	commands := iptablesMock.PopCommands()
	assert.Empty(t, commands,
		"Commands were executed when denying fileshare when it was not previously allowed.")
}

func TestAllowDenyIncoming(t *testing.T) {
	peerPublicKey := "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet="
	peerIPAddress := "48.242.30.25"
	peerAddress := meshnet.UniqueAddress{
		UID:     peerPublicKey,
		Address: netip.MustParseAddr(peerIPAddress),
	}

	allowCommand := fmt.Sprintf("-I INPUT -s %s/32 -m comment --comment nordvpn -j ACCEPT", peerIPAddress)
	blockLANCommands := []string{
		fmt.Sprintf("-I INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", peerIPAddress),
		fmt.Sprintf("-I INPUT -s %s/32 -d 192.168.0.0/16 -m comment --comment nordvpn -j DROP", peerIPAddress),
		fmt.Sprintf("-I INPUT -s %s/32 -d 172.16.0.0/12 -m comment --comment nordvpn -j DROP", peerIPAddress),
		fmt.Sprintf("-I INPUT -s %s/32 -d 10.0.0.0/8 -m comment --comment nordvpn -j DROP", peerIPAddress),
	}

	denyCommand := fmt.Sprintf("-D INPUT -s %s/32 -m comment --comment nordvpn -j ACCEPT", peerIPAddress)
	unblockLANCommands := []string{
		fmt.Sprintf("-D INPUT -s %s/32 -d 169.254.0.0/16 -m comment --comment nordvpn -j DROP", peerIPAddress),
		fmt.Sprintf("-D INPUT -s %s/32 -d 192.168.0.0/16 -m comment --comment nordvpn -j DROP", peerIPAddress),
		fmt.Sprintf("-D INPUT -s %s/32 -d 172.16.0.0/12 -m comment --comment nordvpn -j DROP", peerIPAddress),
		fmt.Sprintf("-D INPUT -s %s/32 -d 10.0.0.0/8 -m comment --comment nordvpn -j DROP", peerIPAddress),
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
			expectedCommandsAfterAllow: append([]string{allowCommand}, blockLANCommands...),
			expectedCommandsAfterDeny:  append([]string{denyCommand}, unblockLANCommands...),
		},
		{
			name:             "failure when allowing",
			lanAllowed:       true,
			invalidCommand:   allowCommand,
			expectedAllowErr: ErrIptablesFailure,
		},
		{
			name:                       "failure when denying",
			lanAllowed:                 true,
			invalidCommand:             denyCommand,
			expectedCommandsAfterAllow: []string{allowCommand},
			expectedDenyErr:            ErrIptablesFailure,
		},
		{
			name:             "firewall disabled",
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, !test.firewallDisabled)

			err := firewallManager.AllowIncoming(peerAddress, test.lanAllowed)
			if test.expectedAllowErr != nil {
				assert.ErrorIs(t, err, test.expectedAllowErr, "Invalid error returned by AllowIncoming.")
				return
			}

			commandsAfterAllow := iptablesMock.PopCommands()
			assert.Equal(t, test.expectedCommandsAfterAllow, commandsAfterAllow,
				"Invalid commands executed when allowing incoming mesh traffic.")

			err = firewallManager.DenyIncoming(peerPublicKey)
			if test.expectedDenyErr != nil {
				assert.ErrorIs(t, err, test.expectedDenyErr, "Invalid error returned by DenyIncoming.")
				return
			}

			commandsAfterDeny := iptablesMock.PopCommands()
			assert.Equal(t, test.expectedCommandsAfterDeny, commandsAfterDeny,
				"Invalid commands executed when denying incoming mesh traffic.")
		})
	}
}

func TestAllowIncoming_AleradyAllowed(t *testing.T) {
	peerAddress := meshnet.UniqueAddress{
		UID:     "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet=",
		Address: netip.MustParseAddr("48.242.30.25"),
	}

	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, true)

	err := firewallManager.AllowIncoming(peerAddress, true)
	assert.Nil(t, err, "AllowIncoming has returned an unexpected error.")

	// remove commands form initial call from the mock
	iptablesMock.PopCommands()

	err = firewallManager.AllowIncoming(peerAddress, true)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error returned on subsequent AllowIncoming.")
	assert.Empty(t, iptablesMock.PopCommands(), "Commands executed after allowing incoming traffic for a second time")

	// rule duplication should be based on peers public key, so it should be detected even if the address has changed
	peerAddress.Address = netip.MustParseAddr("128.236.166.204")
	err = firewallManager.AllowIncoming(peerAddress, true)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive,
		"Invalid error returned on subsequent AllowIncoming.")
	assert.Empty(t, iptablesMock.PopCommands(), "Commands executed after allowing incoming traffic for a second time")
}

func TestDenyIncoming_NotDenied(t *testing.T) {
	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, true)

	err := firewallManager.DenyIncoming("D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet=")
	assert.ErrorIs(t, err, ErrRuleAlreadyActive)
	assert.Empty(t, iptablesMock.PopCommands(), "Commands executed after denying mesh traffic that was not allowed.")
}

func TestBlockUnblockMeshnet(t *testing.T) {
	deviceAddress := "48.242.30.25"

	blockUnrelatedTrafficCommand := fmt.Sprintf("-I INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc %s -m comment --comment nordvpn -j ACCEPT", deviceAddress)
	blockCommands := []string{
		blockUnrelatedTrafficCommand,
		"-I INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP",
	}

	unblockUnrelatedTrafficCommand := fmt.Sprintf("-D INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc %s -m comment --comment nordvpn -j ACCEPT", deviceAddress)
	unblockCommands := []string{
		unblockUnrelatedTrafficCommand,
		"-D INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP",
	}

	tests := []struct {
		name                         string
		firewallDisabled             bool
		expectedCommandsAfterBlock   []string
		expectedCommandsAfterUnblock []string
		invalidCommand               string
		expectedBlockErr             error
		expectedUnblockErr           error
	}{
		{
			name:                         "success",
			expectedCommandsAfterBlock:   blockCommands,
			expectedCommandsAfterUnblock: unblockCommands,
		},
		{
			name:             "block failure",
			invalidCommand:   blockUnrelatedTrafficCommand,
			expectedBlockErr: ErrIptablesFailure,
		},
		{
			name:                       "unblock failure",
			invalidCommand:             unblockUnrelatedTrafficCommand,
			expectedCommandsAfterBlock: blockCommands,
			expectedUnblockErr:         ErrIptablesFailure,
		},
		{
			name:             "firewall disabled",
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, !test.firewallDisabled)

			err := firewallManager.BlockMeshnet(deviceAddress)
			if test.expectedBlockErr != nil {
				assert.ErrorIs(t, err, test.expectedBlockErr, "Invalid error returned by BlockMeshnet.")
				return
			}
			assert.NoError(t, err, "Unexpected error when blocking meshnet.")

			blockCommands := iptablesMock.PopCommands()
			assert.Equal(t, test.expectedCommandsAfterBlock, blockCommands,
				"Invalid commands executed by BlockMeshnet.")

			err = firewallManager.UnblockMeshnet()
			if test.expectedUnblockErr != nil {
				assert.ErrorIs(t, err, test.expectedUnblockErr, "Invalid error returned by UnblockMeshnet.")
				return
			}
			assert.NoError(t, err, "Unexpected error when unblocking meshnet.")

			unblockCommands := iptablesMock.PopCommands()
			assert.Equal(t, test.expectedCommandsAfterUnblock, unblockCommands,
				"Invalid commands executed by UnblockMeshnet.")
		})
	}
}

func TestBlockMeshnet_AlreadyBlocked(t *testing.T) {
	deviceAddress := "48.242.30.25"

	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, true)

	err := firewallManager.BlockMeshnet(deviceAddress)
	assert.NoError(t, err, "Unexpected error when blocking meshnet.")
	iptablesMock.PopCommands()

	err = firewallManager.BlockMeshnet(deviceAddress)
	assert.ErrorIs(t, err, ErrRuleAlreadyActive)
	iptablesMock.PopCommands()

	// using different device address shoudl yield the same result
	err = firewallManager.BlockMeshnet("128.236.166.204")
	assert.ErrorIs(t, err, ErrRuleAlreadyActive)
}

func TestUnblockMeshnet_NotBlocked(t *testing.T) {
	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(nil, &iptablesMock, connmark, true)

	err := firewallManager.UnblockMeshnet()
	assert.Error(t, err, ErrRuleAlreadyActive)
}

func TestSetAllowlist(t *testing.T) {
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
		{
			name:             "firewall disabled",
			firewallDisabled: true,
			deviceFunc:       getDeviceFunc(false, mock.En0Interface),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(test.deviceFunc, &iptablesMock, connmark, !test.firewallDisabled)

			err := firewallManager.SetAllowlist(udpPorts, tcpPorts, subnets)
			if test.expectedErrSet != nil {
				assert.ErrorIs(t, err, test.expectedErrSet, "Invalid error returned by SetAllowlist.")
				return
			}

			commandsAfterSet := iptablesMock.PopCommands()
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
			commandsAfterUnset := iptablesMock.PopCommands()
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
		{
			name:             "firewall disabled",
			deviceFunc:       getDeviceFunc(false, mock.En0Interface),
			firewallDisabled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			iptablesMock := NewIptablesMock(false)
			if test.invalidCommand != "" {
				iptablesMock.AddErrCommand(test.invalidCommand)
			}

			firewallManager := NewFirewallManager(test.deviceFunc, &iptablesMock, connmark, !test.firewallDisabled)

			err := firewallManager.ApiAllowlist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiAllowlist.")
				return
			}

			commandsAfterApiAllowlist := iptablesMock.PopCommands()
			assert.Len(t, commandsAfterApiAllowlist, len(test.expectedAllowlistCommands),
				"Invalid commands executed after api allowlist.")
			for _, expectedCommand := range test.expectedAllowlistCommands {
				assert.Contains(t, commandsAfterApiAllowlist, expectedCommand,
					"Expected command not found after api allowlist.")
			}

			err = firewallManager.ApiDenylist()
			if test.expectedAllowlistError != nil {
				assert.ErrorIs(t, err, test.expectedAllowlistError, "Invalid error returned by ApiDenylist.")
				return
			}

			commandsAfterApiDenylist := iptablesMock.PopCommands()
			assert.Len(t, commandsAfterApiDenylist, len(test.expectedDenylistCommands),
				"Invalid commands executed after api denylist.")
			for _, expectedCommand := range test.expectedDenylistCommands {
				assert.Contains(t, commandsAfterApiDenylist, expectedCommand,
					"Expected command not found after api denylist.")
			}
		})
	}
}

func TestEnableDisable(t *testing.T) {
	iptablesMock := NewIptablesMock(false)
	firewallManager := NewFirewallManager(getDeviceFunc(false, mock.En0Interface), &iptablesMock, connmark, true)

	// static commands must always remain in the same place, at the bottom of the table
	firewallManager.BlockTraffic()
	firewallManager.ApiAllowlist()
	firewallManager.BlockMeshnet("230.239.113.214")
	staticCommands := iptablesMock.PopCommands()

	// transient commands change the order depending on users action, i.e when subnet was added to allowlist, when peer
	// was added/removed, etc.
	peer1Address := meshnet.UniqueAddress{
		UID:     "D3YXjHgrzVw6Tniwd7p5zpXD0RGgx3BpMivueganzet=",
		Address: netip.MustParseAddr("230.165.178.53"),
	}
	firewallManager.AllowIncoming(peer1Address, true)

	peer2Address := meshnet.UniqueAddress{
		UID:     "zQyBhKrCtcmuzZimzugZxZXFihuyznjCPPtZUUjVY=",
		Address: netip.MustParseAddr("230.165.178.53"),
	}
	firewallManager.AllowIncoming(peer2Address, false)

	firewallManager.AllowFileshare(peer1Address)

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
	firewallManager.SetAllowlist(udpPorts, tcpPorts, subnets)

	transientCommands := iptablesMock.PopCommands()

	firewallManager.Disable()

	commandsAfterDisable := iptablesMock.PopCommands()

	expectedStaticCommands := transformCommandsToDelte(t, staticCommands)
	staticCommandsAfterDisable := commandsAfterDisable[:len(expectedStaticCommands)]
	assert.Equal(t,
		expectedStaticCommands,
		staticCommandsAfterDisable,
		"Invalid commands executed after disabling firewall:\nExpected:\n%s\nGot:\n%s",
		transformCommandsForPrinting(t, expectedStaticCommands),
		transformCommandsForPrinting(t, staticCommandsAfterDisable))

	expectedTransientCommands := transformCommandsToDelte(t, transientCommands)
	transientCommandsAfterDisable := commandsAfterDisable[len(expectedStaticCommands):]
	for _, expectedCommand := range expectedTransientCommands {
		assert.Contains(t,
			transientCommandsAfterDisable,
			expectedCommand,
			"Expected command not executed after disabling firewall.")
	}

	firewallManager.Enable()
	commandsAfterReenable := iptablesMock.PopCommands()

	staticCommandsAfterReenable := commandsAfterReenable[:len(staticCommands)]
	assert.Equal(t,
		staticCommands,
		staticCommandsAfterReenable,
		"Invalid commands executed after reenabling firewall:\n%sExpected:\nGot:\n%s",
		transformCommandsForPrinting(t, staticCommands),
		transformCommandsForPrinting(t, staticCommandsAfterReenable))

	transientCommandsAfterReenable := commandsAfterReenable[len(staticCommands):]
	for _, expectedCommand := range transientCommands {
		assert.Contains(t,
			transientCommandsAfterReenable,
			expectedCommand,
			"Expected command not executed after reenablign firewall.")
	}
}
