package config

import (
	"encoding/json"
	"slices"
)

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

func (a *Allowlist) UpdateSubnets(subnet string, remove bool) {
	if remove {
		a.Subnets = slices.DeleteFunc(a.Subnets, func(element string) bool { return element == subnet })
	} else {
		a.Subnets = append(a.Subnets, subnet)
	}
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
