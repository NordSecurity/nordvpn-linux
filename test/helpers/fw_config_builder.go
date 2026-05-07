package helpers

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
)

type FirewallConfigBuilder struct {
	cfg firewall.Config
}

func NewFWConfig() *FirewallConfigBuilder {
	return &FirewallConfigBuilder{}
}

func (b *FirewallConfigBuilder) KillSwitch() *FirewallConfigBuilder {
	b.cfg.KillSwitch = true
	return b
}

func (b *FirewallConfigBuilder) TunnelInterface(iface string) *FirewallConfigBuilder {
	b.cfg.TunnelInterface = iface
	return b
}

func (b *FirewallConfigBuilder) AllowlistTCPPort(port int64) *FirewallConfigBuilder {
	if b.cfg.Allowlist.Ports.TCP == nil {
		b.cfg.Allowlist.Ports.TCP = config.PortSet{}
	}
	b.cfg.Allowlist.Ports.TCP[port] = true
	return b
}

func (b *FirewallConfigBuilder) AllowlistUDPPort(port int64) *FirewallConfigBuilder {
	if b.cfg.Allowlist.Ports.UDP == nil {
		b.cfg.Allowlist.Ports.UDP = config.PortSet{}
	}
	b.cfg.Allowlist.Ports.UDP[port] = true
	return b
}

func (b *FirewallConfigBuilder) AllowlistSubnet(subnet string) *FirewallConfigBuilder {
	b.cfg.Allowlist.Subnets = append(b.cfg.Allowlist.Subnets, subnet)
	return b
}

func (b *FirewallConfigBuilder) BlockFileshare() *FirewallConfigBuilder {
	b.cfg.BlockFileshare = true
	return b
}

func (b *FirewallConfigBuilder) Meshnet(iface string) *FirewallConfigBuilder {
	b.cfg.MeshnetInfo = &firewall.MeshInfo{MeshInterface: iface}
	return b
}

func (b *FirewallConfigBuilder) MeshPeer(peer mesh.MachinePeer) *FirewallConfigBuilder {
	if b.cfg.MeshnetInfo == nil {
		b.cfg.MeshnetInfo = &firewall.MeshInfo{}
	}
	b.cfg.MeshnetInfo.MeshnetMap.Peers = append(b.cfg.MeshnetInfo.MeshnetMap.Peers, peer)
	return b
}

func (b *FirewallConfigBuilder) Build() firewall.Config {
	return b.cfg
}
