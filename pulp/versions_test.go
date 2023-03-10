package pulp

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestAddPrefix(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		given    string
		expected string
	}{
		{
			name:     "empty string",
			expected: "v",
		},
		{
			name:     "with prefix",
			given:    "v1.2.3",
			expected: "v1.2.3",
		},
		{
			name:     "without prefix",
			given:    "1.2.3",
			expected: "v1.2.3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, addPrefix(test.given))
		})
	}
}

func TestRemovePrefix(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		given    string
		expected string
	}{
		{
			name: "empty string",
		},
		{
			name:     "with prefix",
			given:    "v1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "without prefix",
			given:    "1.2.3",
			expected: "1.2.3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, removePrefix(test.given))
		})
	}
}

func TestTransform(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		given    []string
		expected []string
	}{
		{
			name:     "nil slice",
			expected: []string{},
		},
		{
			name:     "empty slice",
			given:    []string{},
			expected: []string{},
		},
		{
			name:     "emptify all elements",
			given:    []string{"v1.2.3", "v1.2.4", "v1.2.5"},
			expected: []string{"", "", ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := transform(test.given, func(string) string { return "" })
			assert.ElementsMatch(t, test.expected, got)
		})
	}
}

func TestUnique(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		given    []string
		expected []string
	}{
		{
			name: "nil slice",
		},
		{
			name:     "empty slice",
			given:    []string{},
			expected: []string{},
		},
		{
			name:     "contains duplicates",
			given:    []string{"", "", "", " ", " "},
			expected: []string{"", " "},
		},
		{
			name:     "already unique",
			given:    []string{"", " "},
			expected: []string{"", " "},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ElementsMatch(t, test.expected, unique(test.given))
		})
	}
}

func TestLast(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		given    []string
		count    uint
		expected []string
	}{
		{
			name: "nil slice",
		},
		{
			name:     "empty slice",
			given:    []string{},
			expected: []string{},
		},
		{
			name:     "exceed elements",
			given:    []string{"1"},
			count:    2,
			expected: []string{"1"},
		},
		{
			name:     "happy path",
			given:    []string{"1", "2", "3", "4"},
			count:    2,
			expected: []string{"3", "4"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ElementsMatch(t, test.expected, last(test.given, test.count))
		})
	}
}

func TestDeleteFrom(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		given    []string
		count    uint
		expected []string
	}{
		{
			name: "return list with 0 less minor versions",
			given: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
				"3.1.2", "3.2.1", "2.11",
			},
			expected: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
				"3.1.2", "3.2.1", "2.11",
			},
		},
		{
			name: "return list with 1 less minor versions",
			given: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
				"3.1.2", "3.2.1", "2.11",
			},
			count: 1,
			expected: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
				"3.1.2", "2.11",
			},
		},
		{
			name: "return list with 3 less minor versions",
			given: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
				"3.1.2", "3.2.1", "2.11",
			},
			count: 3,
			expected: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
			},
		},
		{
			name: "return list with 5 less minor versions",
			given: []string{
				"1.2.3", "1.3.2",
				"2.1.3", "2.3.1",
				"2.7",
				"3.1.2", "3.2.1", "2.11",
			},
			count: 5,
			expected: []string{
				"1.2.3", "1.3.2",
				"2.1.3",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ElementsMatch(t, test.expected, deleteFrom(test.given, test.count))
		})
	}
}
