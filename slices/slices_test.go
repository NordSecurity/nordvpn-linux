package slices

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestContainsFunc(t *testing.T) {
	category.Set(t, category.Unit)

	isEven := func(i int) bool { return i%2 == 0 }

	tests := []struct {
		name     string
		given    []int
		expected bool
	}{
		{
			name:     "empty slice",
			given:    []int{},
			expected: false,
		},
		{
			name:     "nil slice",
			given:    nil,
			expected: false,
		},
		{
			name:     "contains",
			given:    []int{3, 14},
			expected: true,
		},
		{
			name:     "doesn't contain",
			given:    []int{3, 13},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ContainsFunc(test.given, isEven))
		})
	}
}

func TestFilter(t *testing.T) {
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
