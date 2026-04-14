package firewall

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

// Service adapts system firewall configuration to firewall rules
//
// Used by callers.
type Service interface {
	Configure(config Config) error
	Flush() error
	Disable() error
	Enable() error
}

type FirewallBackend interface {
	Configure(config Config) error
	Flush() error
}

// Config keeps all the information needed to configure the firewall
type Config struct {
	TunnelInterface string
	Allowlist       config.Allowlist
	KillSwitch      bool
	MeshnetInfo     *MeshInfo
}

type MeshInfo struct {
	MeshnetMap     mesh.MachineMap
	MeshInterface  string
	BlockFileshare bool
}

func NewMeshInfo(meshnetMap mesh.MachineMap, meshInterface string) *MeshInfo {
	return &MeshInfo{
		MeshnetMap:    meshnetMap,
		MeshInterface: meshInterface,
	}
}

func (c Config) CopyWith(opts ...Option) Config {
	cfg := c

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func (c *Config) IsEmpty() bool {
	return !c.KillSwitch && len(c.TunnelInterface) == 0 && c.MeshnetInfo == nil
}

func (c *Config) IsVpnOrKillSwitchSet() bool {
	return c.KillSwitch || len(c.TunnelInterface) > 0
}

type Option func(*Config)

func WithKillSwitch(v bool) Option {
	return func(c *Config) {
		c.KillSwitch = v
	}
}

func WithAllowlist(allowlist config.Allowlist) Option {
	return func(c *Config) {
		c.Allowlist = allowlist
	}
}

func WithTunnelInterface(tunnelInterface string) Option {
	return func(c *Config) {
		c.TunnelInterface = tunnelInterface
	}
}

func WithMeshnetInfo(meshInfo *MeshInfo) Option {
	return func(c *Config) {
		c.MeshnetInfo = meshInfo
	}
}

func WithBlockFileshare(block bool) Option {
	return func(c *Config) {
		if c.MeshnetInfo != nil {
			c.MeshnetInfo.BlockFileshare = block
		}
	}
}
