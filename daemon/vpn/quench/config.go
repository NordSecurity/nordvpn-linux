//go:build quench

package quench

import (
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/config/remote"
	"github.com/NordSecurity/nordvpn-linux/daemon/vpn"
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
	remoteConfigGetter remote.ConfigGetter
}

// NewNordWhisperConfig builds a NordWhisperConfig struct
func NewNordWhisperConfig(rc remote.ConfigGetter) *NordWhisperConfig {
	return &NordWhisperConfig{
		remoteConfigGetter: rc,
	}
}

// GetConfig is the implementation of NordWhisperConfigGetter interface
// It fetches the remote config for NordWhisper
func (qc *NordWhisperConfig) GetConfig() (vpn.NordWhisperFeatureConfig, error) {
	enableECHParam, err := qc.remoteConfigGetter.GetFeatureParam(remote.FeatureNordWhisper, "enable_ech")
	if err != nil {
		return vpn.NewNordWhisperFeatureConfig(), err
	}

	enableECH, err := strconv.ParseBool(enableECHParam)
	if err != nil {
		return vpn.NewNordWhisperFeatureConfig(), err
	}

	return vpn.NordWhisperFeatureConfig{
		EnableECH: enableECH,
	}, nil
}
