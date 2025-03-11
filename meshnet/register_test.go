package meshnet

import (
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
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
	category.Set(t, category.Unit)

	cm := &mock.ConfigManager{}
	rc := NewRegisteringChecker(cm, &generator{}, &registry{})
	err := rc.Register()
	assert.NoError(t, err)
	meshPK, ok := rc.GetMeshPrivateKey()
	assert.True(t, ok)
	assert.Equal(t, privateKey, meshPK)
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
}

func TestRegister_AlreadyRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	cm := &mock.ConfigManager{
		Cfg: &config.Config{
			MeshDevice: &mesh.Machine{
				ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Address: netip.MustParseAddr("0.0.0.2"),
			},
		},
	}
	rc := NewRegisteringChecker(cm, &generator{}, &registry{})
	rc.meshPrivateKey = "0002"
	err := rc.Register()
	assert.NoError(t, err)

	meshPK, ok := rc.GetMeshPrivateKey()
	assert.True(t, ok)
	assert.Equal(t, privateKey, meshPK) // New private key should be generated
	assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
}

func TestIsRegistered_NotYetRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		cfg            *config.Config
		meshPrivateKey string
	}{
		{
			name: "empty config",
			cfg:  &config.Config{},
		},
		{
			name: "no private key",
			cfg: &config.Config{
				MeshDevice: &mesh.Machine{
					ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					Address: netip.MustParseAddr("0.0.0.2"),
				}},
		},
		{
			name:           "no MeshDevice",
			cfg:            &config.Config{},
			meshPrivateKey: "0002",
		},
		{
			name: "no ID",
			cfg: &config.Config{
				MeshDevice: &mesh.Machine{
					Address: netip.MustParseAddr("0.0.0.2"),
				}},
			meshPrivateKey: "0002",
		},
		{
			name: "no address",
			cfg: &config.Config{
				MeshDevice: &mesh.Machine{
					ID: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				}},
			meshPrivateKey: "0002",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := &mock.ConfigManager{Cfg: test.cfg}
			rc := NewRegisteringChecker(cm, &generator{}, &registry{})
			rc.meshPrivateKey = test.meshPrivateKey
			ok := rc.IsRegistrationInfoCorrect()
			assert.True(t, ok)

			meshPrivateKey, ok := rc.GetMeshPrivateKey()
			assert.True(t, ok)
			assert.Equal(t, privateKey, meshPrivateKey)
			assert.Equal(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
			assert.Equal(t, registryIP, cm.Cfg.MeshDevice.Address.String())
		})
	}
}

func TestIsRegistered_AlreadyRegistered(t *testing.T) {
	category.Set(t, category.Unit)

	cm := &mock.ConfigManager{
		Cfg: &config.Config{
			MeshDevice: &mesh.Machine{
				ID:      uuid.MustParse("00000000-0000-0000-0000-000000000002"),
				Address: netip.MustParseAddr("0.0.0.2"),
			},
		},
	}
	rc := NewRegisteringChecker(cm, &generator{}, &registry{})
	rc.meshPrivateKey = "0002"
	ok := rc.IsRegistrationInfoCorrect()
	assert.True(t, ok)

	// Registration should not be done, values should not change
	meshPrivateKey, ok := rc.GetMeshPrivateKey()
	assert.True(t, ok)
	assert.NotEqual(t, privateKey, meshPrivateKey)
	assert.NotEqual(t, registryUUID, cm.Cfg.MeshDevice.ID.String())
	assert.NotEqual(t, registryIP, cm.Cfg.MeshDevice.Address.String())
}
