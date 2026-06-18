package main

import (
	"errors"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn/openvpn"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

var ErrNordWhisperDisabled = errors.New("NordWhisper technology was disabled in compile time")

type rcForVPN interface {
	remote.FeatureConfig
	Subscribe(remote.RemoteConfigNotifier)
}

func getVpnFactory(
	eventsDBPath string,
	fwmark uint32,
	envIsDev bool,
	libtelioCfg vpn.LibConfigGetter,
	libquenchCfg vpn.NordWhisperConfigGetter,
	appVersion string,
	eventsPublisher *vpn.Events,
	rc rcForVPN,
) daemon.FactoryFunc {
	ensEnabledFn := func() bool { return rc.IsFeatureEnabled(remote.FeatureENS) }

	nordlynxVPN, nordLynxErr := getNordlynxVPN(
		envIsDev,
		eventsDBPath,
		fwmark,
		libtelioCfg,
		appVersion,
		eventsPublisher,
		ensEnabledFn,
	)

	if nordLynxErr != nil {
		// don't exit with `err` here in case the factory will be called with
		// technology different than `config.Technology_NORDLYNX`
		log.Println(internal.ErrorPrefix, "getting NordLynx vpn:", nordLynxErr)
	}

	nordWhisperVPN, nordWhisperErr := getNordWhisperVPN(fwmark, envIsDev, eventsPublisher, libquenchCfg)
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

type ensRecreateOnRCChange struct {
	reconnect      func() error
	lastEnsEnabled bool
}

func newEnsRecreateOnRCChange(reconnect func() error, currentEnsEnabled bool) *ensRecreateOnRCChange {
	return &ensRecreateOnRCChange{reconnect: reconnect, lastEnsEnabled: currentEnsEnabled}
}

func (n *ensRecreateOnRCChange) RemoteConfigUpdate(event remote.RemoteConfigEvent) error {
	if event.ENSFeatureEnabled == n.lastEnsEnabled {
		return nil
	}
	n.lastEnsEnabled = event.ENSFeatureEnabled
	log.Info("ENS remote config toggle changed, recreating NordLynx vpn instance")
	return n.reconnect()
}

func wireENSRecreation(
	factory daemon.FactoryFunc,
	recreateVPNFn func(func() error) error,
	rc rcForVPN,
) {
	nordlynxVPN, _ := factory(config.Technology_NORDLYNX)
	if vpn, ok := nordlynxVPN.(vpn.Recreatable); ok {
		rc.Subscribe(newEnsRecreateOnRCChange(
			func() error { return recreateVPNFn(vpn.Recreate) },
			rc.IsFeatureEnabled(remote.FeatureENS),
		))
	}
}
