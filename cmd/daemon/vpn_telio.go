//go:build telio

package main

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
)

func getVpnFactory(eventsDbPath string, fwmark uint32, envIsDev bool,
	cfg vpn.LibConfigGetter, deviceID, appVersion string, eventsPublisher *vpn.Events,
) daemon.FactoryFunc {
	return func(tech config.Technology) (vpn.VPN, error) {
		switch tech {
		case config.Technology_NORDLYNX:
			telio, err := libtelio.New(!envIsDev, eventsDbPath, fwmark, cfg, deviceID, appVersion, eventsPublisher)
			if err != nil {
				return nil, fmt.Errorf("libtelio creation failed: %w", err)
			}
			return telio, nil
		case config.Technology_OPENVPN:
			return openvpn.New(fwmark, eventsPublisher), nil
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
