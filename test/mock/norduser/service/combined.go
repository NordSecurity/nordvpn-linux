package service

type MockNorduserCombinedService struct{}

func (m *MockNorduserCombinedService) Enable(uid uint32, gid uint32) error { return nil }

func (m *MockNorduserCombinedService) Disable(uid uint32) error { return nil }

func (m *MockNorduserCombinedService) Stop(uid uint32) error { return nil }

func (m *MockNorduserCombinedService) StopAll() {}

func (m *MockNorduserCombinedService) DisableAll() {}
