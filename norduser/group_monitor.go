package norduser

import (
	"fmt"
	"log"
	"os/user"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

const (
	groupFilePath = "/etc/group"
	utmpFilePath  = "/var/run/utmp"
)

type userState int

const (
	notActive = iota
	userActive
	norduserRunning
)

type userSet map[string]userState

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
	file, err := internal.FileRead(groupFilePath)
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
		members[member] = notActive
	}

	return members
}

func getNordVPNGroupMembers() (userSet, error) {
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

type userIDGetter interface {
	getUserID(username string) (userIDs, error)
}

type osGetter struct{}

func (osGetter) getUserID(username string) (userIDs, error) {
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

func parseSessions(sessions string) []string {
	re := regexp.MustCompile(`(?m)^(\S+)`)
	matches := re.FindAllStringSubmatch(sessions, -1)
	if matches == nil {
		return []string{}
	}

	activeUsers := []string{}
	for _, match := range matches {
		activeUsers = append(activeUsers, match[1])
	}

	return activeUsers
}

func updateGroupMembersState(groupMembers userSet, activeUsers []string) userSet {
	for member := range groupMembers {
		if idx := slices.Index(activeUsers, member); idx != -1 {
			groupMembers[member] = userActive
		}
	}

	return groupMembers
}

// NordVPNGroupMonitor monitors the nordvpn system group and starts/stops norduserd for users added/removed from the
// group.
type NordVPNGroupMonitor struct {
	norduserd service.NorduserService
	isSnap    bool
	userIDGetter
}

func NewNordVPNGroupMonitor(service service.NorduserService) NordVPNGroupMonitor {
	return NordVPNGroupMonitor{
		norduserd:    service,
		isSnap:       snapconf.IsUnderSnap(),
		userIDGetter: osGetter{},
	}
}

func (n *NordVPNGroupMonitor) handleGroupUpdate(currentGroupMembers userSet, newGroupMembers userSet) userSet {
	for member, state := range currentGroupMembers {
		// member is still in group and session is active, do nothing
		if newState, ok := newGroupMembers[member]; ok && newState == userActive {
			if state == norduserRunning {
				newGroupMembers[member] = norduserRunning
			}
			continue
		}

		if state != norduserRunning {
			continue
		}

		userIDs, err := n.getUserID(member)
		if err != nil {
			// store the old state, since norduserd is still running for this member
			newGroupMembers[member] = state
			log.Println(internal.ErrorPrefix, "failed to look up UID/GID of deleted group member: ", err)
			continue
		}

		if err := n.norduserd.Stop(userIDs.uid, false); err != nil {
			newGroupMembers[member] = state
			log.Println(internal.ErrorPrefix, "disabling norduserd for user:", err.Error())
		}
	}

	if n.isSnap {
		// in snap environment norduser is enabled on the cli side
		return newGroupMembers
	}

	// enable norduserd for new members
	for member, status := range newGroupMembers {
		// we only want to start norduser when user is active but norduser is not active
		if status != userActive {
			continue
		}

		userID, err := n.getUserID(member)
		if err != nil {
			log.Println("failed to lookup UID/GID for new group member:", err)
			continue
		}

		if err := n.norduserd.Enable(userID.uid, userID.gid, userID.home); err != nil {
			log.Println(internal.ErrorPrefix, "enabling norduserd for member:", err)
			continue
		}

		newGroupMembers[member] = norduserRunning
	}

	return newGroupMembers
}

func (n *NordVPNGroupMonitor) handleGropuFileUpdate(currentGrupMembers userSet) (userSet, error) {
	newGroupMembers, err := getNordVPNGroupMembers()
	if err != nil {
		return nil, fmt.Errorf("failed to get new group members: %w", err)
	}

	activeUsers, err := getActiveUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get active users after group file update: %w", err)
	}
	newGroupMembers = updateGroupMembersState(newGroupMembers, activeUsers)

	return n.handleGroupUpdate(currentGrupMembers, newGroupMembers), nil
}

func (n *NordVPNGroupMonitor) startForEveryGroupMember(groupMembers userSet) {
	for member, status := range groupMembers {
		if status == notActive {
			continue
		}

		user, err := n.getUserID(member)
		if err != nil {
			log.Println(internal.ErrorPrefix, "failed to get UID/GID for group member:", err)
			continue
		}

		if err := n.norduserd.Enable(user.uid, user.gid, user.home); err != nil {
			log.Println(internal.ErrorPrefix, "failed to start norduser for group member:", err)
			continue
		}

		groupMembers[member] = norduserRunning
	}
}

// Start blocks the thread and starts monitoring for changes in the nordvpn group.
func (n *NordVPNGroupMonitor) Start() error {
	const etcPath = "/etc"

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating new watcher: %w", err)
	}

	if err := watcher.Add(etcPath); err != nil {
		return fmt.Errorf("adding group file to watcher: %w", err)
	}

	if err := watcher.Add(utmpFilePath); err != nil {
		return fmt.Errorf("adding utmp file to watcher: %w", err)
	}

	currentGrupMembers, err := getNordVPNGroupMembers()
	if err != nil {
		return fmt.Errorf("getting initial group members: %w", err)
	}

	activeUsers, err := getActiveUsers()
	if err != nil {
		return fmt.Errorf("getting initial active users: %w", err)
	}

	currentGrupMembers = updateGroupMembersState(currentGrupMembers, activeUsers)
	if !n.isSnap { // in snap environment norduser is enabled on the cli side
		n.startForEveryGroupMember(currentGrupMembers)
	}

	defer watcher.Close()
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("groupfile monitor channel closed")
			}

			if event.Name == groupFilePath {
				// Because utilities used to modify the group do so atomically, we also need to monitor for creation of
				// the file instead of modifications.
				if (event.Has(fsnotify.Create) || event.Has(fsnotify.Write)) && event.Name == groupFilePath {
					if newGroupMembers, err := n.handleGropuFileUpdate(currentGrupMembers); err != nil {
						log.Println(internal.ErrorPrefix, "failed to handle change of groupfile: ", err)
					} else {
						currentGrupMembers = newGroupMembers
					}
				}
			} else if event.Name == utmpFilePath {
				activeUsers, err := getActiveUsers()
				if err == nil {
					newGroupMembers := updateGroupMembersState(currentGrupMembers, activeUsers)
					currentGrupMembers = n.handleGroupUpdate(currentGrupMembers, newGroupMembers)
				} else {
					log.Println(internal.ErrorPrefix, "failed to get active users after utmp file update: ", err)
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
