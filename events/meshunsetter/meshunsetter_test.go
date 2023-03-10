package meshunsetter

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/meshnet/mock"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type okUnsetter struct{}

func (okUnsetter) UnSetMesh() error {
	return nil
}

type failingUnsetter struct{}

func (failingUnsetter) UnSetMesh() error {
	return fmt.Errorf("error")
}

type okPublisher struct{}

func (okPublisher) Publish(error) {}

type okConfigManager struct {
	called bool
}

func (m *okConfigManager) SaveWith(config.SaveFunc) error {
	m.called = true
	return nil
}
func (m *okConfigManager) Load(*config.Config) error { return nil }
func (m *okConfigManager) Reset() error              { return nil }

type failingFileshare struct{ meshnet.Fileshare }

func (failingFileshare) Disable(uint32, uint32) error { return fmt.Errorf("error") }

func TestMeshUnsetter_unsetMesh(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name      string
		netw      MeshUnsetter
		fileshare meshnet.Fileshare
	}{
		{
			name:      "no fail",
			netw:      okUnsetter{},
			fileshare: mock.Fileshare{},
		},
		{
			name: "meshnet unset fails but config " +
				"is still updated",
			netw:      failingUnsetter{},
			fileshare: mock.Fileshare{},
		},
		{
			name:      "fileshare disable fails but config is still updated",
			netw:      okUnsetter{},
			fileshare: failingFileshare{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cm := &okConfigManager{}
			unsetter := NewMeshnet(
				cm,
				tt.netw,
				okPublisher{},
				tt.fileshare,
			)
			err := unsetter.unsetMesh()
			assert.NoError(t, err)
			assert.True(t, cm.called)
		})
	}
}
