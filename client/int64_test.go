package client

import (
	"encoding/json"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

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
