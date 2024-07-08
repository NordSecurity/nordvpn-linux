package norduser

import (
	"fmt"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

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

func getGroupMembers(groupEntry string) []string {
	groupEntry = strings.TrimSuffix(groupEntry, "\n")
	sepIndex := strings.LastIndex(groupEntry, ":")
	groupMembers := groupEntry[sepIndex+1:]

	if groupMembers == "" {
		return []string{}
	}

	return strings.Split(groupMembers, ",")
}

func getNordVPNGroupMembers() ([]string, error) {
	groupEntry, err := getGroupEntry(internal.NordvpnGroup)
	if err != nil {
		return nil, fmt.Errorf("getting group entry: %w", err)
	}

	return getGroupMembers(groupEntry), nil
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
