package daemon

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// DedicatedServerState publishes dedicated servers state
type DedicatedServerState struct {
	dedicatedServerStatusPublisher events.Publisher[events.DataDedicatedServerStatus]
	dedicatedServersAPI            core.DedicatedServersAPI
}

func NewDedicatedServerState(dedicatedServerStatusPublisher events.Publisher[events.DataDedicatedServerStatus],
	dedicatedServersAPI core.DedicatedServersAPI) *DedicatedServerState {
	return &DedicatedServerState{
		dedicatedServerStatusPublisher: dedicatedServerStatusPublisher,
		dedicatedServersAPI:            dedicatedServersAPI,
	}
}

// NotifyDedicatedServerStateChange fetches dedicated servers list and publishes dedicated server's status
func (d *DedicatedServerState) NotifyDedicatedServerStateChange(any) error {
	dedicatedServers, err := d.dedicatedServersAPI.DedicatedServers()
	if err != nil {
		return fmt.Errorf(
			"failed to fetch dedicated servers after receiving dedicated servers state change notification: %w", err)
	}

	if len(dedicatedServers) == 0 {
		return errors.New("received zero-length dedicated servers list from the API")
	}

	// currently there can be only one dedicated server
	dedicatedServer := dedicatedServers[0]
	d.dedicatedServerStatusPublisher.Publish(events.DataDedicatedServerStatus{
		Status: string(dedicatedServer.Status),
	})

	return nil
}
