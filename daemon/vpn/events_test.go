package vpn

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestNewInternalVPNEvents_WiresSubjects(t *testing.T) {
	category.Set(t, category.Unit)

	internalEvents := NewInternalVPNEvents()

	assert.NotNil(t, internalEvents.Connected, "Connected subject must be initialized")
	assert.NotNil(t, internalEvents.Disconnected, "Disconnected subject must be initialized")
	assert.NotNil(t, internalEvents.ConnectionError, "ConnectionError subject must be initialized")
}
