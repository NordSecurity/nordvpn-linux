package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_Filter(t *testing.T) {
	category.Set(t, category.Unit)

	isEven := func(i int) bool { return i%2 == 0 }

	tests := []struct {
		name     string
		given    []int
		expected []int
	}{
		{
			name:     "empty slice",
			given:    []int{},
			expected: []int{},
		},
		{
			name:     "nil slice",
			given:    nil,
			expected: []int{},
		},
		{
			name:     "contains",
			given:    []int{3, 14},
			expected: []int{14},
		},
		{
			name:     "doesn't contain",
			given:    []int{3, 13},
			expected: []int{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ElementsMatch(t, test.expected, Filter(test.given, isEven))
		})
	}
}
