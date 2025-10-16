package norduser

import (
	"fmt"
	"log"
	"slices"

	"github.com/fsnotify/fsnotify"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

const (
	etcPath       = "/etc"
	groupFilePath = etcPath + "/group"
	utmpFilePath  = "/var/run/utmp"
)

type norduserState int

const (
	notActive norduserState = iota
	loginGUI
	loginText
	runningGUI
	runningText
)

// changeState will execute actions appropriate for newState for the given username and then update the state to
// appropriate new state based on result of those actions
// Desired state transitions:
//   - notActive 	=> 	loginGUI - start application, update state to runningGUI
//   - notActive 	=> 	loginText - start application, update state to runningText
//   - runningGUI 	=> 	notActive - stop application, update state to notActive
//   - runningText 	=> 	notActive - stop application, update state to notActive
//   - runningGUI 	=> 	loginText - restart application, update state to runningText
//   - runningText 	=> 	loginGUI - update state to runningGUI
//
// Other state transitions should result in a noop.
//
// More on runningGUI to loginText transition:
// Due to library limitations, when user doesn't have any GUI sessions, tray will be disabled. In order to enable it on
// subsequent GUI logins, we need to restart the application.
//
// Such actions are not necessary in case of transitioning from runningText to loginGUI, since in this case tray was
// not started.
func (s *norduserState) changeState(newState norduserState,
	username string,
	userIDGetter userIDGetter,
	norduserSrevice service.Service) {
	if *s == notActive &&
		(newState == loginGUI || newState == loginText) { // user logged in, start norduserd
		userIDs, err := userIDGetter.getUserID(username)
		if err != nil {
			log.Println(internal.ErrorPrefix, "getting user IDs when enabling norduser:", err)
			return
		}

		if err := norduserSrevice.Enable(userIDs.uid, userIDs.gid, userIDs.home); err != nil {
			log.Println(internal.ErrorPrefix, "enabling norduserd for member:", err)
			return
		}

		if newState == loginGUI {
			*s = runningGUI
		} else {
			*s = runningText
		}
	} else if (*s == runningText || *s == runningGUI) &&
		newState == notActive { // user logged out when norduser was running, stop norduserd
		userIDs, err := userIDGetter.getUserID(username)
		if err != nil {
			log.Println(internal.ErrorPrefix, "getting user IDs when disabling norduser:", err)
			return
		}

		if err := norduserSrevice.Stop(userIDs.uid, false); err != nil {
			log.Println(internal.ErrorPrefix, "disabling norduserd for user:", err.Error())
			return
		}

		*s = notActive
	} else if *s == runningGUI && newState == loginText { // user logged out of the GUI process, we need
		// to restart norduserd in order to re-enable tray when user logs back in to GUI
		userIDs, err := userIDGetter.getUserID(username)
		if err != nil {
			log.Println(internal.ErrorPrefix, "getting user IDs when restarting norduser:", err)
			return
		}

		if err := norduserSrevice.Restart(userIDs.uid); err != nil {
			log.Println(internal.ErrorPrefix, "failed to restart norduserd:", err)
			return
		}

		*s = runningText
	} else if *s == runningText && newState == loginGUI { // when user is initially logged in via text
		// interface, we only need to update the state so that subsequent switch from GUI to text can be handled
		// correctly
		*s = runningGUI
	}
}

type userSet map[string]norduserState

// NorduserProcessMonitor monitors the nordvpn system group and starts/stops norduserd for users added/removed from the
// group.
type NorduserProcessMonitor struct {
	norduserd service.Service
	isSnap    bool
	userIDGetter
}

func NewNorduserProcessMonitor(service service.Service) NorduserProcessMonitor {
	return NorduserProcessMonitor{
		norduserd:    service,
		isSnap:       snapconf.IsUnderSnap(),
		userIDGetter: osGetter{},
	}
}

func (n *NorduserProcessMonitor) handleGroupFileUpdate(currentGroupMembers userSet) (userSet, error) {
	newGroupMembers, err := getNordVPNGroupMembers()
	if err != nil {
		return currentGroupMembers, fmt.Errorf("getting nordvpn group members: %w", err)
	}

	activeUsers, err := getActiveUsers()
	if err != nil {
		return currentGroupMembers, fmt.Errorf("getting active users after group file update: %w", err)
	}

	// initialize new group members
	for _, newGroupMemberUsername := range newGroupMembers {
		_, ok := currentGroupMembers[newGroupMemberUsername]
		if ok {
			continue
		}

		state := notActive
		userStatus, ok := activeUsers[newGroupMemberUsername]
		if ok {
			state.changeState(userStatus, newGroupMemberUsername, n.userIDGetter, n.norduserd)
		}
		currentGroupMembers[newGroupMemberUsername] = state
	}

	// update state for removed group members
	for memberUsername, memberState := range currentGroupMembers {
		if contains := slices.Contains(newGroupMembers, memberUsername); !contains {
			memberState.changeState(notActive, memberUsername, n.userIDGetter, n.norduserd)
			delete(currentGroupMembers, memberUsername)
		}
	}

	return currentGroupMembers, nil
}

func (n *NorduserProcessMonitor) handleUTMPFileUpdate(currentGroupMembers userSet) (userSet, error) {
	activeUsers, err := getActiveUsers()
	if err != nil {
		return currentGroupMembers, fmt.Errorf("getting active users after utmp file update: %w", err)
	}

	for username, state := range currentGroupMembers {
		userState, ok := activeUsers[username]
		if ok {
			state.changeState(userState, username, n.userIDGetter, n.norduserd)
		} else {
			state.changeState(notActive, username, n.userIDGetter, n.norduserd)
		}

		currentGroupMembers[username] = state
	}

	return currentGroupMembers, nil
}

func getWatcher(pathsToMonitor ...string) (watcher *fsnotify.Watcher, err error) {
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating new watcher: %w", err)
	}

	defer func() {
		if err != nil && watcher != nil {
			watcher.Close()
		}
	}()

	for _, file := range pathsToMonitor {
		if err := watcher.Add(file); err != nil {
			return nil, fmt.Errorf("adding group file to watcher: %w", err)
		}
	}

	return watcher, nil
}

// Start blocks the thread and starts monitoring for changes in the nordvpn group.
func (n *NorduserProcessMonitor) Start() error {
	watcher, err := getWatcher(etcPath, utmpFilePath)
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer watcher.Close()

	currentGrupMembers, err := n.handleGroupFileUpdate(make(userSet))
	if err != nil {
		return fmt.Errorf("starting norduserd for the initial group members: %w", err)
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
					if newGroupMembers, err := n.handleGroupFileUpdate(currentGrupMembers); err != nil {
						log.Println(internal.ErrorPrefix, "failed to handle change of groupfile:", err)
					} else {
						currentGrupMembers = newGroupMembers
					}
				}
			} else if event.Name == utmpFilePath {
				if newGroupMembers, err := n.handleUTMPFileUpdate(currentGrupMembers); err != nil {
					log.Println(internal.ErrorPrefix, "failed to handle change of utmp file:", err)
				} else {
					currentGrupMembers = newGroupMembers
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
