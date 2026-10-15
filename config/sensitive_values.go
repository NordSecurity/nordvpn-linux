package config

import (
	"slices"
)

// sensitiveServerGroups lists server groups whose connections require stripping
// sensitive fields from analytics.
var sensitiveServerGroups = []ServerGroup{
	ServerGroup_DEDICATED_IP,
	ServerGroup_DEDICATED_SERVER,
}

// hasSensitiveServerGroup reports whether any group the chosen server belongs to
// is on the sensitive list.
// Note. Independent of the user-requested group.
func HasSensitiveServerGroup(groups []ServerGroup) bool {
	for _, g := range groups {
		if slices.Contains(sensitiveServerGroups, g) {
			return true
		}
	}
	return false
}
