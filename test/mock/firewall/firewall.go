package firewall

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"golang.org/x/exp/slices"
)

type FirewallMock struct {
	Rules []firewall.Rule
}

func NewMockFirewall() FirewallMock {
	return FirewallMock{
		Rules: []firewall.Rule{},
	}
}

// Add and apply firewall rules
func (mf *FirewallMock) Add(rules []firewall.Rule) error {
	mf.Rules = append(mf.Rules, rules...)

	return nil
}

// Delete a list of firewall rules by defined names
func (mf *FirewallMock) Delete(names []string) error {
	for _, name := range names {
		nameIndex := slices.IndexFunc(mf.Rules, func(r firewall.Rule) bool { return r.Name == name })
		if nameIndex != -1 {
			mf.Rules = append(mf.Rules[:nameIndex], mf.Rules[nameIndex+1:]...)
		}
	}

	return nil
}

// Enable firewall
func (mf *FirewallMock) Enable() error {
	return nil
}

// Disable firewall
func (mf *FirewallMock) Disable() error {
	return nil
}

// Flush firewall
func (mf *FirewallMock) Flush() error {
	return nil
}
