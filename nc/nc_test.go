package nc

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/network"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestExponentialBackoff(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		tries   int
		average int
		delta   float64
	}{
		{
			name:    "try once",
			tries:   1,
			average: (10 + 5) / 2,
			delta:   (10 - 5 + 1) / 2,
		},
		{
			name:    "try a few times",
			tries:   7,
			average: (60 + 10) / 2,
			delta:   (60 - 10 + 1) / 2,
		},
		{
			name:    "try a dozen times",
			tries:   12,
			average: (300 + 60) / 2,
			delta:   (300 - 60 + 1) / 2,
		},
		{
			name:    "try a lot",
			tries:   50,
			average: (600 + 300) / 2,
			delta:   (600 - 300 + 1) / 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backoff := network.ExponentialBackoff(test.tries).Seconds()
			assert.InDelta(t, test.average, backoff, test.delta)
		})
	}
}
