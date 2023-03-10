package mock

// Fileshare is a stub
type Fileshare struct{}

// Enable is a stub
func (Fileshare) Enable(uint32, uint32) error {
	return nil
}

// Disable is a stub
func (Fileshare) Disable(uint32, uint32) error {
	return nil
}

// Stop is a stub
func (Fileshare) Stop(uint32, uint32) error {
	return nil
}
