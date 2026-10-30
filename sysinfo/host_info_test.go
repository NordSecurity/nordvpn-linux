package sysinfo

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_newHostInfo(t *testing.T) {
	category.Set(t, category.Unit)

	const (
		fullOSRelease = "NAME=\"Ubuntu\"\nPRETTY_NAME=\"Ubuntu 24.04.1 LTS\""
		nameOSRelease = "NAME=\"Ubuntu\""
	)

	tests := []struct {
		name     string
		src      hostInfoSources
		expected HostInfo
	}{
		{
			name: "all sources available",
			src: hostInfoSources{
				osRelease:    release(nameOSRelease),
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
			name: "pretty name falls back to os-release pretty name on error",
			src: hostInfoSources{
				osRelease:    release(fullOSRelease),
				osPrettyName: fail,
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
			name: "pretty name falls back to os-release pretty name when empty",
			src: hostInfoSources{
				osRelease:    release(fullOSRelease),
				osPrettyName: value(""),
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
			name: "pretty name falls back to OS name when missing in os-release",
			src: hostInfoSources{
				osRelease:    release(nameOSRelease),
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
			name: "pretty name falls back to OS name when empty in os-release",
			src: hostInfoSources{
				osRelease:    release(nameOSRelease + "\nPRETTY_NAME=\"\""),
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
				osRelease:    func() (osRelease, error) { return nil, errors.New("fail") },
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

func release(content string) func() (osRelease, error) {
	return func() (osRelease, error) { return osRelease(content), nil }
}
func value(v string) func() (string, error) {
	return func() (string, error) { return v, nil }
}
func fail() (string, error)     { return "", errors.New("fail") }
func x86() string               { return "x86_64" }
func desktop() SystemDeviceType { return SystemDeviceTypeDesktop }
