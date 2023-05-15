package daemon

import "math"

const (
	// R defines earth radius in meters
	R = 6371e3
)

// distance calculates distance between geographical lons
func distance(srcLatitude, srcLongitude, dstLatitude, dstLongitude float64) float64 {
	srcRad := srcLatitude * math.Pi / 180
	dstRad := dstLatitude * math.Pi / 180
	delta := (dstLongitude - srcLongitude) * math.Pi / 180
	return math.Acos(math.Sin(srcRad)*math.Sin(dstRad)+math.Cos(srcRad)*math.Cos(dstRad)*math.Cos(delta)) * R
}
