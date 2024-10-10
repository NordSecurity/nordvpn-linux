package network

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"

	"github.com/stretchr/testify/assert"
)

type workingAgent struct{}

func (workingAgent) Add(firewall.Rule) error    { return nil }
func (workingAgent) Delete(firewall.Rule) error { return nil }
func (workingAgent) Flush() error               { return nil }

type failingAgent struct{}

func (failingAgent) Add(firewall.Rule) error    { return mock.ErrOnPurpose }
func (failingAgent) Delete(firewall.Rule) error { return mock.ErrOnPurpose }
func (failingAgent) Flush() error               { return mock.ErrOnPurpose }

func TestAllowlistIP(t *testing.T) {
	category.Set(t, category.Route)

	tests := []struct {
		name     string
		rules    []firewall.Rule
		agent    firewall.Agent
		ips      []netip.Addr
		hasError bool
	}{
		{
			name:     "empty slice",
			rules:    []firewall.Rule{},
			agent:    &workingAgent{},
			ips:      []netip.Addr{},
			hasError: false,
		},
		{
			name:     "nil slice",
			rules:    nil,
			agent:    &workingAgent{},
			ips:      nil,
			hasError: false,
		},
		{
			name: "block rule",
			rules: []firewall.Rule{
				{
					Name: "block",
				},
			},
			agent: &workingAgent{},
			ips: []netip.Addr{
				netip.AddrFrom16(
					[16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff},
				),
			},
			hasError: false,
		},
		{
			name: "drop rule",
			rules: []firewall.Rule{
				{
					Name: "drop",
				},
			},
			agent: &workingAgent{},
			ips: []netip.Addr{
				netip.AddrFrom16(
					[16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff},
				),
			},
			hasError: false,
		},
		{
			name: "multiple rules added",
			rules: []firewall.Rule{
				{
					Name: "allow",
				},
				{
					Name: "block",
				},
				{
					Name: "permit",
				},
				{
					Name: "drop",
				},
			},
			agent: &workingAgent{},
			ips: []netip.Addr{
				netip.AddrFrom16(
					[16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff},
				),
			},
			hasError: false,
		},
		{
			name:     "agent failure",
			rules:    []firewall.Rule{{}},
			agent:    &failingAgent{},
			ips:      []netip.Addr{},
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := firewall.NewFirewall(test.agent, test.agent, &subs.Subject[string]{}, true)
			fw.Add(test.rules)
			err := allowlistIP(fw, test.name, test.ips...)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
