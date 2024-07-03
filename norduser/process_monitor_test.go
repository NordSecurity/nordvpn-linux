package norduser

import (
	"fmt"
	"slices"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	testnorduser "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
	"github.com/stretchr/testify/assert"
)

type userIDGetterMock struct {
	UsernameToIDs map[string]userIDs
	GetErr        error
}

func newUserIDGetterMock(usernameToUserIDs map[string]userIDs) *userIDGetterMock {
	return &userIDGetterMock{
		UsernameToIDs: usernameToUserIDs,
	}
}

func (u *userIDGetterMock) getUserID(username string) (userIDs, error) {
	if u.GetErr != nil {
		return userIDs{}, u.GetErr
	}

	userID, ok := u.UsernameToIDs[username]
	if !ok {
		return userIDs{}, fmt.Errorf("user ids not found")
	}

	return userID, nil
}

func Test_changeState(t *testing.T) {
	category.Set(t, category.Unit)

	// Desired state transitions:
	//   - notActive 	=> 	loginGUI - start application, update state to runningGUI
	//   - notActive 	=> 	loginText - start application, update state to runningText
	//   - runningGUI 	=> 	notActive - stop application, update state to notActive
	//   - runningText 	=> 	notActive - stop application, update state to notActive
	//   - runningGUI 	=> 	loginText - restart application, update state to runningText
	//   - runningText 	=> 	loginGUI - update state to runningGUI
	// Other transitions should result in an noop

	getUserIDErr := fmt.Errorf("get user id error")
	enableErr := fmt.Errorf("enalbe error")
	stopErr := fmt.Errorf("stop error")
	restartErr := fmt.Errorf("restart error")

	const username string = "user"
	const uid uint32 = 100001

	userIDsMap := map[string]userIDs{username: {
		uid: uid,
	}}

	tests := []struct {
		name           string
		initialState   norduserState
		newState       norduserState
		expectedState  norduserState
		expectedAction testnorduser.Action
		noAction       bool
		enableErr      error
		stopErr        error
		restartErr     error
		getUserIDErr   error
	}{
		// notActive to loginGUI
		{
			name:           "notActive to loginGUI",
			initialState:   notActive,
			newState:       loginGUI,
			expectedState:  runningGUI,
			expectedAction: testnorduser.Enable,
		},
		{
			name:          "notActive to loginGUI get user ID error",
			initialState:  notActive,
			newState:      loginGUI,
			expectedState: notActive,
			noAction:      true,
			getUserIDErr:  getUserIDErr,
		},
		{
			name:          "notActive to loginGUI enalbe error",
			initialState:  notActive,
			newState:      loginGUI,
			expectedState: notActive,
			noAction:      true,
			enableErr:     enableErr,
		},
		// notActive to loginText
		{
			name:           "notActive to loginText",
			initialState:   notActive,
			newState:       loginText,
			expectedState:  runningText,
			expectedAction: testnorduser.Enable,
		},
		{
			name:          "notActive to loginText get user ID error",
			initialState:  notActive,
			newState:      loginText,
			expectedState: notActive,
			noAction:      true,
			getUserIDErr:  getUserIDErr,
		},
		{
			name:          "notActive to loginGUI enalbe error",
			initialState:  notActive,
			newState:      loginText,
			expectedState: notActive,
			noAction:      true,
			enableErr:     enableErr,
		},
		// runningGUI to notActive
		{
			name:           "runningGUI to notActive",
			initialState:   runningGUI,
			newState:       notActive,
			expectedState:  notActive,
			expectedAction: testnorduser.Stop,
		},
		{
			name:          "runningGUI to notActive get user ID error",
			initialState:  runningGUI,
			newState:      notActive,
			expectedState: runningGUI,
			noAction:      true,
			getUserIDErr:  getUserIDErr,
		},
		{
			name:          "runningGUI to notActive get stop error",
			initialState:  runningGUI,
			newState:      notActive,
			expectedState: runningGUI,
			noAction:      true,
			stopErr:       stopErr,
		},
		// runningText to notActive
		{
			name:           "runningText to notActive",
			initialState:   runningText,
			newState:       notActive,
			expectedState:  notActive,
			expectedAction: testnorduser.Stop,
		},
		{
			name:          "runningGUI to notActive get user ID error",
			initialState:  runningText,
			newState:      notActive,
			expectedState: runningText,
			noAction:      true,
			getUserIDErr:  getUserIDErr,
		},
		{
			name:          "runningGUI to notActive get stop error",
			initialState:  runningText,
			newState:      notActive,
			expectedState: runningText,
			noAction:      true,
			stopErr:       stopErr,
		},
		// runningGUI to loginText
		{
			name:           "runningGUI to loginText",
			initialState:   runningGUI,
			newState:       loginText,
			expectedState:  runningText,
			expectedAction: testnorduser.Restart,
		},
		{
			name:          "runningGUI to loginText get user ID error",
			initialState:  runningGUI,
			newState:      loginText,
			expectedState: runningGUI,
			noAction:      true,
			getUserIDErr:  getUserIDErr,
		},
		{
			name:          "runningGUI to loginText restart error",
			initialState:  runningGUI,
			newState:      loginText,
			expectedState: runningGUI,
			noAction:      true,
			restartErr:    restartErr,
		},
		// runningText to loginGUI
		{
			name:          "runningText to loginGUI",
			initialState:  runningText,
			newState:      loginGUI,
			expectedState: runningGUI,
			noAction:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userIDGetterMock := newUserIDGetterMock(userIDsMap)
			userIDGetterMock.GetErr = test.getUserIDErr

			norduserProcessManagerMock := testnorduser.NewMockNorduserCombinedService()
			norduserProcessManagerMock.EnableErr = test.enableErr
			norduserProcessManagerMock.StopErr = test.stopErr
			norduserProcessManagerMock.RestartErr = test.restartErr

			test.initialState.changeState(test.newState, username, userIDGetterMock, &norduserProcessManagerMock)

			assert.Equal(t, test.expectedState, test.initialState,
				"State was not properly updated after handling the state transition.")

			if test.noAction {
				assert.True(t, norduserProcessManagerMock.CheckNoAction(),
					"Unexpected action was taken when changing state when no action was expected.")
				return
			}

			actions := norduserProcessManagerMock.ActionToUIDs[test.expectedAction]
			assert.Len(t, actions, 1, "More than one action taken")

			// check that no other type of action was executed
			assert.True(t, norduserProcessManagerMock.CheckNoAction(test.expectedAction),
				"Unexpected action was taken when changing state.")
		})
	}
}

func Test_changeState_noop(t *testing.T) {
	type transition struct {
		initial norduserState
		new     norduserState
	}
	nonNoopTransitions := []transition{
		{initial: notActive, new: loginGUI},
		{initial: notActive, new: loginText},
		{initial: runningGUI, new: notActive},
		{initial: runningText, new: notActive},
		{initial: runningGUI, new: loginText},
		{initial: runningText, new: loginGUI},
	}

	states := []norduserState{
		notActive,
		loginGUI,
		loginText,
		runningGUI,
		runningText,
	}

	const username string = "user"
	const uid uint32 = 100001

	userIDsMap := map[string]userIDs{username: {
		uid: uid,
	}}

	for _, initialState := range states {
		for _, newState := range states {
			isNonNoop := slices.ContainsFunc(nonNoopTransitions, func(t transition) bool {
				if t.initial == initialState && t.new == newState {
					return true
				}
				return false
			})

			if isNonNoop {
				continue
			}

			userIDGetterMock := newUserIDGetterMock(userIDsMap)
			norduserProcessManagerMock := testnorduser.NewMockNorduserCombinedService()

			// copy initial state so we can verify that state did not change after the transition
			expectedState := initialState

			initialState.changeState(newState, username, userIDGetterMock, &norduserProcessManagerMock)

			assert.Equal(t, expectedState, initialState,
				"Unexpected state change after noop state transition.")
			assert.True(t, norduserProcessManagerMock.CheckNoAction(),
				"Unexpected actions executed after noop state transition.")
		}
	}
}

func Test_stopForDeletedGroupMembers(t *testing.T) {
	category.Set(t, category.Unit)

	const username1 string = "user1"
	const uid1 uint32 = 1001

	const username2 string = "user2"
	const uid2 uint32 = 100001

	const username3 string = "user3"
	const uid3 uint32 = 53200

	userIDsMap := map[string]userIDs{
		username1: {uid: uid1},
		username2: {uid: uid2},
		username3: {uid: uid3},
	}

	tests := []struct {
		name                string
		initialGroupMembers []string
		newGroupMembers     []string
		expectedStoppedUIDs []uint32
		getUIDErr           error
		stopErr             error
		isNoop              bool
	}{
		{
			name:                "all group members removed",
			initialGroupMembers: []string{username1, username2, username3},
			newGroupMembers:     []string{},
			expectedStoppedUIDs: []uint32{uid1, uid2, uid3},
		},
		{
			name:                "stop for single user",
			initialGroupMembers: []string{username1, username2, username3},
			newGroupMembers:     []string{username2, username3},
			expectedStoppedUIDs: []uint32{uid1},
		},
		{
			name:                "initial startup",
			initialGroupMembers: []string{},
			newGroupMembers:     []string{username1, username2, username3},
			isNoop:              true,
		},
		{
			name:                "getUID error",
			initialGroupMembers: []string{username1, username2, username3},
			newGroupMembers:     []string{username2},
			getUIDErr:           fmt.Errorf("getUID error"),
			isNoop:              true,
		},
		{
			name:                "Stop error",
			initialGroupMembers: []string{username1, username2, username3},
			newGroupMembers:     []string{username2},
			stopErr:             fmt.Errorf("Stop error"),
			isNoop:              true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			userIDGetterMock := newUserIDGetterMock(userIDsMap)
			userIDGetterMock.GetErr = test.getUIDErr

			norduserProcessManagerMock := testnorduser.NewMockNorduserCombinedService()
			norduserProcessManagerMock.StopErr = test.stopErr

			norduserProcessMonitor := NorduserProcessMonitor{
				norduserd:    &norduserProcessManagerMock,
				userIDGetter: userIDGetterMock,
			}

			groupMembersAfterUpdate :=
				norduserProcessMonitor.stopForDeletedGroupMembers(test.initialGroupMembers, test.newGroupMembers)
			if test.isNoop {
				assert.True(t, norduserProcessManagerMock.CheckNoAction(),
					`Unexpected actions taken when stopping norduser
					for deleted group members(no action should be taken).`)
				assert.Equal(t, groupMembersAfterUpdate, test.initialGroupMembers,
					"Group update has failed, but new group state has replaced the old group state.")
				return
			}

			stoppedUIDs := norduserProcessManagerMock.ActionToUIDs[testnorduser.Stop]
			assert.Len(t, stoppedUIDs, len(test.expectedStoppedUIDs),
				`Invalid number of stop operations performed when stopping norduser for users removed from group,
				should be equal to number of removed users.`)

			for _, expectedStoppedUID := range test.expectedStoppedUIDs {
				assert.Contains(t, stoppedUIDs, expectedStoppedUID,
					"norduser was not stopped for UID %d.", expectedStoppedUID)
			}

			assert.True(t, norduserProcessManagerMock.CheckNoAction(testnorduser.Stop),
				`Unexpected actions(other than stopping norduser) taken
				when stopping norduser for removed group members.`)
			assert.Equal(t, test.newGroupMembers, groupMembersAfterUpdate, "Group members were not properly updated.")
		})
	}
}
