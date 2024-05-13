package ifgroup

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/device"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vishvananda/netlink"
)

func TestNetlinkManager_SetUnset(t *testing.T) {
	category.Set(t, category.Link)

	devices, err := device.ListPhysical()
	require.NoError(t, err)
	require.Greater(t, len(devices), 0)

	groups := map[int]uint32{}

	// Save the groups before set
	for _, d := range devices {
		link, err := netlink.LinkByIndex(d.Index)
		require.NoError(t, err)
		var group uint32 = 0
		attrs := link.Attrs()
		if attrs != nil {
			group = attrs.Group
		}
		groups[d.Index] = group
	}

	// Set the groups
	manager := NewNetlinkManager(device.ListPhysical)
	assert.NoError(t, manager.Set())

	// Check if set happened correctly
	for _, d := range devices {
		link, err := netlink.LinkByIndex(d.Index)
		assert.NoError(t, err)
		attrs := link.Attrs()
		assert.NotNil(t, attrs)
		if attrs != nil {
			assert.Equal(t, Group, attrs.Group)
		}
	}

	// Unset groups
	assert.NoError(t, manager.Unset())
	assert.NoError(t, manager.Set())

	// Check if unset happened correctly
	for _, d := range devices {
		link, err := netlink.LinkByIndex(d.Index)
		assert.NoError(t, err)
		attrs := link.Attrs()
		assert.NotNil(t, attrs)
		if attrs != nil {
			assert.Equal(t, groups[d.Index], attrs.Group)
		}
	}
}
