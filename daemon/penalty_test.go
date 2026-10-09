package daemon

import (
	"math"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

const (
	DistanceDelta = 0.00001
	LoadDelta     = 0.001
	PenaltyDelta  = 0.001
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

func TestPenalty(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		d, dmin, dmax                                        float64
		load                                                 int64
		userCountry, serverCountry                           string
		hubscore, randomComponent, expected, expectedPartial float64
	}{
		{7000, 500, 8579,
			20, "us", "uk", 0.68, 0,
			3.94854714, -0.05145286,
		},
		{500, 500, 10000,
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
			got, gotPartial := penalty(item.d, item.dmin, item.dmax,
				item.load, item.userCountry, item.serverCountry, hubScore, item.randomComponent)

			assert.LessOrEqual(t, math.Abs(item.expected-got), PenaltyDelta)
			assert.LessOrEqual(t, math.Abs(item.expectedPartial-gotPartial), PenaltyDelta)
		}
	}
}
