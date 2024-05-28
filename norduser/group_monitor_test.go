package norduser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	nordusermock "github.com/NordSecurity/nordvpn-linux/test/mock/norduser/service"
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

func Test_findGroupEntry(t *testing.T) {
	category.Set(t, category.Unit)

	rootGroup := "root:x:0:"
	groups := []string{
		rootGroup,
		"daemon:x:1:user",
		"bin:x:2:",
		"sys:x:3:",
		"adm:x:4:syslog,user",
	}

	groupFile := strings.Join(groups, "\n")

	tests := []struct {
		name           string
		groupName      string
		expectedResult string
	}{
		{
			name:           "group exists",
			groupName:      "root",
			expectedResult: rootGroup,
		},
		{
			name:           "group does not exist",
			groupName:      "test",
			expectedResult: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := findGroupEntry(groupFile, test.groupName)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

func Test_getGroupMembers(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		groupEntry     string
		expectedResult userSet
	}{
		{
			name:       "single user group",
			groupEntry: "nordvpn:x:996:user",
			expectedResult: userSet{
				"user": notActive,
			},
		},
		{
			name:       "multi user group",
			groupEntry: "nordvpn:x:996:user1,user2,user3",
			expectedResult: userSet{
				"user1": notActive,
				"user2": notActive,
				"user3": notActive,
			},
		},
		{
			name:           "empty group",
			groupEntry:     "nordvpn:x:996:",
			expectedResult: userSet{},
		},
		{
			name:           "group name starts with nordvpn",
			groupEntry:     "nordvpn_ddd:x:996:",
			expectedResult: userSet{},
		},
		{
			name:           "group name starts with nordvpn",
			groupEntry:     "nordvpn:",
			expectedResult: userSet{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getGroupMembers(test.groupEntry)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

func Test_parseSessions(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name           string
		activeSessions string
		expectedResult []string
	}{
		{
			name:           "single row",
			activeSessions: "user1 :1       :1               09:32   ?xdm?   2:27m  0.00s /usr/libexec/gdm-x-session --run-script env GNOME_SHELL_SESSION_MODE=ubuntu /usr/bin/gnome-session --session=ubuntu",
			expectedResult: []string{"user1"},
		},
		{
			name:           "no sessions",
			activeSessions: "",
			expectedResult: []string{},
		},
		{
			name: "multiple sessions",
			activeSessions: "user1 :1       :1               09:32   ?xdm?   2:27m  0.00s /usr/libexec/gdm-x-session --run-script env GNOME_SHELL_SESSION_MODE=ubuntu /usr/bin/gnome-session --session=ubuntu" +
				"\nuser2 :1       :1               09:32   ?xdm?   2:27m  0.00s /usr/libexec/gdm-x-session --run-script env GNOME_SHELL_SESSION_MODE=ubuntu /usr/bin/gnome-session --session=ubuntu" +
				"\nuser3 :1       :1               09:32   ?xdm?   2:27m  0.00s /usr/libexec/gdm-x-session --run-script env GNOME_SHELL_SESSION_MODE=ubuntu /usr/bin/gnome-session --session=ubuntu",
			expectedResult: []string{"user1", "user2", "user3"},
		},
		{
			name:           "valid characters",
			activeSessions: "us.er-1 :1       :1               09:32   ?xdm?   2:27m  0.00s /usr/libexec/gdm-x-session --run-script env GNOME_SHELL_SESSION_MODE=ubuntu /usr/bin/gnome-session --session=ubuntu",
			expectedResult: []string{"us.er-1"},
		},
		{
			name:           "max username length",
			activeSessions: "lkwkdsF77Z2diQyD95RycbaLpYFuXEIf :1       :1               09:32   ?xdm?   2:27m  0.00s /usr/libexec/gdm-x-session --run-script env GNOME_SHELL_SESSION_MODE=ubuntu /usr/bin/gnome-session --session=ubuntu",
			expectedResult: []string{"lkwkdsF77Z2diQyD95RycbaLpYFuXEIf"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sessions := parseSessions(test.activeSessions)
			assert.Equal(t, test.expectedResult, sessions)
		})
	}
}

func Test_handleGroupUpdate(t *testing.T) {
	category.Set(t, category.Unit)

	// normally userIDs struct should also contain gid and home directory, but for purpose of this test uid is enough
	user1ID := userIDs{uid: 1000}
	user1Name := "user1"

	user2ID := userIDs{uid: 100001}
	user2Name := "user2"

	usernameToUserIDs := map[string]userIDs{
		user1Name: user1ID,
		user2Name: user2ID,
	}

	tests := []struct {
		name               string
		oldGroup           userSet
		newGroup           userSet
		getUsersErr        error
		expectedEnabled    []uint32
		expectedDisabled   []uint32
		expectedUserStates userSet
	}{
		{
			name:               "enable all",
			oldGroup:           userSet{},
			newGroup:           userSet{user1Name: userActive},
			getUsersErr:        nil,
			expectedEnabled:    []uint32{user1ID.uid},
			expectedDisabled:   []uint32{},
			expectedUserStates: userSet{user1Name: norduserRunning},
		},
		{
			name:               "disable all",
			oldGroup:           userSet{user1Name: norduserRunning},
			newGroup:           userSet{},
			getUsersErr:        nil,
			expectedEnabled:    []uint32{},
			expectedDisabled:   []uint32{user1ID.uid},
			expectedUserStates: userSet{},
		},
		{
			name:               "enalbe and disable",
			oldGroup:           userSet{user1Name: norduserRunning},
			newGroup:           userSet{user2Name: userActive},
			getUsersErr:        nil,
			expectedEnabled:    []uint32{user2ID.uid},
			expectedDisabled:   []uint32{user1ID.uid},
			expectedUserStates: userSet{user2Name: norduserRunning},
		},
		{
			name:               "no users",
			oldGroup:           userSet{},
			newGroup:           userSet{},
			getUsersErr:        nil,
			expectedEnabled:    []uint32{},
			expectedDisabled:   []uint32{},
			expectedUserStates: userSet{},
		},
		{
			name:               "no state change",
			oldGroup:           userSet{user1Name: norduserRunning},
			newGroup:           userSet{user1Name: userActive},
			getUsersErr:        nil,
			expectedEnabled:    []uint32{},
			expectedDisabled:   []uint32{},
			expectedUserStates: userSet{user1Name: norduserRunning},
		},
		// In this test, as function fails to obtain user IDs/take any actions, we expect the resulting state to be
		// unchanged combination of oldGroup and newGroup, i.e norduser is still running for user1 and user2 is active
		// but norduser is not running for them.
		{
			name:               "error when getting users",
			oldGroup:           userSet{user1Name: norduserRunning},
			newGroup:           userSet{user2Name: userActive},
			getUsersErr:        fmt.Errorf("failed to get user"),
			expectedEnabled:    []uint32{},
			expectedDisabled:   []uint32{},
			expectedUserStates: userSet{user1Name: norduserRunning, user2Name: userActive},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockService := nordusermock.NewMockNorduserCombinedService()

			mockUserIDGetter := newUserIDGetterMock(usernameToUserIDs)
			mockUserIDGetter.GetErr = test.getUsersErr

			groupMonitor := NordVPNGroupMonitor{
				norduserd:    &mockService,
				userIDGetter: mockUserIDGetter,
			}

			userStates := groupMonitor.handleGroupUpdate(test.oldGroup, test.newGroup)
			assert.Equal(t, test.expectedEnabled, mockService.Enabled,
				"norduserd was not enabled for the expected users")
			assert.Equal(t, test.expectedUserStates, userStates, "State was not properly updated for some users.")
		})
	}
}
