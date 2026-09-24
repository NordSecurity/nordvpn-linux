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
	return s.connectedTechProto.technology != config.Technology_UNKNOWN_TECHNOLOGY
}

// initTechProto reports the configured technology/protocol on Init. Called with s.mux held.
func (s *Subscriber) initTechProto(cfg config.Config) error {
	if cfg.Technology == config.Technology_UNKNOWN_TECHNOLOGY {
		return fmt.Errorf("setting moose technology: %w", errUnknownTechnology)
	}
	s.configuredTechProto = techProto{technology: cfg.Technology, protocol: cfg.AutoConnectData.Protocol}
	if err := s.response(s.mooseFuncs.setProtocolUserPreference(connectionProtocolToInternalType(cfg.AutoConnectData.Protocol))); err != nil {
		return fmt.Errorf("setting moose protocol: %w", err)
	}
	if err := s.response(s.mooseFuncs.setTechnologyUserPreference(connectionTechnologyToInternalType(cfg.Technology))); err != nil {
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

func (s *Subscriber) NotifyTechnology(data config.Technology) error {
	if data == config.Technology_UNKNOWN_TECHNOLOGY {
		return errUnknownTechnology
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.configuredTechProto.technology = data
	if err := s.response(s.mooseFuncs.setTechnologyUserPreference(connectionTechnologyToInternalType(data))); err != nil {
		return fmt.Errorf("setting technology user preference (%v): %w", data, err)
	}
	if s.isVPNConnected() {
		// while connected the change takes effect on the next connect or on disconnect
		return nil
	}
	if err := s.reportEffectiveConnection(s.configuredTechProto); err != nil {
		return fmt.Errorf("setting technology current state (%v): %w", data, err)
	}
	return nil
}

func (s *Subscriber) NotifyProtocol(data config.Protocol) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.configuredTechProto.protocol = data
	if err := s.response(s.mooseFuncs.setProtocolUserPreference(connectionProtocolToInternalType(data))); err != nil {
		return fmt.Errorf("setting protocol user preference (%v): %w", data, err)
	}
	if s.isVPNConnected() {
		// while connected the change takes effect on the next connect or on disconnect
		return nil
	}
	if err := s.reportEffectiveConnection(s.configuredTechProto); err != nil {
		return fmt.Errorf("setting protocol current state (%v): %w", data, err)
	}
	return nil
}

// reportEffectiveConnection sets current state of technology/protocol and the derived obfuscation metric.
func (s *Subscriber) reportEffectiveConnection(c techProto) error {
	var errs []error

	if c.technology != config.Technology_UNKNOWN_TECHNOLOGY {
		if err := s.response(s.mooseFuncs.setTechnologyCurrentState(connectionTechnologyToInternalType(c.technology))); err != nil {
			errs = append(errs, fmt.Errorf("setting technology current state (%v): %w", c.technology, err))
		}
		obfuscated := c.technology == config.Technology_NORDWHISPER
		if err := s.response(s.mooseFuncs.setObfuscationEnabledUserPreference(obfuscated)); err != nil {
			errs = append(errs, fmt.Errorf("setting obfuscation preference (enabled=%v): %w", obfuscated, err))
		}
	}

	if c.protocol != config.Protocol_UNKNOWN_PROTOCOL {
		if err := s.response(s.mooseFuncs.setProtocolCurrentState(connectionProtocolToInternalType(c.protocol))); err != nil {
			errs = append(errs, fmt.Errorf("setting protocol current state (%v): %w", c.protocol, err))
		}
	}

	return errors.Join(errs...)
}
