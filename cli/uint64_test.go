package cli

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestUint64ToHumanBytes(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    uint64
		expected string
	}{
		{1, "1 B"},
		{575, "575 B"},
		{1024000, "0.98 MiB"},
		{16672358, "15.90 MiB"},
		{200897095270, "187.10 GiB"},
		{1099404253593, "1.00 TiB"},
		{838927371993088, "0.75 PiB"},
		{211331412514360500, "187.70 PiB"},
		{859061628920922100, "0.75 EiB"},
	}

	for _, data := range tests {
		got := uint64ToHumanBytes(data.input)
		assert.Equal(t, got, data.expected)
	}
}
