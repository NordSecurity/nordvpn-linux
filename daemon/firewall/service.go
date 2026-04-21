package firewall

import (
	"bytes"
	"slices"

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
	// is controlled by the fileshare process monitoring
	BlockFileshare bool
	MeshnetInfo    *MeshInfo
}

func NewConfig(opts ...Option) Config {
	return Config{}.CopyWith(opts...)
}

// Only the firewall relevant parts are checked. It is not a fully IsEqual
func (c *Config) HasSimilarMeshInfo(cfg *Config) bool {
	return c.MeshnetInfo != nil &&
		cfg.MeshnetInfo != nil &&
		c.MeshnetInfo.IsSimilar(cfg.MeshnetInfo)
}

type MeshInfo struct {
	MeshnetMap    mesh.MachineMap
	MeshInterface string
}

// Only the firewall relevant parts are checked. It is not a fully IsEqual
func (m *MeshInfo) IsSimilar(meshInfo *MeshInfo) bool {
	if meshInfo == nil {
		return false
	}

	if m.MeshInterface != meshInfo.MeshInterface {
		return false
	}

	sortMeshnetMap(&meshInfo.MeshnetMap)

	arePeersEqual := slices.EqualFunc(
		m.MeshnetMap.Peers,
		meshInfo.MeshnetMap.Peers,
		func(p1, p2 mesh.MachinePeer) bool {
			return p1.ID == p2.ID &&
				p1.Address == p2.Address &&
				p1.DoIAllowInbound == p2.DoIAllowInbound &&
				p1.DoIAllowRouting == p2.DoIAllowRouting &&
				p1.DoIAllowLocalNetwork == p2.DoIAllowLocalNetwork &&
				p1.DoIAllowFileshare == p2.DoIAllowFileshare

		},
	)

	if !arePeersEqual {
		return false
	}

	return true
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
		if meshInfo != nil {
			sortMeshnetMap(&c.MeshnetInfo.MeshnetMap)
		}
	}
}

func WithBlockFileshare(block bool) Option {
	return func(c *Config) {
		c.BlockFileshare = block
	}
}

func sortMeshnetMap(meshMap *mesh.MachineMap) {
	// sort the peers to easier compare for equality
	slices.SortFunc(meshMap.Peers, func(p1, p2 mesh.MachinePeer) int {
		return bytes.Compare(p1.ID[:], p2.ID[:])
	})
}
