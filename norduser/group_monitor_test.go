package norduser

import (
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

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
				"user": false,
			},
		},
		{
			name:       "multi user group",
			groupEntry: "nordvpn:x:996:user1,user2,user3",
			expectedResult: userSet{
				"user1": false,
				"user2": false,
				"user3": false,
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
