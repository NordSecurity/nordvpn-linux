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

func getVpnFactory(eventsDbPath string, fwmark uint32, envIsDev bool,
	cfg vpn.LibConfigGetter, appVersion string, eventsPublisher *vpn.Events,
) daemon.FactoryFunc {
	nordlynxVPN, err := getNordlynxVPN(envIsDev, eventsDbPath, fwmark, cfg, appVersion, eventsPublisher)
	if err != nil {
		// don't exit with `err` here in case the factory will be called with
		// technology different than `config.Technology_NORDLYNX`
		log.Println(internal.ErrorPrefix, "getting NordLynx vpn:", err)
	}
	return func(tech config.Technology) (vpn.VPN, error) {
		switch tech {
		case config.Technology_NORDLYNX:
			return nordlynxVPN, nil
		case config.Technology_OPENVPN:
			return openvpn.New(fwmark, eventsPublisher), nil
		case config.Technology_QUENCH:
			return getQuenchVPN(fwmark)
		case config.Technology_UNKNOWN_TECHNOLOGY:
			fallthrough
		default:
			return nil, errors.New("no such technology")
		}
	}
}
