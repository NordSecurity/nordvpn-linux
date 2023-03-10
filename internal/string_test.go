package internal

import (
	"sort"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
)

func TestTitle(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input, expected string
	}{
		{"a good title", "A_Good_Title"},
		{"  weirdly formatted   title ", "Weirdly_Formatted_Title"},
		{"extra- symbols-title,!", "Extra-_Symbols-Title,!"},
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

func TestSetToStrings(t *testing.T) {
	category.Set(t, category.Unit)

	tests := [][]string{
		{"one", "two", "three"},
		{"o", "o", "o", "n", "p"},
		{"some", "text", "other", "text"},
		{"a", "b", "a", "b"},
	}

	var nilSet mapset.Set = nil
	emptySlice := SetToStrings(nilSet)
	assert.Empty(t, emptySlice)

	for _, stringSlice := range tests {
		set := mapset.NewSet()
		addStringsToSet(&set, stringSlice)

		// sets don't have repeating elements
		stringSlice = uniqueString(stringSlice)

		got := SetToStrings(set)
		sort.Strings(stringSlice)
		sort.Strings(got)

		assert.EqualValues(t, stringSlice, got)
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

func addStringsToSet(set *mapset.Set, values []string) {
	for _, val := range values {
		(*set).Add(val)
	}
}

func uniqueString(input []string) []string {
	unique := make([]string, 0, len(input))
	sliceMap := make(map[string]bool)

	for _, val := range input {
		if _, ok := sliceMap[val]; !ok {
			sliceMap[val] = true
			unique = append(unique, val)
		}
	}

	return unique
}

func TestIntsToStrings(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Equal(t, []string{"32", "56565656", "0", "1", "-1"}, IntsToStrings([]int{32, 56565656, 0, 1, -1}))
	assert.Nil(t, IntsToStrings([]int{}))
	assert.Nil(t, IntsToStrings(nil))
}
