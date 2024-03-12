package norduser

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/fsnotify/fsnotify"
)

const (
	groupfilePath = "/etc/group"
)

type userSet map[string]bool

func (n *NordvpnGroupMonitor) getNordvpnGroupMembers() (userSet, error) {
	const nordvpnGroupName = "nordvpn"

	output, err := exec.Command("getent", "group", nordvpnGroupName).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("running getent: %w", err)
	}

	groupEntry := string(output)
	groupEntry = strings.TrimSuffix(groupEntry, "\n")
	sepIndex := strings.LastIndex(groupEntry, ":")
	groupMembers := groupEntry[sepIndex+1:]

	members := make(userSet)
	for _, member := range strings.Split(groupMembers, ",") {
		members[member] = false
	}

	return members, nil
}

type userIDs struct {
	uid uint32
	gid uint32
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
		uid: uint32(uid),
		gid: uint32(gid),
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
			log.Println("failed to look up UID/GID of deleted group member: ", err.Error())
		}

		if err := n.norduserd.Disable(userIDs.uid); err != nil {
			log.Println("disabling norduserd for user: ", err.Error())
		}
	}

	// enable norduserd for new members
	for member, norduserdEnabled := range newGroupMembers {
		if norduserdEnabled {
			continue
		}

		userID, err := getUID(member)
		if err != nil {
			log.Println("failed to lookup UID/GID for new group member: ", err)
		}

		if err := n.norduserd.Enable(userID.uid, userID.gid); err != nil {
			log.Println("enabling norduserd for member: ", err)
		}
	}
}

// NordvpnGroupMonitor monitors the nordvpn system group and starts/stops norduserd for users added/removed from the
// group.
type NordvpnGroupMonitor struct {
	norduserd *service.Combined
}

func NewNordvpnGroupMonitor(service *service.Combined) NordvpnGroupMonitor {
	return NordvpnGroupMonitor{
		norduserd: service,
	}
}

// Start enabled norduserd for all the users in nordvpn group and starts a group monitor goroutine that will start/stop
// norduserd when user is added/removed from the group.
func (n *NordvpnGroupMonitor) Start() error {
	const etcPath = "/etc"

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating new watcher: %w", err)
	}

	if err := watcher.Add(etcPath); err != nil {
		return fmt.Errorf("adding file to watcher: %w", err)
	}

	currentGrupMembers, err := n.getNordvpnGroupMembers()
	if err != nil {
		return fmt.Errorf("getting initial group members: %w", err)
	}

	for member := range currentGrupMembers {
		usr, err := getUID(member)
		if err != nil {
			log.Println("failed to get user ids when starting nordused: ", err.Error())
		}
		if err := n.norduserd.Enable(usr.uid, usr.gid); err != nil {
			log.Println("failed to start nordused: ", err)
		}
		currentGrupMembers[member] = true
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("monitor groupfile events failed")
					return
				}

				// Because utilities used to modify the group do so atomically, we also need to monitor for creation of
				// the file instead of modifications.
				if (event.Has(fsnotify.Create) || event.Has(fsnotify.Write)) && event.Name == groupfilePath {
					newGroupMembers, err := n.getNordvpnGroupMembers()
					n.handleGroupUpdate(currentGrupMembers, newGroupMembers)
					currentGrupMembers = newGroupMembers

					if err != nil {
						log.Println("Failed to read new group members after groupfile has changed: ", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("monitor groupfile errors failed")
					return
				}
				log.Println("group monitor error:", err)
			}
		}
	}()

	return nil
}
