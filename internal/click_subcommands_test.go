package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestSubcommandURI_ProducesCorrectURI(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct{ subcommand, expected string }{
		{ClaimOnlinePurchaseSubcommand, "nordvpn://claim-online-purchase"},
		{LoginSubcommand, "nordvpn://login"},
		{ConsentSubcommand, "nordvpn://consent"},
		{"something-else", "nordvpn://something-else"},
	}

	for _, item := range tests {
		got := SubcommandURI(item.subcommand)
		assert.Equal(t, item.expected, got)
	}
}
