/*
Package refresher is responsible for refreshing application state on
specific events.
*/
package refresher

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	meshn "github.com/NordSecurity/nordvpn-linux/meshnet"

	"github.com/google/uuid"
)

// Mapper returns meshnet map.
type Mapper interface {
	Map(token string, self uuid.UUID) (*mesh.MachineMap, error)
}

// Refresher updates active meshnet peer list.
type Refresher interface {
	Refresh(mesh.MachineMap) error
}

// Meshnet refreshes peers.
type Meshnet struct {
	api     Mapper
	checker meshn.Checker
	man     config.Manager
	netw    Refresher
}

// NewMeshnet is a default constructor for Meshnet.
func NewMeshnet(
	api Mapper,
	checker meshn.Checker,
	man config.Manager,
	netw Refresher,
) *Meshnet {
	return &Meshnet{api: api, checker: checker, man: man, netw: netw}
}

// NotifyPeerUpdate refreshes meshnet peers.
func (m *Meshnet) NotifyPeerUpdate([]string) error {
	var cfg config.Config
	if err := m.man.Load(&cfg); err != nil {
		return err
	}

	if !cfg.Mesh {
		return errors.New("meshnet not enabled")
	}

	if !m.checker.IsRegistered() {
		return errors.New("not registered to meshnet")
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := m.api.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		return err
	}

	return m.netw.Refresh(*resp)
}
