package remote

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestVersionMatch(t *testing.T) {
	category.Set(t, category.Unit)

	// ~3.7.1 means: >= 3.7.1 and < 3.8.0
	// ^3.7.1 means: >= 3.7.1 and < 4.0.0

	tests := []struct {
		name         string
		srcVer       string
		trgVer       string
		match        bool
		expectsError bool
	}{
		{
			name:         "exact match",
			srcVer:       "1.1.1",
			trgVer:       "1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "invalid 1",
			srcVer:       "-",
			trgVer:       "1.1.1",
			match:        false,
			expectsError: true,
		},
		{
			name:         "invalid 2",
			srcVer:       "",
			trgVer:       "1.1.1",
			match:        false,
			expectsError: true,
		},
		{
			name:         "wildcard 1",
			srcVer:       "1.1.1",
			trgVer:       "*",
			match:        true,
			expectsError: false,
		},
		{
			name:         "wildcard 2 invalid",
			srcVer:       "1.1.1",
			trgVer:       "1*",
			match:        false,
			expectsError: true,
		},
		{
			name:         "wildcard 3",
			srcVer:       "1.1.1",
			trgVer:       "1.*",
			match:        true,
			expectsError: false,
		},
		{
			name:         "wildcard 4 invalid",
			srcVer:       "1.1.1",
			trgVer:       "1.1.1.*",
			match:        false,
			expectsError: true,
		},
		{
			name:         "patch 1",
			srcVer:       "1.1.3",
			trgVer:       "~1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "patch 2",
			srcVer:       "1.2.3",
			trgVer:       "~1.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "fix 1",
			srcVer:       "1.2.3",
			trgVer:       "^1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "fix 2",
			srcVer:       "2.2.3",
			trgVer:       "^1.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "gt 1",
			srcVer:       "2.2.3",
			trgVer:       ">=1.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "gt 2",
			srcVer:       "2.2.3",
			trgVer:       ">=3.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "lt 1",
			srcVer:       "2.2.3",
			trgVer:       "<=1.1.1",
			match:        false,
			expectsError: false,
		},
		{
			name:         "lt 2",
			srcVer:       "2.2.3",
			trgVer:       "<=3.1.1",
			match:        true,
			expectsError: false,
		},
		{
			name:         "wildcard lt",
			srcVer:       "2.2.3",
			trgVer:       "<=1.*",
			match:        false,
			expectsError: false,
		},
		{
			name:         "wildcard gt",
			srcVer:       "2.2.3",
			trgVer:       ">=1.*",
			match:        true,
			expectsError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rz, err := isVersionMatching(test.srcVer, test.trgVer)
			assert.Equal(t, rz, test.match)
			fmt.Println("match:", rz, ";; err:", err)
			assert.True(t, (!test.expectsError && err == nil) || (test.expectsError && err != nil))
		})
	}
}
