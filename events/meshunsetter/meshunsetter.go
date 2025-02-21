// Package MeshUnsetter responsible for unsetting meshnet if got 404 on api request
package meshunsetter

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
)

type MeshUnsetter interface {
	UnSetMesh() error
}

type Meshnet struct {
	man                      config.Manager
	netw                     MeshUnsetter
	meshPrivateKeyController meshnet.PrivateKeyController
	errPublisher             events.Publisher[error]
	norduser                 service.NorduserFileshareClient
}

func NewMeshnet(
	man config.Manager,
	netw MeshUnsetter,
	errPublisher events.Publisher[error],
	norduser service.NorduserFileshareClient,
	meshPrivateKeyController meshnet.PrivateKeyController,
) *Meshnet {
	return &Meshnet{
		man:                      man,
		netw:                     netw,
		errPublisher:             errPublisher,
		norduser:                 norduser,
		meshPrivateKeyController: meshPrivateKeyController,
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

	if err := m.norduser.StopFileshare(cfg.Meshnet.EnabledByUID); err != nil {
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

	m.meshPrivateKeyController.ClearMeshPrivateKey()

	return m.man.SaveWith(func(c config.Config) config.Config {
		c.Mesh = false
		c.MeshDevice = nil
		return c
	})
}
