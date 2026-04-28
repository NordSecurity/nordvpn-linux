package daemon

import "fmt"

// RegisterDedicatedServers registers the device for dedicated servers if dedicated servers service is available to the user.
func (r *RPC) RegisterDedicatedServers() error {
	hasDedicatedServerService, err := r.ac.HasDedicatedServerService()
	if err != nil {
		return fmt.Errorf("checking service status: %w", err)
	}

	if !hasDedicatedServerService {
		return nil
	}

	data := r.dedicatedServersKeyManager.CheckAndRegisterDedicatedServers()
	if data == nil {
		return fmt.Errorf("failed to register the device")
	}

	return nil
}
