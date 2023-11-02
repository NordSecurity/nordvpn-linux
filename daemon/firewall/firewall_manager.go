package firewall

import (
	"errors"
	"fmt"
	"log"
	"net/netip"
	"os/exec"
	"sort"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
)

var ErrIncomingAlreadyDenied = errors.New("allow incoming rule for the given peer not found")

type CommandExecutor interface {
	ExecuteCommand(command string) error
}

type ExecCommandExecutor struct {
	command string
}

func NewCommandExecutor(command string) ExecCommandExecutor {
	return ExecCommandExecutor{
		command: command,
	}
}

func (e ExecCommandExecutor) ExecuteCommand(command string) error {
	log.Printf("DEBUG: %s %s", e.command, command)
	commandArgs := strings.Split(command, " ")

	_, err := exec.Command(e.command, commandArgs...).CombinedOutput()
	return err
}

type allowIncomingRule struct {
	allowIncomingRule string
	blockLANRules     []string
}

type PortRange struct {
	min int
	max int
}

type FirewallManager struct {
	commandExecutor      CommandExecutor
	devices              device.ListFunc              // list network interfaces
	allowIncomingRules   map[string]allowIncomingRule // peer public key to allow incoming rule
	fileshareRules       map[string]string            // peers public key to allow fileshare rule
	allowlistRules       []string
	connmark             uint32
	meshnetDeviceAddress string // used for unblocking meshnet after if has been blocked and for tracking meshnet block state
	trafficBlocked       bool
	enabled              bool
}

func NewFirewallManager(devices device.ListFunc, commandExecutor CommandExecutor, connmark uint32, enabled bool) FirewallManager {
	return FirewallManager{
		commandExecutor:    commandExecutor,
		devices:            devices,
		allowIncomingRules: make(map[string]allowIncomingRule),
		fileshareRules:     make(map[string]string),
		connmark:           connmark,
		enabled:            enabled,
	}
}

func (f *FirewallManager) Disable() error {
	if !f.enabled {
		return fmt.Errorf("firewall is already disabled")
	}

	// remove traffic block
	if f.trafficBlocked {
		if err := f.removeBlockTrafficRules(); err != nil {
			return fmt.Errorf("unblocking traffic: %w", err)
		}
	}

	// remove api allowlist
	if err := f.manageApiAllowlist(false); err != nil {
		return fmt.Errorf("removing api allowlist %w", err)
	}

	// remove allowlist
	for _, rule := range f.allowlistRules {
		if err := f.commandExecutor.ExecuteCommand("-D" + rule); err != nil {
			return fmt.Errorf("removing allowlist rule: %w", err)
		}
	}

	// remove meshnet block rules
	if f.meshnetDeviceAddress != "" {
		if err := f.removeMeshnetBlockRules(f.meshnetDeviceAddress); err != nil {
			return fmt.Errorf("removing meshnet block rules: %w", err)
		}
	}

	// remove allow incoming rules
	for _, rule := range f.allowIncomingRules {
		if err := f.removeIncomingRule(rule); err != nil {
			return fmt.Errorf("removing incoming rules: %w", err)
		}
	}

	// remove allow fileshare rules
	for _, rule := range f.fileshareRules {
		if err := f.commandExecutor.ExecuteCommand("-D " + rule); err != nil {
			return fmt.Errorf("removing fileshare allow rule: %w", err)
		}
	}

	f.enabled = false

	return nil
}

func (f *FirewallManager) Enable() error {
	if f.enabled {
		return fmt.Errorf("firewall is already enabled")
	}

	// add traffic block
	if f.trafficBlocked {
		if err := f.addBlockTrafficRules(); err != nil {
			return fmt.Errorf("blocking traffic: %w", err)
		}
	}

	// add api allowlist
	if err := f.manageApiAllowlist(true); err != nil {
		return fmt.Errorf("adding api allowlist %w", err)
	}

	// add allowlist
	for _, rule := range f.allowlistRules {
		if err := f.commandExecutor.ExecuteCommand("-I" + rule); err != nil {
			return fmt.Errorf("adding allowlist rule: %w", err)
		}
	}

	// add meshnet block rules
	if f.meshnetDeviceAddress != "" {
		if err := f.addMeshnetBlockRules(f.meshnetDeviceAddress); err != nil {
			return fmt.Errorf("adding meshnet block rules: %w", err)
		}
	}

	// add allow incoming rules
	for _, rule := range f.allowIncomingRules {
		if err := f.addIncomingRule(rule); err != nil {
			return fmt.Errorf("adding incoming rules: %w", err)
		}
	}

	// add allow fileshare rules
	for _, rule := range f.fileshareRules {
		if err := f.commandExecutor.ExecuteCommand("-I " + rule); err != nil {
			return fmt.Errorf("adding fileshare allow rule: %w", err)
		}
	}

	f.enabled = true

	return nil
}

// FileshareAllow adds ACCEPT rule for all incoming connections to tcp port 49111 from the peer with given UniqueAddress.
func (f *FirewallManager) FileshareAllow(peer meshnet.UniqueAddress) error {
	if _, ok := f.fileshareRules[peer.UID]; ok {
		return fmt.Errorf("fileshare allow rule for the given peer already exists")
	}

	rule := fmt.Sprintf("INPUT -s %s/32 -p tcp -m tcp --dport 49111 -m comment --comment nordvpn -j ACCEPT", peer.Address.String())
	if f.enabled {
		if err := f.commandExecutor.ExecuteCommand("-I " + rule); err != nil {
			return fmt.Errorf("adding fileshare allow rule: %w", err)
		}
	}

	f.fileshareRules[peer.UID] = rule
	return nil
}

// FileshareDeny removes ACCEPT rule for all incoming connections to tcp port 49111 from the peer with given UniqueAddress.
func (f *FirewallManager) FileshareDeny(peer meshnet.UniqueAddress) error {
	rule, ok := f.fileshareRules[peer.UID]
	if !ok {
		return fmt.Errorf("fileshare allow rule for the given peer doesn't exist")
	}

	if f.enabled {
		if err := f.commandExecutor.ExecuteCommand("-D " + rule); err != nil {
			return fmt.Errorf("removing fileshare allow rule: %w", err)
		}
	}

	delete(f.fileshareRules, peer.UID)
	return nil
}

func (f *FirewallManager) addBlockTrafficRules() error {
	interfaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	// -I INPUT -i <iface> -m comment --comment nordvpn -j DROP
	// -I OUTPUT -o <iface> -m comment --comment nordvpn -j DROP
	for _, iface := range interfaces {
		inputCommand := fmt.Sprintf("-I INPUT -i %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.ExecuteCommand(inputCommand); err != nil {
			return fmt.Errorf("blocking input traffic: %w", err)
		}

		outputCommand := fmt.Sprintf("-I OUTPUT -o %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.ExecuteCommand(outputCommand); err != nil {
			return fmt.Errorf("blocking output traffic: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) BlockTraffic() error {
	if f.trafficBlocked {
		return fmt.Errorf("traffic is already blocked")
	}

	if f.enabled {
		if err := f.addBlockTrafficRules(); err != nil {
			return fmt.Errorf("adding block traffic rules: %w", err)
		}
	}

	f.trafficBlocked = true

	return nil
}

func (f *FirewallManager) removeBlockTrafficRules() error {
	interfaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	// -D INPUT -i <iface> -m comment --comment nordvpn -j DROP
	// -D OUTPUT -o <iface> -m comment --comment nordvpn -j DROP
	for _, iface := range interfaces {
		inputCommand := fmt.Sprintf("-D INPUT -i %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.ExecuteCommand(inputCommand); err != nil {
			return fmt.Errorf("unblocking input traffic: %w", err)
		}

		outputCommand := fmt.Sprintf("-D OUTPUT -o %s -m comment --comment nordvpn -j DROP", iface.Name)
		if err := f.commandExecutor.ExecuteCommand(outputCommand); err != nil {
			return fmt.Errorf("unblocking output traffic: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) UnblockTraffic() error {
	if !f.trafficBlocked {
		return fmt.Errorf("traffic is already unblocked")
	}

	if f.enabled {
		if err := f.removeBlockTrafficRules(); err != nil {
			return fmt.Errorf("removing traffic block rules: %w", err)
		}
	}

	f.trafficBlocked = false

	return nil
}

func (f *FirewallManager) addIncomingRule(rule allowIncomingRule) error {
	if err := f.commandExecutor.ExecuteCommand("-I " + rule.allowIncomingRule); err != nil {
		return fmt.Errorf("adding allow incoming rule: %w", err)
	}

	for _, blockLANRule := range rule.blockLANRules {
		if err := f.commandExecutor.ExecuteCommand("-I " + blockLANRule); err != nil {
			return fmt.Errorf("adding block peer lan rule: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) AllowIncoming(peer meshnet.UniqueAddress, allowLocal bool) error {
	if _, ok := f.allowIncomingRules[peer.UID]; ok {
		return fmt.Errorf("allow incoming rule for the given peer already exists")
	}

	rule := fmt.Sprintf("INPUT -s %s/32 -m comment --comment nordvpn -j ACCEPT", peer.Address)

	blockLANRules := []string{}
	if !allowLocal {
		lans := []string{
			"169.254.0.0/16",
			"192.168.0.0/16",
			"172.16.0.0/12",
			"10.0.0.0/8",
		}

		for _, lan := range lans {
			blockLANRule := fmt.Sprintf("INPUT -s %s/32 -d %s -m comment --comment nordvpn -j DROP", peer, lan)
			blockLANRules = append(blockLANRules, blockLANRule)
		}
	}

	allowIncomingRule := allowIncomingRule{
		allowIncomingRule: rule,
		blockLANRules:     blockLANRules,
	}

	if f.enabled {
		if err := f.addIncomingRule(allowIncomingRule); err != nil {
			return fmt.Errorf("adding incoming rule: %w", err)
		}
	}

	f.allowIncomingRules[peer.UID] = allowIncomingRule

	return nil
}

func (f *FirewallManager) removeIncomingRule(rule allowIncomingRule) error {
	if err := f.commandExecutor.ExecuteCommand("-D " + rule.allowIncomingRule); err != nil {
		return fmt.Errorf("adding allow incoming rule: %w", err)
	}

	for _, blockLANCommand := range rule.blockLANRules {
		if err := f.commandExecutor.ExecuteCommand("-D " + blockLANCommand); err != nil {
			return fmt.Errorf("deleting block peer lan rule: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) DenyIncoming(peer meshnet.UniqueAddress) error {
	rule, ok := f.allowIncomingRules[peer.UID]

	if !ok {
		return ErrIncomingAlreadyDenied
	}

	if f.enabled {
		if err := f.removeIncomingRule(rule); err != nil {
			return fmt.Errorf("removing incoming rule: %w", err)
		}
	}

	delete(f.allowIncomingRules, peer.UID)

	return nil
}

func (f *FirewallManager) removeMeshnetBlockRules(deviceAddress string) error {
	// -D INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc <device address> -m comment --comment nordvpn -j ACCEPT
	// -D INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP
	command := fmt.Sprintf("-I INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc %s -m comment --comment nordvpn -j ACCEPT", f.meshnetDeviceAddress)
	if err := f.commandExecutor.ExecuteCommand(command); err != nil {
		return fmt.Errorf("blocking unrelated mesh traffic: %w", err)
	}

	err := f.commandExecutor.ExecuteCommand("-I INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP")
	if err != nil {
		return fmt.Errorf("blocking mesh traffic: %w", err)
	}

	return nil
}

func (f *FirewallManager) UnblockMeshnet() error {
	if f.meshnetDeviceAddress == "" {
		return fmt.Errorf("meshnet is already unblocked")
	}

	if f.enabled {
		if err := f.removeMeshnetBlockRules(f.meshnetDeviceAddress); err != nil {
			return fmt.Errorf("removing meshnet block rules: %w", err)
		}
	}

	f.meshnetDeviceAddress = ""

	return nil
}

func (f *FirewallManager) addMeshnetBlockRules(deviceAddress string) error {
	// -I INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc <device address> -m comment --comment nordvpn -j ACCEPT
	// -I INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP
	command := fmt.Sprintf("-I INPUT -s 100.64.0.0/10 -m conntrack --ctstate RELATED,ESTABLISHED --ctorigsrc %s -m comment --comment nordvpn -j ACCEPT", deviceAddress)
	if err := f.commandExecutor.ExecuteCommand(command); err != nil {
		return fmt.Errorf("blocking unrelated mesh traffic: %w", err)
	}

	err := f.commandExecutor.ExecuteCommand("-I INPUT -s 100.64.0.0/10 -m comment --comment nordvpn -j DROP")
	if err != nil {
		return fmt.Errorf("blocking mesh traffic: %w", err)
	}

	return nil
}

func (f *FirewallManager) BlockMeshnet(deviceAddress string) error {
	if f.meshnetDeviceAddress != "" {
		return fmt.Errorf("meshnet is already blocked")
	}

	if f.enabled {
		if err := f.addMeshnetBlockRules(deviceAddress); err != nil {
			return fmt.Errorf("adding meshnet block rules: %w", err)
		}
	}

	f.meshnetDeviceAddress = deviceAddress

	return nil
}

// PortsToPortRanges groups ports into ranges
func PortsToPortRanges(ports []int) []PortRange {
	if len(ports) == 0 {
		return nil
	}

	sort.Ints(ports)

	var ranges []PortRange
	pPort := ports[0]
	r := PortRange{min: pPort, max: pPort}
	for i, port := range ports[1:] {
		if port == ports[i]+1 {
			r.max = port
			continue
		}
		ranges = append(ranges, r)
		r = PortRange{min: port, max: port}
	}

	return append(ranges, r)
}

func (f *FirewallManager) allowlistPort(iface string, protocol string, portRange PortRange) error {
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --dport <port> -m comment --comment nordvpn -j ACCEPT
	// -A INPUT -i <interface> -p <protocol> -m <protocol> --sport <port> -m comment --comment nordvpn -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --sport <port> -m comment --comment nordvpn -j ACCEPT
	// -A OUTPUT -o <interface> -p <protocol> -m <protocol> --dport <port> -m comment --comment nordvpn -j ACCEPT
	inputDportRule := fmt.Sprintf("INPUT -i %s -p %s -m %s --dport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	inputSportRule := fmt.Sprintf("INPUT -i %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputDportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --dport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)
	outputSportRule := fmt.Sprintf("OUTPUT -o %s -p %s -m %s --sport %d:%d -m comment --comment nordvpn -j ACCEPT", iface, protocol, protocol, portRange.min, portRange.max)

	if f.enabled {
		if err := f.commandExecutor.ExecuteCommand("-I " + inputDportRule); err != nil {
			return fmt.Errorf("allowlisting input dport: %w", err)
		}

		if err := f.commandExecutor.ExecuteCommand("-I " + inputSportRule); err != nil {
			return fmt.Errorf("allowlisting input sport: %w", err)
		}

		if err := f.commandExecutor.ExecuteCommand("-I " + outputDportRule); err != nil {
			return fmt.Errorf("allowlisting output dport: %w", err)
		}

		if err := f.commandExecutor.ExecuteCommand("-I " + outputSportRule); err != nil {
			return fmt.Errorf("allowlisting input dport: %w", err)
		}
	}

	f.allowlistRules = append(f.allowlistRules, inputDportRule)
	f.allowlistRules = append(f.allowlistRules, inputSportRule)
	f.allowlistRules = append(f.allowlistRules, outputDportRule)
	f.allowlistRules = append(f.allowlistRules, outputSportRule)

	return nil
}

func (f *FirewallManager) SetAllowlist(udpPorts []int, tcpPorts []int, subnets []netip.Prefix) error {
	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, subnet := range subnets {
		for _, iface := range ifaces {
			inputRule := fmt.Sprintf("INPUT -s %s -i %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)
			outputRule := fmt.Sprintf("OUTPUT -d %s -o %s -m comment --comment nordvpn -j ACCEPT", subnet.String(), iface.Name)

			if f.enabled {
				if err := f.commandExecutor.ExecuteCommand("-I " + inputRule); err != nil {
					return fmt.Errorf("adding input accept rule for subnet: %w", err)
				}
				if err := f.commandExecutor.ExecuteCommand("-I " + outputRule); err != nil {
					return fmt.Errorf("adding output accept rule for subnet: %w", err)
				}
			}

			f.allowlistRules = append(f.allowlistRules, inputRule)
			f.allowlistRules = append(f.allowlistRules, outputRule)
		}
	}

	udpPortRanges := PortsToPortRanges(udpPorts)
	for _, portRange := range udpPortRanges {
		for _, iface := range ifaces {
			if err := f.allowlistPort(iface.Name, "udp", portRange); err != nil {
				return fmt.Errorf("allowlisting udp ports: %w", err)
			}
		}
	}

	tcpPortRanges := PortsToPortRanges(tcpPorts)
	for _, portRange := range tcpPortRanges {
		for _, iface := range ifaces {
			if err := f.allowlistPort(iface.Name, "tcp", portRange); err != nil {
				return fmt.Errorf("allowlisting tcp ports: %w", err)
			}
		}
	}

	return nil
}

func (f *FirewallManager) UnsetAllowlist() error {
	if f.enabled {
		for _, rule := range f.allowlistRules {
			if err := f.commandExecutor.ExecuteCommand("-D " + rule); err != nil {
				return fmt.Errorf("removing allowlist rule: %w", err)
			}
		}
	}

	f.allowlistRules = nil

	return nil
}

func (f *FirewallManager) manageApiAllowlist(allow bool) error {
	iptablesMode := "-I "
	if !allow {
		iptablesMode = "-D "
	}

	ifaces, err := f.devices()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	for _, iface := range ifaces {
		inputRule := fmt.Sprintf("INPUT -i %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.ExecuteCommand(iptablesMode + inputRule); err != nil {
			return fmt.Errorf("adding api allowlist INPUT rule: %w", err)
		}

		outputRule :=
			fmt.Sprintf("OUTPUT -o %s -m mark --mark %d -m comment --comment nordvpn -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
				iface.Name, f.connmark)
		if err := f.commandExecutor.ExecuteCommand(iptablesMode + outputRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}

		outputConnmarkRule := fmt.Sprintf("OUTPUT -o %s -m connmark --mark %d -m comment --comment nordvpn -j ACCEPT", iface.Name, f.connmark)
		if err := f.commandExecutor.ExecuteCommand(iptablesMode + outputConnmarkRule); err != nil {
			return fmt.Errorf("adding api allowlist OUTPUT rule: %w", err)
		}
	}

	return nil
}

func (f *FirewallManager) ApiAllowlist() error {
	if !f.enabled {
		return nil
	}

	return f.manageApiAllowlist(true)
}

func (f *FirewallManager) ApiDenylist() error {
	if !f.enabled {
		return nil
	}

	return f.manageApiAllowlist(false)
}
