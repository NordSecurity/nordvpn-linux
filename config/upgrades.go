package config

import (
	"encoding/json"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/google/uuid"
)

type configV3 struct {
	Technology   Technology `json:"technology,omitempty"`
	Firewall     bool       `json:"firewall"` // omitempty breaks this
	FirewallMark uint32     `json:"fwmark"`
	Routing      TrueField  `json:"routing"`
	Analytics    TrueField  `json:"analytics"`
	Mesh         bool       `json:"mesh"`
	// MeshPrivateKey is base64 encoded
	MeshPrivateKey  string              `json:"mesh_private_key"`
	MeshDevice      *mesh.Machine       `json:"mesh_device"`
	KillSwitch      bool                `json:"kill_switch,omitempty"`
	AutoConnect     bool                `json:"auto_connect,omitempty"`
	IPv6            bool                `json:"ipv6"`
	Meshnet         meshnet             `json:"meshnet"`
	AutoConnectData AutoConnectData     `json:"auto_connect_data"` // omitempty breaks this
	UsersData       *UsersData          `json:"users_data,omitempty"`
	TokensData      map[int64]TokenData `json:"tokens_data,omitempty"`
	MachineID       uuid.UUID           `json:"machine_id,omitempty"`
	LanDiscovery    bool                `json:"lan_discovery"`
	RemoteConfig    string              `json:"remote_config,omitempty"`
	RCLastUpdate    time.Time           `json:"rc_last_update,omitempty"`
	// Indicates whether the virtual servers are used. True by default
	VirtualLocation TrueField `json:"virtual_location,omitempty"`
}

func (f *FilesystemConfigManager) upgradeIfNeeded(raw []byte) (*Config, error) {
	var header header
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, err
	}

	if isUpgradeNeeded(&header) {
		// if upgrade is needed, do the upgrade and save new config
		cfg, err := upgrade(&header, raw)
		if err != nil {
			return nil, err
		}
		if err := f.save(*cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	// no upgrade was needed, just unmarshal to regular [config.Config]
	var cfg *Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func isUpgradeNeeded(header *header) bool {
	// XXX: move 4 to constant or use `DaemonApiVersion_CURRENT_VERSION`
	return header.Version == nil || *header.Version != 4
}

func upgrade(header *header, raw []byte) (*Config, error) {
	var cfg *Config
	switch header.Version {
	// upgrade from v3 to v4
	case nil:
		var old configV3
		if err := json.Unmarshal(raw, &old); err != nil {
			return nil, err
		}
		cfg = upgradeFromV3(&old)
	}
	return cfg, nil
}

func upgradeFromV3(old *configV3) *Config {
	return &Config{
		Version:         4,
		Technology:      old.Technology,
		Firewall:        old.Firewall,
		FirewallMark:    old.FirewallMark,
		Routing:         old.Routing,
		Analytics:       defaultAnalytics(),
		Mesh:            old.Mesh,
		MeshPrivateKey:  old.MeshPrivateKey,
		MeshDevice:      old.MeshDevice,
		KillSwitch:      old.KillSwitch,
		AutoConnect:     old.AutoConnect,
		IPv6:            old.IPv6,
		Meshnet:         old.Meshnet,
		AutoConnectData: old.AutoConnectData,
		UsersData:       old.UsersData,
		TokensData:      old.TokensData,
		MachineID:       old.MachineID,
		LanDiscovery:    old.LanDiscovery,
		RemoteConfig:    old.RemoteConfig,
		RCLastUpdate:    old.RCLastUpdate,
		VirtualLocation: old.VirtualLocation,
	}
}

func defaultAnalytics() Analytics {
	// XXX: update when model is finished
	return Analytics{}
}
