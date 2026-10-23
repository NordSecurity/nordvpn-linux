package sysinfo

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_newHostInfo(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		src      hostInfoSources
		expected HostInfo
	}{
		{
			name: "all sources available",
			src: hostInfoSources{
				osName:       value("Ubuntu"),
				osPrettyName: value("Ubuntu 24.04.1 LTS"),
				architecture: x86,
				deviceType:   desktop,
			},
			expected: HostInfo{
				OSName:       "Ubuntu",
				OSPrettyName: "Ubuntu 24.04.1 LTS",
				Architecture: "x86_64",
				DeviceType:   SystemDeviceTypeDesktop,
			},
		},
		{
			name: "pretty name falls back to OS name on error",
			src: hostInfoSources{
				osName:       value("Ubuntu"),
				osPrettyName: fail,
				architecture: x86,
				deviceType:   desktop,
			},
			expected: HostInfo{
				OSName:       "Ubuntu",
				OSPrettyName: "Ubuntu",
				Architecture: "x86_64",
				DeviceType:   SystemDeviceTypeDesktop,
			},
		},
		{
			name: "pretty name falls back to OS name when empty",
			src: hostInfoSources{
				osName:       value("Ubuntu"),
				osPrettyName: value(""),
				architecture: x86,
				deviceType:   desktop,
			},
			expected: HostInfo{
				OSName:       "Ubuntu",
				OSPrettyName: "Ubuntu",
				Architecture: "x86_64",
				DeviceType:   SystemDeviceTypeDesktop,
			},
		},
		{
			name: "nothing available",
			src: hostInfoSources{
				osName:       fail,
				osPrettyName: fail,
				architecture: func() string { return "" },
				deviceType:   func() SystemDeviceType { return SystemDeviceTypeUnknown },
			},
			expected: HostInfo{DeviceType: SystemDeviceTypeUnknown},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, newHostInfo(tt.src))
		})
	}
}

func Test_NewHostInfo(t *testing.T) {
	category.Set(t, category.Integration)

	info := NewHostInfo()
	assert.NotEmpty(t, info.OSName)
	assert.NotEmpty(t, info.OSPrettyName)
	assert.NotEmpty(t, info.Architecture)
}

func value(v string) func() (string, error) {
	return func() (string, error) { return v, nil }
}
func fail() (string, error)     { return "", errors.New("fail") }
func x86() string               { return "x86_64" }
func desktop() SystemDeviceType { return SystemDeviceTypeDesktop }
