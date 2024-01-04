package cli

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestParseTimespan(t *testing.T) {
	category.Set(t, category.Unit)

	type reply struct {
		years   int
		months  int
		days    int
		seconds int
		err     error
	}

	tests := []struct {
		input    string
		expected reply
	}{
		{"", reply{0, 0, 0, 0, fmt.Errorf("Time span parsing error: ''")}},
		{"Z", reply{0, 0, 0, 0, fmt.Errorf("Time span parsing error: 'Z'")}},
		{"1Z", reply{0, 0, 0, 0, fmt.Errorf("Time span unit parsing error: 'Z'")}},
		{"1 Z", reply{0, 0, 0, 0, fmt.Errorf("Time span unit parsing error: 'Z'")}},
		{"0", reply{0, 0, 0, 0, nil}},
		{"1", reply{0, 0, 0, 1, nil}},
		{"0s", reply{0, 0, 0, 0, nil}},
		{"1s", reply{0, 0, 0, 1, nil}},
		{"0 s", reply{0, 0, 0, 0, nil}},
		{"1 s", reply{0, 0, 0, 1, nil}},
		{"1m", reply{0, 0, 0, 60, nil}},
		{"1min", reply{0, 0, 0, 60, nil}},
		{"1h", reply{0, 0, 0, 3600, nil}},
		{"1hour", reply{0, 0, 0, 3600, nil}},
		{"1d", reply{0, 0, 1, 0, nil}},
		{"2 days", reply{0, 0, 2, 0, nil}},
		{"1w", reply{0, 0, 7, 0, nil}},
		{"2 weeks", reply{0, 0, 14, 0, nil}},
		{"1M", reply{0, 1, 0, 0, nil}},
		{"2 months", reply{0, 2, 0, 0, nil}},
		{"1y", reply{1, 0, 0, 0, nil}},
		{"2 years", reply{2, 0, 0, 0, nil}},
		{"1 y 2 M 3 weeks 2days 10h 15 min 3s", reply{1, 2, 23, 36903, nil}},
	}

	for _, data := range tests {
		years, months, days, seconds, err := parseTimespan(data.input)
		assert.Equal(t, data.expected.years, years, data.input)
		assert.Equal(t, data.expected.months, months, data.input)
		assert.Equal(t, data.expected.days, days, data.input)
		assert.Equal(t, data.expected.seconds, seconds, data.input)
		assert.Equal(t, data.expected.err, err, data.input)
	}
}
