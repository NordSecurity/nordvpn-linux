package snapconf

import (
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
