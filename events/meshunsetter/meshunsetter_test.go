package meshunsetter

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"

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
func (m *okConfigManager) Reset(bool) error          { return nil }

func TestMeshUnsetter_unsetMesh(t *testing.T) {
	category.Set(t, category.Unit)
	for _, tt := range []struct {
		name              string
		netw              MeshUnsetter
		startFileshareErr error
	}{
		{
			name:              "no fail",
			netw:              okUnsetter{},
			startFileshareErr: nil,
		},
		{
			name: "meshnet unset fails but config " +
				"is still updated",
			netw:              failingUnsetter{},
			startFileshareErr: nil,
		},
		{
			name:              "fileshare disable fails but config is still updated",
			netw:              okUnsetter{},
			startFileshareErr: fmt.Errorf("failed to start fileshare"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cm := &okConfigManager{}
			unsetter := NewMeshnet(
				cm,
				tt.netw,
				okPublisher{},
				testnorduser.NewMockNorduserClient(tt.startFileshareErr),
			)
			err := unsetter.unsetMesh()
			assert.NoError(t, err)
			assert.True(t, cm.called)
		})
	}
}
