package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestisGUIAvailableUnderSnap(t *testing.T) {
	category.Set(t, category.Unit)

	t.Setenv("PATH", t.TempDir())

	t.Setenv(snapconf.EnvSnapName, "")
	assert.False(t, isGUIAvailable())

	t.Setenv(snapconf.EnvSnapName, "nordvpn")
	assert.True(t, isGUIAvailable())
}
