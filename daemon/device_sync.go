package daemon

import "fmt"

// RegisterDedicatedServers registers the device for dedicated servers if dedicated servers service is available to the user.
func (r *RPC) RegisterDedicatedServers() error {
	dedicatedServerService, err := r.ac.GetDedicatedServerService()
	if err != nil {
		return fmt.Errorf("checking service status: %w", err)
	}

	if !dedicatedServerService.Active {
		return nil
	}

	data := r.dedicatedServerKeyManager.CheckAndRegisterDedicatedServers()
	if data == nil {
		return fmt.Errorf("failed to register the device")
	}

	return nil
}
