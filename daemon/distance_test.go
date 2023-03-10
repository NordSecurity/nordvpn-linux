package daemon

import (
	"math"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type TestDistanceData struct {
	srcLat, srcLong, dstLat, dstLong, distance float64
}

const DELTA = 1.0 // 1 meter error
func TestDistance(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []TestDistanceData{
		{40.7648, -73.9808, 9, -79.5, 3573834.87},
		{-49.84199, 28.12457, -7.46627, 6.89593, 5108827.84},
		{12.34567, -12.545454, 12.34567, -12.545454, 0},
		{23.57786, 51.08247, 64.68667, -82.42718, 9421907.52},
		{37.58333, 127, 37.28417, 127.01917, 33308.11},
	}
	for _, d := range tests {
		dist := distance(d.srcLat, d.srcLong, d.dstLat, d.dstLong)
		assert.True(t, ApproxEquals(dist, d.distance, DELTA))
	}
}
func ApproxEquals(a, b, delta float64) bool {
	return math.Abs(a-b) < delta
}
