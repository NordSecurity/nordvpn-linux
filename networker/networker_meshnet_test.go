package networker

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events/refresher"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"gotest.tools/v3/assert"
)

// For LVPN-8087 if meshnet is disabled while an NC is received, the notifications reconfigures the FW
// Check that the system is not configured if the meshnet is not running in networker, even if in settings it is enabled
func TestHandleNCForMeshnetEnabledInSettingsButNotRunning(t *testing.T) {
	category.Set(t, category.Unit)
	cfg := &config.Config{
		Mesh: true,
		MeshDevice: &mesh.Machine{
			ID: uuid.UUID{},
		},
		AutoConnectData: config.AutoConnectData{ID: 1},
		TokensData: map[int64]config.TokenData{
			1: {Token: "xx"},
		},
	}

	meshnetMap := mesh.MachineMap{
		Machine: *cfg.MeshDevice,
	}

	cm := &mock.ConfigManager{
		Cfg: cfg,
	}

	// enable Meshnet
	combined := GetTestCombined()

	r := refresher.NewMeshnet(
		&mock.CachingMapperMock{
			Value: &meshnetMap,
		},
		mock.RegistrationCheckerMock{},
		cm,
		combined,
	)

	// NC ignored if meshnet is true in settings, but it is not running
	err := r.NotifyPeerUpdate([]string{})
	assert.ErrorIs(t, err, meshnet.ErrMeshnetNotEnabled)

	// enable meshnet
	err = combined.SetMesh(meshnetMap, netip.MustParseAddr("100.64.0.100"), "key")
	assert.NilError(t, err)

	// NC notification is received to reconfigure, meshnet is enabled
	err = r.NotifyPeerUpdate([]string{})
	assert.NilError(t, err)

	// disable meshnet
	err = combined.UnSetMesh()
	assert.NilError(t, err)

	// NC event while meshnet is still true in settings
	err = r.NotifyPeerUpdate([]string{})
	assert.ErrorIs(t, err, meshnet.ErrMeshnetNotEnabled)
}
