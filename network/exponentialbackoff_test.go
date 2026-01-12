package network

import (
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_ExponentialBackoff(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		tries       int
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{
			name:        "tries2",
			tries:       2,
			expectedMin: 5 * time.Second,
			expectedMax: 10 * time.Second,
		},
		{
			name:        "tries8",
			tries:       8,
			expectedMin: 10 * time.Second,
			expectedMax: 60 * time.Second,
		},
		{
			name:        "tries19",
			tries:       19,
			expectedMin: 60 * time.Second,
			expectedMax: 300 * time.Second,
		},
		{
			name:        "triesDefault",
			tries:       200,
			expectedMin: 300 * time.Second,
			expectedMax: 600 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.LessOrEqual(t, test.expectedMin, ExponentialBackoff(test.tries))
			assert.GreaterOrEqual(t, test.expectedMax, ExponentialBackoff(test.tries))
		})
	}
}
