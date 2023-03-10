package nstrings

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestBoolFromString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input    string
		answer   bool
		err      error
		hasError bool
	}{
		{
			input:    "false",
			answer:   false,
			err:      nil,
			hasError: false,
		},
		{
			input:    "0",
			answer:   false,
			err:      nil,
			hasError: false,
		},
		{
			input:    "off",
			answer:   false,
			err:      nil,
			hasError: false,
		},
		{
			input:    "disabled",
			answer:   false,
			err:      nil,
			hasError: false,
		},
		{
			input:    "disable",
			answer:   false,
			err:      nil,
			hasError: false,
		},
		{
			input:    "true",
			answer:   true,
			err:      nil,
			hasError: false,
		},
		{
			input:    "on",
			answer:   true,
			err:      nil,
			hasError: false,
		},
		{
			input:    "enable",
			answer:   true,
			err:      nil,
			hasError: false,
		},
		{
			input:    "1",
			answer:   true,
			err:      nil,
			hasError: false,
		},
		{
			input:    "enabled",
			answer:   true,
			err:      nil,
			hasError: false,
		},
		{
			input:    "maybe",
			answer:   false,
			err:      fmt.Errorf(boolFromStringError, "maybe"),
			hasError: true,
		},
		{
			input:    "kinda",
			answer:   false,
			err:      fmt.Errorf(boolFromStringError, "kinda"),
			hasError: true,
		},
		{
			input:    "3",
			answer:   false,
			err:      fmt.Errorf(boolFromStringError, "3"),
			hasError: true,
		},
		{
			input:    "enabling",
			answer:   false,
			err:      fmt.Errorf(boolFromStringError, "enabling"),
			hasError: true,
		},
	}

	for _, test := range tests {
		got, err := BoolFromString(test.input)
		if test.hasError {
			assert.Error(t, test.err)
			assert.Equal(t, err.Error(), test.err.Error())
		}
		assert.Equal(t, test.answer, got)
	}
}

func TestCanParseFalseFromString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input  string
		result bool
	}{
		{"false", true},
		{"0", true},
		{"off", true},
		{"disabled", true},
		{"disable", true},
		{"true", false},
		{"on", false},
		{"enable", false},
		{"1", false},
		{"enabled", false},
		{"maybe", false},
		{"kinda", false},
		{"3", false},
		{"enabling", false},
	}

	for _, test := range tests {
		got := CanParseFalseFromString(test.input)
		assert.Equal(t, got, test.result)
	}
}

func TestCanParseTrueFromString(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		input  string
		result bool
	}{
		{"false", false},
		{"0", false},
		{"off", false},
		{"disabled", false},
		{"disable", false},
		{"true", true},
		{"on", true},
		{"enable", true},
		{"1", true},
		{"enabled", true},
		{"maybe", false},
		{"kinda", false},
		{"3", false},
		{"enabling", false},
	}

	for _, test := range tests {
		got := CanParseTrueFromString(test.input)
		assert.Equal(t, got, test.result)
	}
}

func TestGetBools(t *testing.T) {
	category.Set(t, category.Unit)

	boolMap := map[string]bool{
		"1":        true,
		"true":     true,
		"enable":   true,
		"enabled":  true,
		"on":       true,
		"0":        true,
		"false":    true,
		"disable":  true,
		"disabled": true,
		"off":      true,
	}

	got := GetBools()
	for _, test := range got {
		assert.Contains(t, boolMap, test)
	}
}

func TestGetBoolLabel_True(t *testing.T) {
	category.Set(t, category.Unit)

	expected := "enabled"
	got := GetBoolLabel(true)
	assert.Equal(t, got, expected)
}

func TestGetBoolLabel_False(t *testing.T) {
	category.Set(t, category.Unit)

	expected := "disabled"
	got := GetBoolLabel(false)
	assert.Equal(t, got, expected)
}
