package client

import (
	"encoding/json"

	mapset "github.com/deckarep/golang-set"
)

func SetToInt64s(set mapset.Set) []int64 {
	var ints []int64
	if set == nil {
		return ints
	}
	for item := range set.Iter() {
		ints = append(ints, InterfaceToInt64(item))
	}
	return ints
}

func InterfacesToInt64s(interfaces []interface{}) []int64 {
	var ints []int64
	if interfaces == nil {
		return ints
	}
	for _, item := range interfaces {
		ints = append(ints, InterfaceToInt64(item))
	}
	return ints
}

func InterfaceToInt64(item interface{}) int64 {
	var i int64
	switch item.(type) {
	case json.Number:
		i, _ = item.(json.Number).Int64()
	case int64:
		i = item.(int64)
	}
	return i
}
