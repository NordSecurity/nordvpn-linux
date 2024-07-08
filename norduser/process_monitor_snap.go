package norduser

import (
	"fmt"
	"log"
	"slices"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fsnotify/fsnotify"
)

func (n *NorduserProcessMonitor) stopForDeletedGroupMembers(currentGroupMembers []string,
	newGroupMembers []string) []string {
	groupMembersUpdate := []string{}
	for _, username := range currentGroupMembers {
		if slices.Contains(newGroupMembers, username) {
			groupMembersUpdate = append(groupMembersUpdate, username)
			continue
		}

		userIDs, err := n.userIDGetter.getUserID(username)
		if err != nil {
			log.Println(internal.ErrorPrefix, "getting users ID:", err)
			groupMembersUpdate = append(groupMembersUpdate, username)
			continue
		}

		if err := n.norduserd.Stop(userIDs.uid, false); err != nil {
			groupMembersUpdate = append(groupMembersUpdate, username)
			log.Println(internal.ErrorPrefix, "stopping norduser:", err)
		}
	}

	return groupMembersUpdate
}

// StartSnap starts a simplified norduser process monitor routine. norduser processes will be stopped for users removed
// form the nordvpn group, no other actions will be taken. Because of snap, starting/restarting the process has to be
// handled in the process itself.
func (n *NorduserProcessMonitor) StartSnap() error {
	watcher, err := getWatcher(etcPath, utmpFilePath)
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer watcher.Close()

	groupMembers, err := getNordVPNGroupMembers()
	if err != nil {
		return fmt.Errorf("getting initial nordvpn group members: %w", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("groupfile monitor channel closed")
			}

			if event.Name == groupFilePath {
				// Because utilities used to modify the group do so atomically, we also need to monitor for creation of
				// the file instead of modifications.
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					newGroupMembers, err := getNordVPNGroupMembers()
					if err != nil {
						log.Println(internal.ErrorPrefix, "getting new group members:", err)
					} else {
						groupMembers = n.stopForDeletedGroupMembers(groupMembers, newGroupMembers)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("groupfile monitor error channel closed")
			}
			log.Println(internal.ErrorPrefix, "group monitor error:", err)
		}
	}
}
