package internal

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestTitle(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input, expected string
	}{
		{"a good title", "A_Good_Title"},
		{"  weirdly formatted   title ", "Weirdly_Formatted_Title"},
		{"extra- symbols-title,!", "Extra-_Symbols-Title"},
	}
	for _, item := range tests {
		got := Title(item.input)
		assert.Equal(t, item.expected, got)
	}
}

func TestSnakeCase(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input, expected string
	}{
		{"A Good Title", "a_good_title"},
		{"  Weirdly Formatted   Text", "weirdly_formatted_text"},
	}
	for _, item := range tests {
		got := SnakeCase(item.input)
		assert.Equal(t, item.expected, got)
	}
}

func TestStringsToInterfaces(t *testing.T) {
	category.Set(t, category.Unit)

	tests := [][]string{
		{"a", "b", "c", "d"},
		{"a", "a", "a", "b", "b", "banana"},
		{"item", "item2", "item3", "item"},
	}

	for _, item := range tests {
		got := StringsToInterfaces(item)
		for i := range got {
			c, ok := got[i].(string)
			assert.True(t, ok)
			assert.Equal(t, c, item[i])
		}
	}
}

func TestStringsContains(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    []string
		find     string
		expected bool
	}{
		{
			input:    []string{"good", "food", "hood"},
			find:     "food",
			expected: true,
		},
		{
			input:    []string{"bad", "mad", "lad"},
			find:     "id",
			expected: false,
		},
	}

	for _, tt := range tests {
		got := StringsContains(tt.input, tt.find)
		assert.Equal(t, tt.expected, got)
	}
}

func TestStringsGetNext(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    []string
		find     string
		expected string
	}{
		{
			input:    []string{"good", "food", "hood"},
			find:     "food",
			expected: "hood",
		},
	}

	for _, tt := range tests {
		got := StringsGetNext(tt.input, tt.find)
		assert.Equal(t, tt.expected, got)
	}
}

func TestIntsToStrings(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, []string{"32", "56565656", "0", "1", "-1"}, IntsToStrings([]int{32, 56565656, 0, 1, -1}))
	assert.Nil(t, IntsToStrings([]int{}))
	assert.Nil(t, IntsToStrings(nil))
}
