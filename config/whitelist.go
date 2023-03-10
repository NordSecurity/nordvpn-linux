package config

import (
	"encoding/json"
)

// NewWhitelist ready to use
func NewWhitelist(udpPorts []int64, tcpPorts []int64, subnets []string) Whitelist {
	udp := map[int64]bool{}
	for _, port := range udpPorts {
		udp[port] = true
	}

	tcp := map[int64]bool{}
	for _, port := range tcpPorts {
		tcp[port] = true
	}

	subs := map[string]bool{}
	for _, sub := range subnets {
		subs[sub] = true
	}

	return Whitelist{
		Ports: Ports{
			UDP: udp,
			TCP: tcp,
		},
		Subnets: subs,
	}
}

// Whitelist is a collection of ports and subnets
type Whitelist struct {
	Ports   Ports   `json:"ports"`
	Subnets Subnets `json:"subnets"`
}

// Subnets is a set of subnets.
type Subnets map[string]bool

// MarshalJSON into []string.
func (s Subnets) MarshalJSON() ([]byte, error) {
	var subnets []string
	for subnet := range s {
		subnets = append(subnets, subnet)
	}

	return json.Marshal(subnets)
}

// UnmarshalJSON into map[string]bool.
func (s *Subnets) UnmarshalJSON(b []byte) error {
	var i []string
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	subnets := map[string]bool{}
	for _, subnet := range i {
		subnets[subnet] = true
	}

	*s = subnets
	return nil
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
