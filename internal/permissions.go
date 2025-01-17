package internal

import (
	"fmt"
	"os/user"
)

var allowedGroups []string = []string{"nordvpn"}
var ErrNoPermission error = fmt.Errorf("requesting user does not have permissions")

// IsInAllowedGroup returns true if user with the given UID is in nordvpn privileged group
func IsInAllowedGroup(uid uint32) (bool, error) {
	userInfo, err := user.LookupId(fmt.Sprintf("%d", uid))
	if err != nil {
		return false, fmt.Errorf("authenticate user, lookup user info: %s", err)
	}
	// user belongs to the allowed group?
	groups, err := userInfo.GroupIds()
	if err != nil {
		return false, fmt.Errorf("authenticate user, check user groups: %s", err)
	}

	for _, groupId := range groups {
		groupInfo, err := user.LookupGroupId(groupId)
		if err != nil {
			return false, fmt.Errorf("authenticate user, check user group: %s", err)
		}
		for _, allowGroupName := range allowedGroups {
			if groupInfo.Name == allowGroupName {
				return true, nil
			}
		}
	}

	return false, nil
}
