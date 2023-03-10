// Package MeshUnsetter responsible for unsetting meshnet if got 404 on api request
package meshunsetter

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
)

type MeshUnsetter interface {
	UnSetMesh() error
}

type Meshnet struct {
	man          config.Manager
	netw         MeshUnsetter
	errPublisher events.Publisher[error]
	fileshare    meshnet.Fileshare
}

func NewMeshnet(
	man config.Manager,
	netw MeshUnsetter,
	errPublisher events.Publisher[error],
	fileshare meshnet.Fileshare,
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

	if err := m.fileshare.Disable(cfg.Meshnet.EnabledByUID, cfg.Meshnet.EnabledByGID); err != nil {
		m.errPublisher.Publish(fmt.Errorf(
			"disabling fileshare: %w",
			err,
		))
	}

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
