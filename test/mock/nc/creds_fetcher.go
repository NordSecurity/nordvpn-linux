package nc

import "time"

type MockTime struct {
	SecondsSinceTimestamp int64
}

func (m *MockTime) GetDurationSinceTimestamp(time.Time) time.Duration {
	return time.Duration(m.SecondsSinceTimestamp) * time.Second
}
