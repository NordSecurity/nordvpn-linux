//go:build telio

package main

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
)

func getVpnFactory(eventsDbPath string, fwmark uint32, envIsDev bool,
	telioCfg remote.RemoteConfigGetter, deviceID, appVersion string) daemon.FactoryFunc {
	var telio = libtelio.New(!envIsDev, eventsDbPath, fwmark, telioCfg, deviceID, appVersion)
	return func(tech config.Technology) (vpn.VPN, error) {
		switch tech {
		case config.Technology_NORDLYNX:
			return telio, nil
		case config.Technology_OPENVPN:
			return openvpn.New(fwmark), nil
		default:
			return nil, errors.New("no such technology")
		}
	}
}

func meshnetImplementation(fn daemon.FactoryFunc) (meshnet.Mesh, error) {
	vpn, err := fn(config.Technology_NORDLYNX)
	if err != nil {
		return nil, err
	}

	mesh, ok := vpn.(meshnet.Mesh)
	if !ok {
		return nil, errors.New("not a meshnet")
	}

	return mesh, nil
}

func keygenImplementation(fn daemon.FactoryFunc) (meshnet.KeyGenerator, error) {
	vpn, err := fn(config.Technology_NORDLYNX)
	if err != nil {
		return nil, err
	}

	keygen, ok := vpn.(meshnet.KeyGenerator)
	if !ok {
		return nil, errors.New("not a keygen")
	}

	return keygen, nil
}
