// Package nstrings contains utility string functions (nstrings it to not confuse with go std strings)
package nstrings

import (
	"fmt"
	"strings"
)

var boolFromStringError = "bool: failed to parse from string: %s"

const (
	enabled   = "enabled"
	disabled  = "disabled"
	undefined = "undefined"
	notSet    = "not set"
)

var falseMap = map[string]bool{
	"0":        false,
	"false":    false,
	"disable":  false,
	"disabled": false,
	"off":      false,
}

var trueMap = map[string]bool{
	"1":       true,
	"true":    true,
	"enable":  true,
	"enabled": true,
	"on":      true,
}

// BoolFromString takes in the string and tries to parse it as a boolean
func BoolFromString(arg string) (bool, error) {
	arg = strings.ToLower(arg)
	if v, ok := falseMap[arg]; ok {
		return v, nil
	}
	if v, ok := trueMap[arg]; ok {
		return v, nil
	}
	return false, fmt.Errorf(boolFromStringError, arg)
}

// CanParseFalseFromString takes in the string and checks it if can be parsed as false
func CanParseFalseFromString(arg string) bool {
	arg = strings.ToLower(arg)
	_, ok := falseMap[arg]
	return ok
}

// CanParseTrueFromString takes in the string and checks it if can be parsed as true
func CanParseTrueFromString(arg string) bool {
	arg = strings.ToLower(arg)
	_, ok := trueMap[arg]
	return ok
}

// GetBools returns all supported bool values
func GetBools() []string {
	var res []string
	for k := range falseMap {
		res = append(res, k)
	}
	for k := range trueMap {
		res = append(res, k)
	}
	return res
}

// GetBoolLabel returns disabled if false, enabled otherwise
func GetBoolLabel(option bool) string {
	if option {
		return enabled
	}
	return disabled
}
