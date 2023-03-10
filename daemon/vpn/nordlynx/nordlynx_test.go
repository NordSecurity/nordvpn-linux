package nordlynx

import (
	"net"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestUpWGInterface(t *testing.T) {
	category.Set(t, category.Link)

	iName := "winterface"
	_, err := net.InterfaceByName(iName)
	assert.Error(t, err)

	err = upWGInterface(iName)
	assert.NoError(t, err)

	iface, err := net.InterfaceByName(iName)
	assert.NoError(t, err)
	defer deleteInterface(*iface)
}

func TestAddDevice(t *testing.T) {
	category.Set(t, category.Link)

	t.Run("successful adding", func(t *testing.T) {
		device := "testdev"
		devType := "wireguard"
		defer removeDevice(device)

		err := addDevice(device, devType)
		assert.NoError(t, err)
	})

	t.Run("duplicate adding", func(t *testing.T) {
		device := "faildev"
		devType := "wireguard"
		defer removeDevice(device)

		err := addDevice(device, devType)
		assert.NoError(t, err)

		err = addDevice(device, devType)
		assert.Error(t, err)
	})
}

func TestRemoveDevice(t *testing.T) {
	category.Set(t, category.Link)

	t.Run("non existing device", func(t *testing.T) {
		device := "nodev"

		_, err := removeDevice(device)
		assert.Error(t, err)
	})
}
