package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/recents"
	"github.com/NordSecurity/nordvpn-linux/events"
)

// StorePendingRecentConnection stores current pending recent connection to memory
func StorePendingRecentConnection(
	store *recents.RecentConnectionsStore,
	eventPublisher func(events.DataRecentsChanged)) {
	exists, recentModel := store.GetPending()
	if !exists {
		return
	}

	if store.Add(recentModel) == nil {
		eventPublisher(events.DataRecentsChanged{})
	}
}
