package daemon

import (
	"slices"

	"github.com/NordSecurity/nordvpn-linux/config"
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
		// dedicated servers are not kept on the server list, so we have to skip them
		isDedicatedServer := slices.ContainsFunc(server.Groups, func(group core.Group) bool {
			return group.ID == config.ServerGroup_DEDICATED_SERVERS
		})
		if netw.IsVPNActive() && !isDedicatedServer {
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
