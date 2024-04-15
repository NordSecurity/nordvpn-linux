package norduser

import (
	"fmt"
	"log"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/fsnotify/fsnotify"
)

const (
	groupfilePath = "/etc/group"
)

type userSet map[string]bool

func findGroupEntry(groups string, groupName string) string {
	r, _ := regexp.Compile(fmt.Sprintf("^%s:", groupName))

	for _, groupEntry := range strings.Split(groups, "\n") {
		if r.MatchString(groupEntry) {
			return groupEntry
		}
	}

	return ""
}

func getGroupEntry(groupName string) (string, error) {
	file, err := internal.FileRead(groupfilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read group file: %w", err)
	}

	groupEntry := findGroupEntry(string(file), groupName)
	if groupEntry == "" {
		return "", fmt.Errorf("group entry not found: %w", err)
	}

	return groupEntry, nil
}

func getGroupMembers(groupEntry string) userSet {
	groupEntry = strings.TrimSuffix(groupEntry, "\n")
	sepIndex := strings.LastIndex(groupEntry, ":")
	groupMembers := groupEntry[sepIndex+1:]
	members := make(userSet)

	if groupMembers == "" {
		return members
	}

	for _, member := range strings.Split(groupMembers, ",") {
		members[member] = false
	}

	return members
}

func getNordvpnGroupMembers() (userSet, error) {
	const nordvpnGroupName = "nordvpn"

	groupEntry, err := getGroupEntry(nordvpnGroupName)
	if err != nil {
		return nil, fmt.Errorf("getting group entry: %w", err)
	}

	groupMembers := getGroupMembers(groupEntry)

	return groupMembers, nil
}

type userIDs struct {
	uid  uint32
	gid  uint32
	home string
}

func getUID(username string) (userIDs, error) {
	user, err := user.Lookup(username)
	if err != nil {
		return userIDs{}, fmt.Errorf("looking up user: %w", err)
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return userIDs{}, fmt.Errorf("converting uid string to int: %w", err)
	}

	gid, err := strconv.Atoi(user.Gid)
	if err != nil {
		return userIDs{}, fmt.Errorf("converting uid string to int: %w", err)
	}

	return userIDs{
		uid:  uint32(uid),
		gid:  uint32(gid),
		home: user.HomeDir,
	}, nil
}

func (n *NordvpnGroupMonitor) handleGroupUpdate(currentGroupMembers userSet, newGroupMembers userSet) {
	for member := range currentGroupMembers {
		// member not modified, do nothing
		if ok := newGroupMembers[member]; ok {
			newGroupMembers[member] = true
			continue
		}

		userIDs, err := getUID(member)
		if err != nil {
			log.Println("failed to look up UID/GID of deleted group member:", err)
			continue
		}

		if err := n.norduserd.Stop(userIDs.uid); err != nil {
			log.Println("disabling norduserd for user:", err.Error())
		}
	}

	if n.isSnap {
		// in snap environment norduser is enabled on the cli side
		return
	}

	// enable norduserd for new members
	for member, norduserdEnabled := range newGroupMembers {
		if norduserdEnabled {
			continue
		}

		userID, err := getUID(member)
		if err != nil {
			log.Println("failed to lookup UID/GID for new group member:", err)
			continue
		}

		if err := n.norduserd.Enable(userID.uid, userID.gid, userID.home); err != nil {
			log.Println("enabling norduserd for member:", err)
		}
	}
}

func (n *NordvpnGroupMonitor) startForEveryGroupMember(groupMembers userSet) {
	for member := range groupMembers {
		user, err := getUID(member)
		if err != nil {
			log.Println("failed to get UID/GID for group member:", err)
			continue
		}

		if err := n.norduserd.Enable(user.uid, user.gid, user.home); err != nil {
			log.Println("failed to start norduser for group member:", err)
		}
	}
}

// NordvpnGroupMonitor monitors the nordvpn system group and starts/stops norduserd for users added/removed from the
// group.
type NordvpnGroupMonitor struct {
	norduserd service.NorduserService
	isSnap    bool
}

func NewNordvpnGroupMonitor(service service.NorduserService) NordvpnGroupMonitor {
	return NordvpnGroupMonitor{
		norduserd: service,
		isSnap:    snapconf.IsUnderSnap(),
	}
}

// Start blocks the thread and starts monitoring for changes in the nordvpn group.
func (n *NordvpnGroupMonitor) Start() error {
	const etcPath = "/etc"

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating new watcher: %w", err)
	}

	if err := watcher.Add(etcPath); err != nil {
		return fmt.Errorf("adding file to watcher: %w", err)
	}

	currentGrupMembers, err := getNordvpnGroupMembers()
	if err != nil {
		return fmt.Errorf("getting initial group members: %w", err)
	} else if !n.isSnap { // in snap environment norduser is enabled on the cli side
		n.startForEveryGroupMember(currentGrupMembers)
	}

	defer watcher.Close()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("groupfile monitor channel closed")
			}

			// Because utilities used to modify the group do so atomically, we also need to monitor for creation of
			// the file instead of modifications.
			if (event.Has(fsnotify.Create) || event.Has(fsnotify.Write)) && event.Name == groupfilePath {
				newGroupMembers, err := getNordvpnGroupMembers()
				if err == nil {
					n.handleGroupUpdate(currentGrupMembers, newGroupMembers)
					if err != nil {
						log.Println("Failed to read new group members after groupfile has changed:", err)
					} else {
						currentGrupMembers = newGroupMembers
					}
				} else {
					log.Println("Failed to get new group members:", err)
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
