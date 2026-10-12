package main

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/log"
)

var ErrNordWhisperDisabled = errors.New("NordWhisper technology was disabled in compile time")

func getVpnFactory(
	eventsDbPath string,
	fwmark uint32,
	envIsDev bool,
	libtelioCfg vpn.LibConfigGetter,
	libquenchCfg vpn.NordWhisperConfigGetter,
	appVersion string,
	eventsPublisher *vpn.Events,
	ensEnabled bool,
) daemon.FactoryFunc {
	nordlynxVPN, nordLynxErr := getNordlynxVPN(
		envIsDev,
		eventsDbPath,
		fwmark,
		libtelioCfg,
		appVersion,
		eventsPublisher,
		ensEnabled)

	if nordLynxErr != nil {
		// don't exit with `err` here in case the factory will be called with
		// vpn protocol different than `config.VPNProtocol_VPN_PROTOCOL_NORDLYNX`
		log.Error("getting NordLynx vpn:", nordLynxErr)
	}

	nordWhisperVPN, nordWhisperErr := getNordWhisperVPN(fwmark, envIsDev, eventsPublisher, libquenchCfg)
	if nordWhisperErr != nil {
		log.Error("getting NordWhisper vpn:", nordWhisperErr)
	}

	return func(proto config.VPNProtocol) (vpn.VPN, error) {
		switch proto {
		case config.VPNProtocol_VPN_PROTOCOL_NORDLYNX:
			return nordlynxVPN, nordLynxErr

		case config.VPNProtocol_VPN_PROTOCOL_OPENVPN_TCP, config.VPNProtocol_VPN_PROTOCOL_OPENVPN_UDP:
			return openvpn.New(fwmark, eventsPublisher), nil

		case config.VPNProtocol_VPN_PROTOCOL_NORDWHISPER:
			return nordWhisperVPN, nordWhisperErr

		case config.VPNProtocol_VPN_PROTOCOL_UNSPECIFIED:
			fallthrough
		default:
			return nil, errors.New("no such technology")
		}
	}
}
