package device

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestInterfaceNameFromIpRoute(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    string
		output   string
		hasError bool
	}{
		{
			input:    "172.17.0.0/16 dev docker0 proto kernel scope link src 172.17.0.1 linkdown",
			output:   "docker0",
			hasError: false,
		},
		{
			input:    "172.17.0.0/16 br0 proto kernel scope link src 172.17.0.1 linkdown",
			output:   "",
			hasError: true,
		},
		{
			input:    "",
			output:   "",
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			name, err := interfaceNameFromIPRoute(test.input)
			assert.Equal(t, test.hasError, err != nil)
			assert.Equal(t, name, test.output)
		})
	}
}
