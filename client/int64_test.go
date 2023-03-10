package client

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	mapset "github.com/deckarep/golang-set"
	"github.com/stretchr/testify/assert"
)

func TestSetToInt64s(t *testing.T) {
	category.Set(t, category.Unit)

	tests := [][]int64{
		{1, 2, 45, 3, 7, 1603},
		{32, 32, 32, 32, 32},
		{12, 32, 32, 13, 14},
		{1, 2, 3, 4, 5, 5, 6, 7},
		{20, 21, 22, 24, 25},
	}

	var nilSet mapset.Set = nil
	emptySlice := SetToInt64s(nilSet)
	assert.Empty(t, emptySlice)

	for _, intSlice := range tests {
		set := mapset.NewSet()
		addIntsToSet(&set, intSlice)

		// sets don't have repeating elements
		intSlice = uniqueInt64(intSlice)

		got := SetToInt64s(set)
		sort.Slice(intSlice, func(i, j int) bool {
			return intSlice[i] < intSlice[j]
		})
		sort.Slice(got, func(i, j int) bool {
			return got[i] < got[j]
		})

		assert.EqualValues(t, intSlice, got)
	}
}

func TestInterfaceToInt64(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		info     interface{}
		expected int64
	}{
		{int64(123), 123},
		{json.Number("54103"), 54103},
		{"asd", 0},
		{true, 0},
		{false, 0},
		{123.5, 0},
		{int64(9223372036854775807), 9223372036854775807},
	}

	for _, item := range tests {
		got := InterfaceToInt64(item.info)
		assert.Equal(t, got, item.expected)
	}
}

func TestInterfacesToInt64s(t *testing.T) {
	category.Set(t, category.Unit)

	tests := [][]interface{}{
		{int64(123), int64(23), int64(150), int64(13)},
		{int64(12573458), int64(1), int64(1), int64(2), int64(15), int64(17)},
		{int64(1), int64(5), int64(5), int64(2), int64(7), int64(314)},
		nil,
	}

	for _, item := range tests {
		intSlice := InterfacesToInt64s(item)
		assert.Equal(t, len(intSlice), len(item))
		for j := range item {
			n, ok := item[j].(int64)
			assert.True(t, ok)
			assert.Equal(t, n, intSlice[j])
		}
	}
}

func addIntsToSet(set *mapset.Set, values []int64) {
	for _, val := range values {
		(*set).Add(val)
	}
}

func uniqueInt64(input []int64) []int64 {
	unique := make([]int64, 0, len(input))
	sliceMap := make(map[int64]bool)

	for _, val := range input {
		if _, ok := sliceMap[val]; !ok {
			sliceMap[val] = true
			unique = append(unique, val)
		}
	}

	return unique
}
