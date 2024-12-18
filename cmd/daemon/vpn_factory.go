package main

import (
	"errors"
	"log"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

var ErrNordWhisperDisabled = errors.New("NordWhisper technology was disabled in compile time")

func getVpnFactory(eventsDbPath string, fwmark uint32, envIsDev bool,
	cfg vpn.LibConfigGetter, appVersion string, eventsPublisher *vpn.Events,
) daemon.FactoryFunc {
	nordlynxVPN, nordLynxErr := getNordlynxVPN(envIsDev, eventsDbPath, fwmark, cfg, appVersion, eventsPublisher)
	if nordLynxErr != nil {
		// don't exit with `err` here in case the factory will be called with
		// technology different than `config.Technology_NORDLYNX`
		log.Println(internal.ErrorPrefix, "getting NordLynx vpn:", nordLynxErr)
	}

	nordWhisperVPN, nordWhisperErr := getNordWhisperVPN(fwmark)
	if nordWhisperErr != nil {
		log.Println(internal.ErrorPrefix, "getting NordWhisper vpn:", nordWhisperErr)
	}

	return func(tech config.Technology) (vpn.VPN, error) {
		switch tech {
		case config.Technology_NORDLYNX:
			return nordlynxVPN, nordLynxErr
		case config.Technology_OPENVPN:
			return openvpn.New(fwmark, eventsPublisher), nil
		case config.Technology_NORDWHISPER:
			return nordWhisperVPN, nordWhisperErr
		case config.Technology_UNKNOWN_TECHNOLOGY:
			fallthrough
		default:
			return nil, errors.New("no such technology")
		}
	}
}
