/*
Package refresher is responsible for refreshing application state on
specific events.
*/
package refresher

import (
	"errors"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/internal"
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
func (m *Meshnet) NotifyPeerUpdate(peerIds []string) error {
	var cfg config.Config
	if err := m.man.Load(&cfg); err != nil {
		return err
	}

	if !cfg.Mesh {
		return errors.New("meshnet not enabled")
	}

	if !m.checker.IsRegistrationInfoCorrect() {
		return errors.New("not registered to meshnet")
	}

	token := cfg.TokensData[cfg.AutoConnectData.ID].Token
	resp, err := m.api.Map(token, cfg.MeshDevice.ID)
	if err != nil {
		return err
	}

	if internal.Contains(peerIds, cfg.MeshDevice.ID.String()) && !cfg.MeshDevice.IsEqual(resp.Machine) {
		// update info about current device when meshnet info are different
		log.Println(internal.InfoPrefix, "update current machine information")
		err := m.man.SaveWith(func(c config.Config) config.Config {
			c.MeshDevice = &resp.Machine
			return c
		})
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to save new machine information", err)
		}
	}
	// TODO: check if this should not be called only when current machine is affected
	return m.netw.Refresh(*resp)
}
