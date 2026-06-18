//go:build telio

package main

import (
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/nordlynx/libtelio"
	"github.com/NordSecurity/nordvpn-linux/meshnet"
)

func getNordlynxVPN(
	envIsDev bool,
	eventsDbPath string,
	fwmark uint32,
	cfg vpn.LibConfigGetter,
	appVersion string,
	eventsPublisher *vpn.Events,
	ensEnabledFn func() bool,
) (*libtelio.Holder, error) {
	telio, err := libtelio.NewHolder(
		!envIsDev,
		eventsDbPath,
		fwmark,
		cfg,
		appVersion,
		eventsPublisher,
		ensEnabledFn,
	)
	if err != nil {
		return telio, fmt.Errorf("creating telio instance: %w", err)
	}
	return telio, nil
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
