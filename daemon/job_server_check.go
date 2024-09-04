package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/networker"
)

// JobServerCheck marks servers as offline if connection to them drops
func JobServerCheck(
	dm *DataManager,
	api core.ServersAPI,
	netw networker.Networker,
	server core.Server,
) func() {
	return func() {
		if netw.IsVPNActive() {
			srv, err := api.Server(server.ID)
			if err != nil || srv == nil {
				return
			}

			err = dm.UpdateServerPenalty(*srv)
			if err != nil {
				return
			}

			err = dm.SetServerStatus(server, server.Status)
			if err != nil {
				return
			}
		}
	}
}
