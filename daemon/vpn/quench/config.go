//go:build quench

package quench

import (
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
	"github.com/NordSecurity/nordvpn-linux/log"
)

type Spec struct {
	TlsDomain string `json:"tls_domain"`
	EnableECH bool   `json:"enable_ech"`
}

type Protocol struct {
	Addr string `json:"addr"`
	Spec Spec   `json:"spec"`
}

type Config struct {
	Protocol Protocol `json:"protocol"`
}

// NordWhisperConfig is responsible with fetching the remote configuration for NordWhisper
type NordWhisperConfig struct {
	cm                 config.Manager
	remoteConfigGetter remote.ConfigGetter
}

// NewNordWhisperConfig builds a NordWhisperConfig struct
func NewNordWhisperConfig(cm config.Manager, rc remote.ConfigGetter) *NordWhisperConfig {
	return &NordWhisperConfig{
		cm:                 cm,
		remoteConfigGetter: rc,
	}
}

// GetConfig is the implementation of NordWhisperConfigGetter interface.
//
// ECH is gated by both remote config and the user setting: the effective value is
// remoteECH AND userECH. The remote config acts as a global kill switch (it can force ECH
// off regardless of user preference), while the user can turn ECH off locally. Both default
// to true (matching the previous remote-config-only behavior) when unavailable.
func (qc *NordWhisperConfig) GetConfig() (vpn.NordWhisperFeatureConfig, error) {
	// Remote config gate. Default to enabled if the param is missing or malformed, preserving
	// the historical default (see vpn.NewNordWhisperFeatureConfig).
	remoteECH := vpn.NewNordWhisperFeatureConfig().EnableECH
	enableECHParam, err := qc.remoteConfigGetter.GetFeatureParam(remote.FeatureNordWhisper, "enable_ech")
	if err == nil {
		if parsed, parseErr := strconv.ParseBool(enableECHParam); parseErr == nil {
			remoteECH = parsed
		} else {
			log.Warn("parsing remote enable_ech, defaulting to enabled:", parseErr)
		}
	} else {
		log.Warn("fetching remote enable_ech, defaulting to enabled:", err)
	}

	// User setting. TrueField defaults to true when unset.
	userECH := true
	var cfg config.Config
	if err := qc.cm.Load(&cfg); err == nil {
		userECH = cfg.ECH.Get()
	} else {
		log.Warn("loading config for ECH setting, defaulting to enabled:", err)
	}

	return vpn.NordWhisperFeatureConfig{
		EnableECH: remoteECH && userECH,
	}, nil
}
