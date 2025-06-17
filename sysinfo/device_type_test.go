package sysinfo

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_detectBySystemDefaultTarget(t *testing.T) {
	category.Set(t, category.Integration)

	devType := detectBySystemDefaultTarget()
	assert.NotEqual(t, SystemDeviceTypeUnknown, devType, "device type must be known")
}

func Test_detectByXDGSession(t *testing.T) {
	category.Set(t, category.Integration)

	devType := detectByXDGSession()
	assert.NotEqual(t, SystemDeviceTypeUnknown, devType, "device type must be known")
}

func Test_detectByGraphicalEnv(t *testing.T) {
	category.Set(t, category.Integration)

	devType := detectByGraphicalEnv()
	assert.NotEqual(t, SystemDeviceTypeServer, devType, "device cannot be determined as a server")
}

func TestGetDeviceType(t *testing.T) {
	category.Set(t, category.Integration)

	devType := DeviceType()
	assert.NotEqual(t, SystemDeviceTypeUnknown, devType, "device type must be known")
}
