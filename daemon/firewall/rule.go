package firewall

import (
	"net"
	"net/netip"

	"golang.org/x/exp/slices"
)

// PortsDirection represents direction in which ports are open to, source, destination or both
type PortsDirection int

const (
	SourceAndDestination PortsDirection = iota
	Destination
	Source
)

// ConnectionState defines a state of a connection
type ConnectionState int

const (
	// Established means that packet is associated with a connection
	Established ConnectionState = iota
	// Related means that packet creates a new connection, but it is related with the existing one
	Related
	// New means that packet creates a new connection
	New
)

// Direction defines a direction of packages to which rule is applicable
type Direction int

const (
	// Inbound defines that rule is applicable for incoming packets
	Inbound Direction = iota
	// Outbound defines that rule is applicable for outgoing packets
	Outbound
	// TwoWay defines that rule is applicable for both incoming and outgoing packets
	TwoWay
)

type ConnectionStates struct {
	SrcAddr netip.Addr
	States  []ConnectionState
}

// Rule defines a single firewall rule which is applicable for set of addresses, ports and protocols
type Rule struct {
	// Name of the firewall rule
	Name string `json:"name"`
	// Interfaces define a list of network interfaces to which rule is applicable
	Interfaces []net.Interface `json:"interfaces"`
	// Networks is a list of IP networks to which rule is applicable
	RemoteNetworks []netip.Prefix `json:"remote_networks"`
	LocalNetworks  []netip.Prefix `json:"local_networks"`

	// Ports is a list of ports to which rule is applicable
	Ports []int `json:"ports"`
	// PortsDirection is a direction that ports are open to
	PortsDirection PortsDirection
	// Protocols is a list of protocol string values to which rule is applicable
	Protocols []string `json:"protocols"`
	// Direction defines to which packets rule is applicable
	Direction Direction `json:"direction"`
	// ConnectionStates defines to which connection states rule is applicable
	ConnectionStates ConnectionStates `json:"connection_states"`
	// Marks defines that packets marked with any of the marks are
	// affected by the firewall rule
	Marks []uint32
	// Allow defines if rule denies packets via current rule or allows them
	Allow bool `json:"allow"`

	Ipv6Only         bool   `json:"ipv6_only"`
	Icmpv6Types      []int  `json:"icmp6_types"`
	HopLimit         uint8  `json:"hop_limit"`
	SourcePorts      []int  `json:"source_ports"`
	DestinationPorts []int  `json:"destination_ports"`
	Comment          string `json:"comment"`
}

// OrderedRules stores rules in an order they were added.
type OrderedRules struct {
	// rules is unexported in order to prevent direct appends
	rules []Rule
}

func byName(name string) func(Rule) bool {
	return func(rule Rule) bool { return rule.Name == name }
}

func (or *OrderedRules) Add(rule Rule) error {
	if rule.Name == "" {
		return NewError(ErrRuleWithoutName)
	}

	if slices.ContainsFunc(or.rules, byName(rule.Name)) {
		return NewError(ErrRuleAlreadyExists)
	}

	or.rules = append(or.rules, rule)
	return nil
}

func (or *OrderedRules) Get(name string) (Rule, error) {
	index := slices.IndexFunc(or.rules, byName(name))
	if index == -1 {
		return Rule{}, NewError(ErrRuleNotFound)
	}
	return or.rules[index], nil
}

// Delete rule by name if found.
func (or *OrderedRules) Delete(name string) error {
	index := slices.IndexFunc(or.rules, byName(name))
	if index == -1 {
		return NewError(ErrRuleNotFound)
	}

	if len(or.rules) <= 1 {
		or.rules = []Rule{}
		return nil
	}

	or.rules = slices.Delete(or.rules, index, index+1)
	return nil
}
