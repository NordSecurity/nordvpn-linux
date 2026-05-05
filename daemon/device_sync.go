package daemon

import "fmt"

// RegisterDedicatedServers registers the device for dedicated servers if dedicated servers service is available to the user.
func (r *RPC) RegisterDedicatedServers() error {
	dedicatedServersService, err := r.ac.GetDedicatedServersService()
	if err != nil {
		return fmt.Errorf("checking service status: %w", err)
	}

	if !dedicatedServersService.Active {
		return nil
	}

	data := r.dedicatedServersKeyManager.CheckAndRegisterDedicatedServers()
	if data == nil {
		return fmt.Errorf("failed to register the device")
	}

	return nil
}
