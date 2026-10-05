package norduser

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/NordSecurity/nordvpn-linux/filewatch"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/norduser/service"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

const (
	etcPath       = "/etc"
	groupFilePath = etcPath + "/group"
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
			log.ProcessMonitor.Error("getting user IDs when enabling norduser:", err)
			return
		}

		if err := norduserSrevice.Enable(userIDs.uid, userIDs.gid, userIDs.home); err != nil {
			log.ProcessMonitor.Error("enabling norduserd for member:", err)
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
			log.ProcessMonitor.Error("getting user IDs when disabling norduser:", err)
			return
		}

		if err := norduserSrevice.Stop(userIDs.uid, false); err != nil {
			log.ProcessMonitor.Error("disabling norduserd for user:", err.Error())
			return
		}

		*s = notActive
	} else if *s == runningGUI && newState == loginText { // user logged out of the GUI process, we need
		// to restart norduserd in order to re-enable tray when user logs back in to GUI
		userIDs, err := userIDGetter.getUserID(username)
		if err != nil {
			log.ProcessMonitor.Error("getting user IDs when restarting norduser:", err)
			return
		}

		if err := norduserSrevice.Restart(userIDs.uid); err != nil {
			log.ProcessMonitor.Error("failed to restart norduserd:", err)
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
	norduserd     service.Service
	sessionGetter sessionGetter
	isSnap        bool
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

	activeUsers, err := n.sessionGetter.getActiveUsers()
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
	activeUsers, err := n.sessionGetter.getActiveUsers()
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
func (n *NorduserProcessMonitor) Start(ctx context.Context) error {
	sessionGetter, err := newSessionGetter()
	if err != nil {
		return fmt.Errorf("creating session getter: %w", err)
	}
	defer sessionGetter.close()
	n.sessionGetter = sessionGetter

	sessionDatabasePath := sessionGetter.getDatabasePath()
	if databaseLinkPath, err := filepath.EvalSymlinks(sessionDatabasePath); err != nil {
		log.ProcessMonitor.Warn(
			"failed to read session database link path, will attempt monitoring with unresolved link:", err)
	} else {
		sessionDatabasePath = databaseLinkPath
	}
	log.ProcessMonitor.Info("monitoring", sessionDatabasePath, "for session state changes")

	watcher, err := filewatch.GetFileWatcher(etcPath, sessionDatabasePath)
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}
	defer watcher.Close()

	currentGroupMembers, err := n.handleGroupFileUpdate(make(userSet))
	if err != nil {
		return fmt.Errorf("starting norduserd for the initial group members: %w", err)
	}

	// Instead of performing state update right away, a monitored file update arms one of the state update timers. When
	// the timer fires, an appropriate state update will be performed. This is done to throttle the state update in case
	// of multiple file update events arriving in short succession.
	updateCoalesceTimer := time.Millisecond * 100
	var sessionUpdate <-chan time.Time
	var groupUpdate <-chan time.Time

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
				if groupUpdate == nil && (event.Has(fsnotify.Create) || event.Has(fsnotify.Write)) {
					groupUpdate = time.After(updateCoalesceTimer)
				}
			case sessionDatabasePath:
				if sessionUpdate == nil {
					sessionUpdate = time.After(updateCoalesceTimer)
				}
			}
		case <-groupUpdate:
			groupUpdate = nil
			if newGroupMembers, err := n.handleGroupFileUpdate(currentGroupMembers); err != nil {
				log.ProcessMonitor.Error("failed to handle change of groupfile:", err)
			} else {
				currentGroupMembers = newGroupMembers
			}
		case <-sessionUpdate:
			sessionUpdate = nil
			if newGroupMembers, err := n.handleUTMPFileUpdate(currentGroupMembers); err != nil {
				log.ProcessMonitor.Error("failed to handle change of session database file:", err)
			} else {
				currentGroupMembers = newGroupMembers
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("groupfile monitor error channel closed")
			}
			log.ProcessMonitor.Error("group monitor error:", err)
		case <-ctx.Done():
			return nil
		}
	}
}
