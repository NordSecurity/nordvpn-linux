package tray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestIsGuiAvailableUnderSnap(t *testing.T) {
	category.Set(t, category.Unit)

	t.Setenv(snapconf.EnvSnapName, "nordvpn")

	assert.True(t, isGuiAvailable())
}
