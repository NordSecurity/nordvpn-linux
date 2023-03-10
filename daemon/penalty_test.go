package daemon

import (
	"math"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const (
	DistanceDelta    = 0.00001
	ObfuscationDelta = 0.00001
	LoadDelta        = 0.001
	PenaltyDelta     = 0.001
)

func TestDistancePenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		dist, dmin, dmax, expected float64
	}{
		{1500, 300, 10000, 0.246208},
		{500, 0, 4000, 0.247487},
		{0, 0, 9625, 0},
		{7000, 500, 8579, 0.627879},
	}

	for _, item := range tests {
		got := distancePenalty(item.dist, item.dmin, item.dmax)
		assert.LessOrEqual(t, math.Abs(item.expected-got), DistanceDelta)
	}
}

func TestCountryPenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		user, server string
		expected     float64
	}{
		{"uk", "uk", 0},
		{"us", "tl", 1},
		{"lv", "vl", 1},
		{"cz", "cz", 0},
	}

	for _, item := range tests {
		got := countryPenalty(item.user, item.server)
		assert.Equal(t, item.expected, got)
	}
}

func TestLoadPenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		load     int64
		expected float64
	}{
		{100, 10000000000},
		{10, 1},
		{20, 4},
		{30, 27},
		{47, 1441.6503},
	}

	for _, item := range tests {
		got := loadPenalty(item.load)
		assert.LessOrEqual(t, math.Abs(got-item.expected), LoadDelta)
	}
}

func TestHubPenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		hubScore float64
		expected float64
	}{
		{1, 1},
		{2, 2},
		{1.56, 1.56},
		{1049.314, 1049.314},
	}

	nilTest := hubPenalty(nil)
	assert.Equal(t, nilTest, float64(0))

	for _, item := range tests {
		hubScore := &item.hubScore
		got := hubPenalty(hubScore)
		assert.Equal(t, got, item.hubScore)
	}
}

func TestObfuscationPenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		obfuscated    bool
		t, tmin, tmax int64
		expected      float64
	}{
		{false, 12357387, 1239, 1867484685, 0},
		{true, 1483279200, 1467374400, 1522781999, 0.999975},
		{true, 1522781999, 1522781999, 1522981000, 1},
		{true, 1552329600, 1545329600, 1555329600, 0.98764686},
	}

	for _, item := range tests {
		got := obfuscationPenalty(item.obfuscated, item.t, item.tmin, item.tmax)
		assert.LessOrEqual(t, math.Abs(item.expected-got), ObfuscationDelta)
	}
}

func TestPenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		obfuscated                                           bool
		d, dmin, dmax                                        float64
		t, tmin, tmax, load                                  int64
		userCountry, serverCountry                           string
		hubscore, randomComponent, expected, expectedPartial float64
	}{
		{true, 7000, 500, 8579,
			1552329600, 1545329600, 1555329600,
			20, "us", "uk", 0.68, 0,
			4.936194, 0.936194,
		},
		{false, 500, 500, 10000,
			1552329600, 1522781999, 1522981000,
			45, "tl", "tl", 0, 0,
			869.8740656, 0.000142,
		},
	}

	for _, item := range tests {
		// set appropriate hubScore values
		hubScore := &item.hubscore
		if item.hubscore == 0 {
			hubScore = nil
		}
		for i := 0; i < 500; i++ {
			// run through some different random values
			item.randomComponent = randFloat(time.Now().UnixNano(), 0, 0.001)
			got, gotPartial := penalty(item.obfuscated, item.d, item.dmin, item.dmax, item.t, item.tmin, item.tmax,
				item.load, item.userCountry, item.serverCountry, hubScore, item.randomComponent)

			assert.LessOrEqual(t, math.Abs(item.expected-got), PenaltyDelta)
			assert.LessOrEqual(t, math.Abs(item.expectedPartial-gotPartial), PenaltyDelta)
		}
	}
}
