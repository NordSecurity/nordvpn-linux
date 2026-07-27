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
// remoteECH AND userECH. The remote config acts as a global switch (it can force ECH
// off regardless of user preference), while the user can turn ECH off locally.
func (qc *NordWhisperConfig) GetConfig() (vpn.NordWhisperFeatureConfig, error) {
	featureCfg := vpn.NewNordWhisperFeatureConfig()
	defaultEchVal := featureCfg.EnableECH

	enableECHParam, err := qc.remoteConfigGetter.GetFeatureParam(remote.FeatureNordWhisper, "enable_ech")
	if err == nil {
		if parsed, parseErr := strconv.ParseBool(enableECHParam); parseErr == nil {
			featureCfg.EnableECH = parsed
		} else {
			log.Warn("parsing remote enable_ech, defaulting to: ", defaultEchVal, " err:", parseErr)
		}
	} else {
		log.Warn("fetching remote enable_ech, defaulting to: ", defaultEchVal, " err:", err)
	}

	if !featureCfg.EnableECH {
		return featureCfg, nil
	}

	var cfg config.Config
	if err := qc.cm.Load(&cfg); err == nil {
		featureCfg.EnableECH = cfg.ECH.Get()
	} else {
		log.Warn("loading config for ECH setting, defaulting to: ", defaultEchVal, " err:", err)
	}

	return featureCfg, nil
}
