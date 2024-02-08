// Package MeshUnsetter responsible for unsetting meshnet if got 404 on api request
package meshunsetter

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/fileshare_process"
)

type MeshUnsetter interface {
	UnSetMesh() error
}

type Meshnet struct {
	man          config.Manager
	netw         MeshUnsetter
	errPublisher events.Publisher[error]
	fileshare    fileshare_process.FileshareProcess
}

func NewMeshnet(
	man config.Manager,
	netw MeshUnsetter,
	errPublisher events.Publisher[error],
	fileshare fileshare_process.FileshareProcess,
) *Meshnet {
	return &Meshnet{
		man:          man,
		netw:         netw,
		errPublisher: errPublisher,
		fileshare:    fileshare,
	}
}

func (m *Meshnet) NotifyDisabled(any) error {
	return m.unsetMesh()
}

// NotifySelfRemoved unsets meshnet.
func (m *Meshnet) NotifySelfRemoved(any) error {
	return m.unsetMesh()
}

func (m *Meshnet) unsetMesh() error {
	var cfg config.Config
	if err := m.man.Load(&cfg); err != nil {
		return err
	}

	m.fileshare.Disable()

	if err := m.netw.UnSetMesh(); err != nil {
		m.errPublisher.Publish(fmt.Errorf(
			"unsetting meshnet: %w",
			err,
		))
	}

	return m.man.SaveWith(func(c config.Config) config.Config {
		c.Mesh = false
		c.MeshDevice = nil
		c.MeshPrivateKey = ""
		return c
	})
}
