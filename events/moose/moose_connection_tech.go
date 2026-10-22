//go:build moose

package moose

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
)

var errUnknownVPNProtocol = errors.New("unknown vpn protocol")

// isVPNConnected reports whether a VPN tunnel is up. connectedVPNProtocol is set only on a
// successful VPN connect and cleared on disconnect.
func (s *Subscriber) isVPNConnected() bool {
	return s.connectedVPNProtocol != config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED
}

// initVPNProtocol reports the configured VPN protocol on Init. Called with s.mux held.
func (s *Subscriber) initVPNProtocol(cfg config.Config) error {
	if cfg.VPNProtocol == config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED {
		return fmt.Errorf("setting moose vpn protocol: %w", errUnknownVPNProtocol)
	}
	s.configuredVPNProtocol = cfg.VPNProtocol
	if err := s.response(s.mooseFuncs.setProtocolUserPreference(connectionProtocolToInternalType(cfg.VPNProtocol))); err != nil {
		return fmt.Errorf("setting moose protocol: %w", err)
	}
	if err := s.response(s.mooseFuncs.setTechnologyUserPreference(connectionTechnologyToInternalType(cfg.VPNProtocol))); err != nil {
		return fmt.Errorf("setting moose technology: %w", err)
	}
	effective := s.configuredVPNProtocol
	if s.isVPNConnected() {
		effective = s.connectedVPNProtocol
	}
	if err := s.reportEffectiveConnection(effective); err != nil {
		return fmt.Errorf("setting moose connection state: %w", err)
	}
	return nil
}

// NotifyVPNProtocol reports the VPN protocol as separate moose technology and protocol values.
func (s *Subscriber) NotifyVPNProtocol(data config.VPNProtocol) error {
	if data == config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED {
		return errUnknownVPNProtocol
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.configuredVPNProtocol = data
	if err := s.response(s.mooseFuncs.setTechnologyUserPreference(connectionTechnologyToInternalType(data))); err != nil {
		return fmt.Errorf("setting technology user preference (%v): %w", data, err)
	}
	if err := s.response(s.mooseFuncs.setProtocolUserPreference(connectionProtocolToInternalType(data))); err != nil {
		return fmt.Errorf("setting protocol user preference (%v): %w", data, err)
	}
	if s.isVPNConnected() {
		// while connected the change takes effect on the next connect or on disconnect
		return nil
	}
	if err := s.reportEffectiveConnection(s.configuredVPNProtocol); err != nil {
		return fmt.Errorf("setting vpn protocol current state (%v): %w", data, err)
	}
	return nil
}

// reportEffectiveConnection sets current state of technology/protocol.
func (s *Subscriber) reportEffectiveConnection(vpnProtocol config.VPNProtocol) error {
	if vpnProtocol == config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED {
		return nil
	}

	var errs []error
	if err := s.response(s.mooseFuncs.setTechnologyCurrentState(connectionTechnologyToInternalType(vpnProtocol))); err != nil {
		errs = append(errs, fmt.Errorf("setting technology current state (%v): %w", vpnProtocol, err))
	}
	if err := s.response(s.mooseFuncs.setProtocolCurrentState(connectionProtocolToInternalType(vpnProtocol))); err != nil {
		errs = append(errs, fmt.Errorf("setting protocol current state (%v): %w", vpnProtocol, err))
	}
	return errors.Join(errs...)
}
