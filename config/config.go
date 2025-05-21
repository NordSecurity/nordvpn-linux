// Package config provides functions for managing configuration of the daemon application.
package config

import (
	"time"

	"github.com/google/uuid"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

const defaultFWMarkValue uint32 = 0xe1f1

func newConfig(machineIDGetter MachineIDGetter) *Config {
	// XXX: new config needs to be updatd with breaking changes
	return &Config{
		Version:      4, // XXX: move to something like "currentVersion"
		Technology:   Technology_NORDLYNX,
		Firewall:     true,
		FirewallMark: defaultFWMarkValue,
		AutoConnectData: AutoConnectData{
			Protocol: Protocol_UDP,
		},
		MachineID:  machineIDGetter.GetMachineID(),
		UsersData:  &UsersData{NotifyOff: UidBoolMap{}, TrayOff: UidBoolMap{}},
		TokensData: map[int64]TokenData{},
	}
}

// newConfigWithLoginData returns a clean/default config where login data(TokensData and AutoconnectData.ID) is
// initialized to values from the parrentConfig.
func newConfigWithLoginData(machineIDGetter MachineIDGetter, parrentConfig Config) *Config {
	cfg := newConfig(machineIDGetter)
	cfg.AutoConnectData.ID = parrentConfig.AutoConnectData.ID
	cfg.TokensData = parrentConfig.TokensData
	return cfg
}

// Config stores application settings and tokens.
//
// Config should be evolved is such a way, that it does not
// require any use of constructors by the caller.
type Config struct {
	Version      int        `json:"version"`
	Technology   Technology `json:"technology,omitempty"`
	Firewall     bool       `json:"firewall"` // omitempty breaks this
	FirewallMark uint32     `json:"fwmark"`
	Routing      TrueField  `json:"routing"`
	Analytics    Analytics  `json:"analytics"`
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

type AutoConnectData struct {
	ID        int64  `json:"id,omitempty"`
	ServerTag string `json:"server_tag,omitempty"`
	Country   string
	City      string
	Group     ServerGroup
	Protocol  Protocol `json:"protocol,omitempty"`
	// TODO: rename json key when v4 comes out.
	ThreatProtectionLite bool      `json:"cybersec,omitempty"`
	Obfuscate            bool      `json:"obfuscate,omitempty"`
	DNS                  DNS       `json:"dns,omitempty"`
	Allowlist            Allowlist `json:"whitelist,omitempty"`
	PostquantumVpn       bool      `json:"postquantum_vpn"`
}

type DNS []string

// Or provides defaultValue in case of an empty/nil slice.
// Inspired by https://doc.rust-lang.org/std/option/enum.Option.html#method.or
func (d DNS) Or(defaultValue []string) DNS {
	if len(d) == 0 { // also covers nil slices
		return DNS(defaultValue)
	}
	return d
}

type NCData struct {
	UserID         uuid.UUID `json:"user_id,omitempty"`
	Username       string    `json:"username,omitempty"`
	Password       string    `json:"password,omitempty"`
	Endpoint       string    `json:"endpoint,omitempty"`
	ExpirationDate time.Time `json:"timestamp,omitempty"`
}

type meshnet struct {
	EnabledByUID uint32 `json:"enabled_by_uid"` // Linux user which enabled meshnet
	EnabledByGID uint32 `json:"enabled_by_gid"` // Group of Linux user which enabled meshnet
}

func (d *NCData) IsUserIDEmpty() bool {
	return d.UserID == uuid.Nil
}

type Analytics struct {
	Basic    bool `json:"basic"`
	Extended bool `json:"extended"`
}

func (a Analytics) Get() bool {
	return a.Basic
}

func (a *Analytics) Set(val bool) {
	a.Basic = val
}

// XXX: docs
type header struct {
	Version *int `json:"version,omitempty"`
}
