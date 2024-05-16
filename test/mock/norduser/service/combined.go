package service

type MockNorduserCombinedService struct{}

func (m *MockNorduserCombinedService) Enable(uint32, uint32, string) error { return nil }

func (m *MockNorduserCombinedService) Disable(uint32) error { return nil }

func (m *MockNorduserCombinedService) Stop(uint32, bool) error { return nil }

func (m *MockNorduserCombinedService) Restart(uint32) error { return nil }

func (m *MockNorduserCombinedService) StopAll() {}

func (m *MockNorduserCombinedService) DisableAll() {}
