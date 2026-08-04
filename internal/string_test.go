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
