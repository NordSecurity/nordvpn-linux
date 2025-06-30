package remote

import (
	"fmt"
	"math"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGroupDistribution verifies that the distribution of groups created by NewGroup is uniform.
// The test generates an arbitrary number of 10kk random UUIDs and assigns them to groups using NewGroup,
// then checks that the distribution of groups across five equal-sized brackets is statistically
// balanced. It validates:
//  1. Each bracket (1-20, 21-40, 41-60, 61-80, 81-100) contains approximately 20% of the groups
//  2. The deviation from the expected distribution is within a 2% tolerance
//  3. The standard deviation is within acceptable bounds (±2σ)
//  4. The total count of groups matches the number of iterations
func TestGroupDistribution(t *testing.T) {
	category.Set(t, category.Unit)
	const (
		iterations  = 10_000_000
		bracketSize = 20
		maxGroup    = 100
		tolerance   = 0.02 // 2% tolerance
	)

	uuid.EnableRandPool()
	defer uuid.DisableRandPool()

	groups := make([]int, iterations)
	for i := range iterations {
		id := uuid.New()
		g, err := NewGroup(id, maxGroup)
		require.NoError(t, err)
		groups[i] = g.value
	}

	// calculate specific brackets counts
	brackets := make([]int, 5)
	for _, g := range groups {
		bracketIndex := (g - 1) / bracketSize
		brackets[bracketIndex]++
	}

	// some expected statistics per bracket
	expectedCount := iterations / 5
	expectedPercentage := 20.0 // 100% / 5 brackets

	diffs := make([]float64, len(brackets))
	sum := 0
	for i, count := range brackets {
		sum += count
		diffs[i] = float64(count - expectedCount)
	}

	sumSquares := 0.0
	for _, diff := range diffs {
		sumSquares += diff * diff
	}
	stdDev := math.Sqrt(sumSquares / float64(len(brackets)))

	bracketNames := []string{"1-20", "21-40", "41-60", "61-80", "81-100"}

	for i, count := range brackets {
		percentage := float64(count) / float64(iterations) * 100
		deviations := diffs[i] / stdDev

		t.Logf("Bracket %-8s: count=%d (%.2f%%) diff=%+.0f stddev=%+.2f σ",
			bracketNames[i], count, percentage, diffs[i], deviations)

		assert.InDelta(t, expectedPercentage, percentage, tolerance*100,
			"Bracket %s percentage outside tolerance", bracketNames[i])

		assert.InDelta(t, 0, math.Abs(deviations), 2.0,
			"Bracket %s deviation outside tolerance", bracketNames[i])
	}

	// Assert total count matches iterations
	assert.Equal(t, iterations, sum,
		"Sum of bracket counts (%d) doesn't match total number of generated UUIDs (%d)", sum, iterations)
}

func TestNewGroupErrors(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		max     int
		wantErr string
	}{
		{
			name:    "zero max value",
			max:     0,
			wantErr: "max value must be positive, got 0",
		},
		{
			name:    "negative max value",
			max:     -1,
			wantErr: "max value must be positive, got -1",
		},
		{
			name:    "exceeding max allowed value",
			max:     DefaultMaxGroup + 1,
			wantErr: fmt.Sprintf("max value must not exceed %d, got %d", DefaultMaxGroup, DefaultMaxGroup+1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			group, err := NewGroup(id, tt.max)

			assert.Nil(t, group)
			assert.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestNewGroupValue(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		max     int
		wantMin int
		wantMax int
	}{
		{
			name:    "max value 20",
			max:     20,
			wantMin: 1,
			wantMax: 20,
		},
		{
			name:    "max value 100",
			max:     100,
			wantMin: 1,
			wantMax: 100,
		},
		{
			name:    "max value 50",
			max:     50,
			wantMin: 1,
			wantMax: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := uuid.New()
			group, err := NewGroup(id, tt.max)

			require.NoError(t, err)
			require.NotNil(t, group)
			assert.GreaterOrEqual(t, group.value, tt.wantMin)
			assert.LessOrEqual(t, group.value, tt.wantMax)
		})
	}
}
