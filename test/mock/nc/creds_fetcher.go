package nc

type MockTime struct {
	SecondsSinceTimestamp float64
}

func (m *MockTime) GetSecondsSinceTimestamp(int64) float64 {
	return m.SecondsSinceTimestamp
}
