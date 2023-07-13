package service

// MockFileshare is a stub
type MockFileshare struct{}

// Enable is a stub
func (MockFileshare) Enable(uint32, uint32) error {
	return nil
}

// Disable is a stub
func (MockFileshare) Disable(uint32, uint32) error {
	return nil
}

// Stop is a stub
func (MockFileshare) Stop(uint32, uint32) error {
	return nil
}
