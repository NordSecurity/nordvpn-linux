package daemon

import (
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestRandFloat(t *testing.T) {
	category.Set(t, category.Unit)

	testData := []struct {
		min, max float64
	}{
		{0, 1},
		{-3, -1},
		{15, 2000},
		{5.5, 5.6},
		{0.001, 0.002},
	}

	for _, item := range testData {
		for i := 0; i < 500; i++ {
			got := randFloat(time.Now().Unix(), item.min, item.max)
			assert.Greater(t, got, item.min)
			assert.Less(t, got, item.max)
		}
	}
}
