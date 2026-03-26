package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"slices"
)

// ErrSubnetAlreadyCovered is returned when a subnet being added
// is already covered by a wider existing subnet.
var ErrSubnetAlreadyCovered = errors.New("subnet is already covered by existing subnet")

// NewAllowlist ready to use
func NewAllowlist(udpPorts []int64, tcpPorts []int64, subnets []string) Allowlist {
	udp := map[int64]bool{}
	for _, port := range udpPorts {
		udp[port] = true
	}

	tcp := map[int64]bool{}
	for _, port := range tcpPorts {
		tcp[port] = true
	}

	return Allowlist{
		Ports: Ports{
			UDP: udp,
			TCP: tcp,
		},
		Subnets: subnets,
	}
}

// Allowlist is a collection of ports and subnets
type Allowlist struct {
	Ports   Ports    `json:"ports"`
	Subnets []string `json:"subnets"` // TODO change to netip.Prefix and refactor
}

func (a *Allowlist) UpdateUDPPorts(ports []int64, remove bool) {
	for _, port := range ports {
		if remove {
			delete(a.Ports.UDP, port)
		} else {
			a.Ports.UDP[port] = true
		}
	}
}

func (a *Allowlist) UpdateTCPPorts(ports []int64, remove bool) {
	for _, port := range ports {
		if remove {
			delete(a.Ports.TCP, port)
		} else {
			a.Ports.TCP[port] = true
		}
	}
}

// NormalizeSubnets find overlapping subnets and merge them, also cleanup any invalid
// subnet values, if any found, do not return errors
func (a *Allowlist) NormalizeSubnets(onRemove func(removed, reason string)) {
	type parsed struct {
		raw    string
		prefix netip.Prefix
	}

	items := make([]parsed, 0, len(a.Subnets))
	for _, s := range a.Subnets {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			if onRemove != nil {
				onRemove(s, "invalid value")
			}
			continue
		}
		if !p.Addr().Is4() {
			if onRemove != nil {
				onRemove(s, "invalid value (non IPv4)")
			}
			continue
		}
		items = append(items, parsed{raw: s, prefix: p})
	}

	var result []string
	for i, a := range items {
		coveredByWider := false
		var coveredBy string
		for j, b := range items {
			if i == j {
				continue
			}
			if !a.prefix.Overlaps(b.prefix) {
				continue
			}
			// b is strictly wider (fewer bits) — a is redundant.
			// On equal bits keep the one that appears first.
			if b.prefix.Bits() < a.prefix.Bits() ||
				(b.prefix.Bits() == a.prefix.Bits() && i < j) {
				coveredByWider = true
				coveredBy = b.raw
				break
			}
		}
		if coveredByWider {
			if onRemove != nil {
				onRemove(a.raw, fmt.Sprintf("covered by: %s", coveredBy))
			}
		} else {
			result = append(result, a.raw)
		}
	}
	a.Subnets = result
}

// WouldEliminateSubnets check if new subnet would cover (i.e. eliminate) some existing subnet(-s)
func (a *Allowlist) WouldEliminateSubnets(subnet string) ([]string, error) {
	wouldbeEliminatedSubnets := []string{}
	newPrefix, err := netip.ParsePrefix(subnet)
	if err != nil {
		return nil, fmt.Errorf("parsing subnet %q: %w", subnet, err)
	}

	for _, existingSubnet := range a.Subnets {
		existingPrefix, err := netip.ParsePrefix(existingSubnet)
		if err != nil {
			return nil, fmt.Errorf("parsing existing subnet %q: %w", existingSubnet, err)
		}

		if !newPrefix.Overlaps(existingPrefix) {
			continue
		}

		// if new subnet is wider
		if newPrefix.Bits() < existingPrefix.Bits() {
			wouldbeEliminatedSubnets = append(wouldbeEliminatedSubnets, existingSubnet)
		}
	}
	return wouldbeEliminatedSubnets, nil
}

// addSubnet try to add subnet, check if subnet being added is covered by some existing subnet,
// also check if subnet being added covers (i.e. eliminates) some existing subnet
func (a *Allowlist) addSubnet(subnet string, onRemove func(removed, coveredBy string)) error {
	newPrefix, err := netip.ParsePrefix(subnet)
	if err != nil {
		return fmt.Errorf("parsing subnet %q: %w", subnet, err)
	}

	for i := 0; i < len(a.Subnets); i++ {
		existingPrefix, err := netip.ParsePrefix(a.Subnets[i])
		if err != nil {
			return fmt.Errorf("parsing existing subnet %q: %w", a.Subnets[i], err)
		}

		if !newPrefix.Overlaps(existingPrefix) {
			continue
		}

		// New subnet is smaller or equal — existing already covers it.
		if newPrefix.Bits() >= existingPrefix.Bits() {
			return fmt.Errorf("%w: %s", ErrSubnetAlreadyCovered, a.Subnets[i])
		}

		// New subnet is wider — remove the existing narrower one and
		// keep checking remaining entries.
		if onRemove != nil {
			onRemove(a.Subnets[i], subnet)
		}
		a.Subnets = slices.Delete(a.Subnets, i, i+1)
		i--
	}

	a.Subnets = append(a.Subnets, subnet)
	return nil
}

func (a *Allowlist) UpdateSubnets(subnet string, remove bool, onRemove func(removed, coveredBy string)) error {
	if remove {
		a.Subnets = slices.DeleteFunc(a.Subnets, func(element string) bool { return element == subnet })
	} else {
		return a.addSubnet(subnet, onRemove)
	}
	return nil
}

// GetUDPPorts returns a slice of all UDP ports within the allowlist
func (a *Allowlist) GetUDPPorts() []int64 {
	ports := []int64{}
	for port := range a.Ports.UDP {
		ports = append(ports, port)
	}

	return ports
}

// GetTCPPorts returns a slice of all TCP ports within the allowlist
func (a *Allowlist) GetTCPPorts() []int64 {
	ports := []int64{}
	for port := range a.Ports.TCP {
		ports = append(ports, port)
	}

	return ports
}

// Ports is a collection of TCP and UDP ports.
type Ports struct {
	TCP PortSet `json:"tcp"`
	UDP PortSet `json:"udp"`
}

// PortSet is a set of ports.
type PortSet map[int64]bool

// MarshalJSON into []float64.
func (p PortSet) MarshalJSON() ([]byte, error) {
	var ports []float64
	for port := range p {
		ports = append(ports, float64(port))
	}

	return json.Marshal(ports)
}

// UnmarshalJSON into map[int64]bool.
func (p *PortSet) UnmarshalJSON(b []byte) error {
	var i []float64
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	ports := map[int64]bool{}
	for _, port := range i {
		ports[int64(port)] = true
	}

	*p = ports
	return nil
}

func (p *PortSet) ToSlice() []int64 {
	result := make([]int64, 0, len(*p))
	for subnet := range *p {
		result = append(result, subnet)
	}
	return result
}
