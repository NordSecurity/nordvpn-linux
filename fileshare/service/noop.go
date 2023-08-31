package service

// NoopFileshare is a stub
type NoopFileshare struct{}

// Enable is a stub
func (NoopFileshare) Enable(uint32, uint32) error {
	return nil
}

// Disable is a stub
func (NoopFileshare) Disable(uint32, uint32) error {
	return nil
}

// Stop is a stub
func (NoopFileshare) Stop(uint32, uint32) error {
	return nil
}
