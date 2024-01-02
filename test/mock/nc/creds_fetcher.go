package nc

import "time"

type MockTime struct {
	SecondsSinceTimestamp int
}

func (m *MockTime) GetDurationSinceTimestamp(int64) time.Duration {
	return time.Duration(m.SecondsSinceTimestamp) * time.Second
}
