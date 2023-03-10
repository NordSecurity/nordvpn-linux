package firewall

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		original error
	}{
		{
			name:     "rule not found",
			original: ErrRuleNotFound,
		},
		{
			name:     "rule already exists",
			original: ErrRuleAlreadyExists,
		},
		{
			name:     "nameless rule",
			original: ErrRuleWithoutName,
		},
		{
			name:     "already enabled",
			original: ErrFirewallAlreadyEnabled,
		},
		{
			name:     "already disabled",
			original: ErrFirewallAlreadyDisabled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewError(test.original)
			assert.Equal(t, test.original.Error(), err.Error())
			assert.True(t, errors.Is(err, test.original))
		})
	}
}
