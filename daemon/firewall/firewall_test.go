package firewall_test

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	firewallmock "github.com/NordSecurity/nordvpn-linux/test/mock/firewall"
	"github.com/stretchr/testify/assert"
)

func TestFirewallService(t *testing.T) {
	assert.Implements(t, (*firewall.Service)(nil), &firewall.Firewall{})
}

func TestFirewallConfigure(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		backend firewall.FirewallBackend
		enabled bool
		err     error
	}{
		{
			name:    "successful configured",
			enabled: true,
			backend: &firewallmock.FirewallBackend{},
		},
		{
			name:    "configure fails when backend returns error",
			enabled: true,
			backend: &firewallmock.FirewallBackend{Err: mock.ErrOnPurpose},
			err:     mock.ErrOnPurpose,
		},
		{
			name:    "no error is returned when the firewall is not enabled",
			enabled: false,
			backend: &firewallmock.FirewallBackend{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := firewall.NewFirewall(test.backend, test.enabled, "", &subs.Subject[events.DebuggerEvent]{})

			assert.ErrorIs(t, fw.Configure(firewall.Config{}), test.err)
		})
	}
}

func TestFirewallEnable(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		enabled  bool
		hasError bool
	}{
		{
			name:     "enabled",
			enabled:  true,
			hasError: true,
		},
		{
			name:     "disabled",
			enabled:  false,
			hasError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := firewall.NewFirewall(&firewallmock.FirewallBackend{}, test.enabled, "", &subs.Subject[events.DebuggerEvent]{})
			err := fw.Enable()
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*firewall.Error)
				assert.True(t, ok)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFirewallDisable(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		enabled  bool
		hasError bool
	}{
		{
			name:     "enabled",
			enabled:  true,
			hasError: false,
		},
		{
			name:     "disabled",
			enabled:  false,
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fw := firewall.NewFirewall(&firewallmock.FirewallBackend{}, test.enabled, "", &subs.Subject[events.DebuggerEvent]{})
			err := fw.Disable()
			if test.hasError {
				assert.Error(t, err)
				_, ok := err.(*firewall.Error)
				assert.True(t, ok)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
