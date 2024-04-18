package service

import (
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_parseNorduserPIDs(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		pids         []string
		expectedPIDs []int
	}{
		{
			name:         "empty list",
			pids:         []string{},
			expectedPIDs: []int{},
		},
		{
			name:         "non empty list",
			pids:         []string{" 35139", " 35153", " 35144"},
			expectedPIDs: []int{35139, 35153, 35144},
		},
		{
			name:         "list contains malformed entries",
			pids:         []string{" 35139", " aaaa", " 35144"},
			expectedPIDs: []int{35139, 35144},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pids := strings.Join(test.pids, "\n")
			result := parseNorduserPIDs(pids)
			assert.Equal(t, test.expectedPIDs, result)
		})
	}
}

func Test_findPIDOfUID(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name        string
		uidToPID    []string
		expectedPID int
	}{
		{
			name:        "empty list",
			uidToPID:    []string{},
			expectedPID: -1,
		},
		{
			name:        "uid not present",
			uidToPID:    []string{" 1004 35139", " 1003 35153", " 1002 35144"},
			expectedPID: -1,
		},
		{
			name:        "invalid pid",
			uidToPID:    []string{" 1001 aaaa", " 1003 35153", " 1002 35144"},
			expectedPID: -1,
		},
		{
			name:        "pid found",
			uidToPID:    []string{" 1001 35255", " 1003 35153", " 1002 35144"},
			expectedPID: 35255,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pids := strings.Join(test.uidToPID, "\n")
			result := findPIDOfUID(pids, 1001)
			assert.Equal(t, test.expectedPID, result)
		})
	}
}
