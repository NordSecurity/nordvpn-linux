package snapconf

import (
	"os"
	"strconv"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestSub(t *testing.T) {
	category.Set(t, category.Unit)
	for i, tt := range []struct {
		s1  []int
		s2  []int
		res []int
	}{
		{s1: []int{1}, s2: []int{1}, res: []int{}},
		{s1: []int{1, 2}, s2: []int{1}, res: []int{2}},
		{s1: []int{1, 2, 3, 4}, s2: []int{3, 1}, res: []int{2, 4}},
		{s1: []int{1}, s2: []int{1, 2, 3}, res: []int{}},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res, sub(tt.s1, tt.s2))
		})
	}
}

func TestContainsAll(t *testing.T) {
	category.Set(t, category.Unit)
	for i, tt := range []struct {
		s1  []int
		s2  []int
		res bool
	}{
		{s1: []int{1}, s2: []int{1}, res: true},
		{s1: []int{1, 2}, s2: []int{1}, res: true},
		{s1: []int{1, 2, 3, 4}, s2: []int{3, 1}, res: true},
		{s1: []int{}, s2: []int{}, res: true},
		{s1: []int{}, s2: []int{1}, res: false},
		{s1: []int{1, 3}, s2: []int{1, 2, 3, 4}, res: false},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res, containsAll(tt.s1, tt.s2))
		})
	}
}

func TestRealUserHomeDir(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		env      map[string]string
		expected string
	}{
		{
			name:     "Not running in snap",
			expected: "",
		},
		{
			name:     "$SNAP_REAL_HOME is set",
			env:      map[string]string{"SNAP_REAL_HOME": "/home/user"},
			expected: "/home/user",
		},
		{
			name:     "$SNAP_REAL_HOME is not set, $SNAP_USER_DATA != $HOME",
			env:      map[string]string{"SNAP_USER_DATA": "/home/user/"},
			expected: "",
		},
		{
			name: "$SNAP_REAL_HOME is not set, $SNAP_USER_DATA is equal to $HOME",
			env: map[string]string{
				"SNAP_USER_DATA": "/home/user/snap/nordvpn/1",
				"HOME":           "/home/user/snap/nordvpn/1",
			},
			expected: "/home/user",
		},
		{
			name: "$SNAP_REAL_HOME and $HOME are not set",
			env: map[string]string{
				"HOME": "",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			originalEnvValues := make(map[string]*string)
			for key, value := range test.env {
				existingValue, ok := os.LookupEnv(key)
				if ok {
					originalEnvValues[key] = &existingValue
				} else {
					originalEnvValues[key] = nil
				}

				os.Setenv(key, value)
			}

			defer func() {
				for key, value := range originalEnvValues {
					if value != nil {
						os.Setenv(key, *value)
					} else {
						os.Unsetenv(key)
					}
				}
			}()

			dir := RealUserHomeDir()
			assert.Equal(t, test.expected, dir)
		})
	}
}
