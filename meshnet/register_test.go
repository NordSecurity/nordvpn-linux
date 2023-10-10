package meshnet

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const privateKey = "0001"

type generator struct {
	KeyGenerator
}

func (*generator) Public(string) string { return "0001" }
func (*generator) Private() string      { return privateKey }

type registry struct {
	mesh.Registry
}

const registryUUID = "00000000-0000-0000-0000-000000000001"
const registryIP = "0.0.0.1"

func (r *registry) Register(token string, self mesh.Machine) (*mesh.Machine, error) {
	return &mesh.Machine{
		ID:      uuid.MustParse(registryUUID),
		Address: netip.MustParseAddr(registryIP),
	}, nil
}

func TestRegister_NotYetRegistered(t *testing.T) {
	cm := &mock.ConfigManager{}
	rc := NewRegisteringChecker(cm, &generator{}, &registry{})
	err := rc.Register()
	assert.NoError(t, err)
	assert.Equal(t, privateKey, cm.Cfg.MeshPrivateKey)
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
}

func TestRegister_AlreadyRegistered(t *testing.T) {
	cm := &mock.ConfigManager{
		Cfg: &config.Config{
			MeshPrivateKey: "0002",
			MeshDevice: &mesh.Machine{
				ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Address: netip.MustParseAddr("0.0.0.2"),
			},
		},
	}
	rc := NewRegisteringChecker(cm, &generator{}, &registry{})
	err := rc.Register()
	assert.NoError(t, err)
	assert.NotEqual(t, privateKey, cm.Cfg.MeshPrivateKey) // Existing private key should be kept
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
}

func TestIsRegistered_NotYetRegistered(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		newPrivateKey bool
	}{
		{
			name:          "empty config",
			cfg:           &config.Config{},
			newPrivateKey: true,
		},
		{
			name: "no private key",
			cfg: &config.Config{
				MeshDevice: &mesh.Machine{
					ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					Address: netip.MustParseAddr("0.0.0.2"),
				}},
			newPrivateKey: true,
		},
		{
			name: "no MeshDevice",
			cfg: &config.Config{
				MeshPrivateKey: "0002",
			},
		},
		{
			name: "no ID",
			cfg: &config.Config{
				MeshPrivateKey: "0002",
				MeshDevice: &mesh.Machine{
					Address: netip.MustParseAddr("0.0.0.2"),
				}},
		},
		{
			name: "no address",
			cfg: &config.Config{
				MeshPrivateKey: "0002",
				MeshDevice: &mesh.Machine{
					ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := &mock.ConfigManager{Cfg: test.cfg}
			rc := NewRegisteringChecker(cm, &generator{}, &registry{})
			ok := rc.IsRegistered()
			assert.True(t, ok)
			assert.Equal(t, test.newPrivateKey, privateKey == cm.Cfg.MeshPrivateKey)
			assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
			assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
		})
	}
}

func TestIsRegistered_AlreadyRegistered(t *testing.T) {
	cm := &mock.ConfigManager{
		Cfg: &config.Config{
			MeshPrivateKey: "0002",
			MeshDevice: &mesh.Machine{
				ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Address: netip.MustParseAddr("0.0.0.2"),
			},
		},
	}
	rc := NewRegisteringChecker(cm, &generator{}, &registry{})
	ok := rc.IsRegistered()
	assert.True(t, ok)
	// Registration should not be done, values should not change
	assert.NotEqual(t, privateKey, cm.Cfg.MeshPrivateKey)
	assert.NotEqual(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.NotEqual(t, registryIP, cm.Cfg.MeshDevice.Address.String())
}
