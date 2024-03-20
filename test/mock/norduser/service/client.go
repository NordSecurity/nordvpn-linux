package service

type MockNorduserClient struct {
	startFileshareErr error
}

func NewMockNorduserClient(startFileshareErr error) *MockNorduserClient {
	return &MockNorduserClient{
		startFileshareErr: startFileshareErr,
	}
}

func (m *MockNorduserClient) StartFileshare(uid uint32) error {
	return m.startFileshareErr
}

func (m *MockNorduserClient) StopFileshare(uid uint32) error {
	return m.startFileshareErr
}
