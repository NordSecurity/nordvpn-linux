package daemon

import "fmt"

// SyncDevice registers the device for dedicated servers if dedicated servers service is available to the user.
func (r *RPC) SyncDevice() error {
	hasDedicatedServerService, err := r.ac.HasDedicatedServerService()
	if err != nil {
		return fmt.Errorf("checking service status: %w", err)
	}

	if !hasDedicatedServerService {
		return nil
	}

	ok := r.dedicatedServersKeyManager.CheckAndRegisterDedicatedServers()
	if !ok {
		return fmt.Errorf("failed to registed device")
	}

	return nil
}
