package norduser

import (
	"fmt"
	"slices"

	"github.com/fsnotify/fsnotify"

	"github.com/NordSecurity/nordvpn-linux/filewatch"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
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
			log.Error("getting user IDs when enabling norduser:", err)
			return
		}

		if err := norduserSrevice.Enable(userIDs.uid, userIDs.gid, userIDs.home); err != nil {
			log.Error("enabling norduserd for member:", err)
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
			log.Error("getting user IDs when disabling norduser:", err)
			return
		}

		if err := norduserSrevice.Stop(userIDs.uid, false); err != nil {
			log.Error("disabling norduserd for user:", err.Error())
			return
		}

		*s = notActive
	} else if *s == runningGUI && newState == loginText { // user logged out of the GUI process, we need
		// to restart norduserd in order to re-enable tray when user logs back in to GUI
		userIDs, err := userIDGetter.getUserID(username)
		if err != nil {
			log.Error("getting user IDs when restarting norduser:", err)
			return
		}

		if err := norduserSrevice.Restart(userIDs.uid); err != nil {
			log.Error("failed to restart norduserd:", err)
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
	norduserd        service.Service
	isSnap           bool
	filesystemHandle internal.FileSystemHandle
	userIDGetter
}

func NewNorduserProcessMonitor(service service.Service) NorduserProcessMonitor {
	return NorduserProcessMonitor{
		norduserd:        service,
		isSnap:           snapconf.IsUnderSnap(),
		filesystemHandle: internal.StdFilesystemHandle{},
		userIDGetter:     osGetter{},
	}
}

// handleGroupFileUpdate updates the system group members list and performs the following actions:
//  1. user socket directory is created for users who were added to the system group
//  2. norduser is started for active users who were added to the system group
//  3. norduser is stopped for users who were removed from the system group
//  4. user socket directory is removed for users who were removed from the system group
//
// If simpleMode is true, steps 2 and 3 are skipped and only user socket directory actions will be performed. This is
// useful for environments where utmp is not available(such as docker) where it is impossible to determine if user is
// active.
func (n *NorduserProcessMonitor) handleGroupFileUpdate(currentGroupMembers userSet, simpleMode bool) (userSet, error) {
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

		err := createSocketDirectory(newGroupMemberUsername, n.userIDGetter, n.filesystemHandle)
		if err != nil {
			log.Error("failed to create users socket directory:", err)
			continue
		}

		if !simpleMode {
			currentGroupMembers[newGroupMemberUsername] = notActive
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
			if !simpleMode {
				memberState.changeState(notActive, memberUsername, n.userIDGetter, n.norduserd)
				delete(currentGroupMembers, memberUsername)
			}

			if err := removeSocketDirectory(memberUsername, n.userIDGetter, n.filesystemHandle); err != nil {
				log.Error("failed to remove user socket directory:", err)
			}
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

// Start blocks the thread and starts monitoring for changes in the nordvpn group.
func (n *NorduserProcessMonitor) Start() error {
	simpleMode := false

	watcher, err := filewatch.GetFileWatcher(etcPath, utmpFilePath)
	if err != nil {
		// Watcher creation will fail if utmp file is not available, which is possible in certain environments such
		// as docker. In such cases we start a simplified watcher that monitors only the group file, for user socket
		// directory creation purposes.
		log.Warn("failed to create group/session file watcher:", err)
		simpleMode = true

		watcher, err = filewatch.GetFileWatcher(etcPath)
		if err != nil {
			return fmt.Errorf("creating file watcher: %w", err)
		}
	}
	defer watcher.Close()

	currentGrupMembers, err := n.handleGroupFileUpdate(make(userSet), simpleMode)
	if err != nil {
		return fmt.Errorf("starting norduserd for the initial group members: %w", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("groupfile monitor channel closed")
			}

			switch event.Name {
			case groupFilePath:
				// Because utilities used to modify the group do so atomically, we also need to monitor for creation of
				// the file instead of modifications.
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					if newGroupMembers, err := n.handleGroupFileUpdate(currentGrupMembers, simpleMode); err != nil {
						log.Error("failed to handle change of groupfile:", err)
					} else {
						currentGrupMembers = newGroupMembers
					}
				}
			case utmpFilePath:
				if newGroupMembers, err := n.handleUTMPFileUpdate(currentGrupMembers); err != nil {
					log.Error("failed to handle change of utmp file:", err)
				} else {
					currentGrupMembers = newGroupMembers
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("groupfile monitor error channel closed")
			}
			log.Error("group monitor error:", err)
		}
	}
}
