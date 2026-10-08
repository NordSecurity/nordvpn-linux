//go:build moose

package moose

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
)

var errUnknownTechnology = errors.New("unknown technology")

// techProto is the technology/protocol a VPN connection is (or would be) made with.
type techProto struct {
	technology config.Technology
	protocol   config.Protocol
}

// isVPNConnected reports whether a VPN tunnel is up. connectedTechProto is set only on a
// successful VPN connect and cleared on disconnect.
func (s *Subscriber) isVPNConnected() bool {
	return s.connectedTechProto.technology != config.TechnologyUnknown
}

// initTechProto reports the configured technology/protocol on Init. Called with s.mux held.
func (s *Subscriber) initTechProto(cfg config.Config) error {
	technology := cfg.VPNProtocol.Technology()
	if technology == config.TechnologyUnknown {
		return fmt.Errorf("setting moose technology: %w", errUnknownTechnology)
	}
	protocol := analyticsProtocol(cfg.VPNProtocol)
	s.configuredTechProto = techProto{technology: technology, protocol: protocol}
	if err := s.response(s.mooseFuncs.setProtocolUserPreference(connectionProtocolToInternalType(protocol))); err != nil {
		return fmt.Errorf("setting moose protocol: %w", err)
	}
	if err := s.response(s.mooseFuncs.setTechnologyUserPreference(connectionTechnologyToInternalType(technology))); err != nil {
		return fmt.Errorf("setting moose technology: %w", err)
	}
	effective := s.configuredTechProto
	if s.isVPNConnected() {
		effective = s.connectedTechProto
	}
	if err := s.reportEffectiveConnection(effective); err != nil {
		return fmt.Errorf("setting moose connection state: %w", err)
	}
	return nil
}

// NotifyVPNProtocol reports the VPN protocol as separate moose technology and protocol values.
func (s *Subscriber) NotifyVPNProtocol(data config.VPNProtocol) error {
	technology := data.Technology()
	if technology == config.TechnologyUnknown {
		return errUnknownTechnology
	}
	protocol := analyticsProtocol(data)

	s.mux.Lock()
	defer s.mux.Unlock()

	s.configuredTechProto = techProto{technology: technology, protocol: protocol}
	if err := s.response(s.mooseFuncs.setTechnologyUserPreference(connectionTechnologyToInternalType(technology))); err != nil {
		return fmt.Errorf("setting technology user preference (%v): %w", data, err)
	}
	if err := s.response(s.mooseFuncs.setProtocolUserPreference(connectionProtocolToInternalType(protocol))); err != nil {
		return fmt.Errorf("setting protocol user preference (%v): %w", data, err)
	}
	if s.isVPNConnected() {
		// while connected the change takes effect on the next connect or on disconnect
		return nil
	}
	if err := s.reportEffectiveConnection(s.configuredTechProto); err != nil {
		return fmt.Errorf("setting vpn protocol current state (%v): %w", data, err)
	}
	return nil
}

// reportEffectiveConnection sets current state of technology/protocol.
func (s *Subscriber) reportEffectiveConnection(c techProto) error {
	var errs []error

	if c.technology != config.TechnologyUnknown {
		if err := s.response(s.mooseFuncs.setTechnologyCurrentState(connectionTechnologyToInternalType(c.technology))); err != nil {
			errs = append(errs, fmt.Errorf("setting technology current state (%v): %w", c.technology, err))
		}
	}

	if c.protocol != config.Protocol_UNKNOWN_PROTOCOL {
		if err := s.response(s.mooseFuncs.setProtocolCurrentState(connectionProtocolToInternalType(c.protocol))); err != nil {
			errs = append(errs, fmt.Errorf("setting protocol current state (%v): %w", c.protocol, err))
		}
	}

	return errors.Join(errs...)
}
