package config

import (
	"encoding/json"
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

	subs := map[string]bool{}
	for _, sub := range subnets {
		subs[sub] = true
	}

	return Allowlist{
		Ports: Ports{
			UDP: udp,
			TCP: tcp,
		},
		Subnets: subs,
	}
}

// Allowlist is a collection of ports and subnets
type Allowlist struct {
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

func (s *Subnets) ToSlice() []string {
	result := make([]string, 0, len(*s))
	for subnet := range *s {
		result = append(result, subnet)
	}
	return result
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
