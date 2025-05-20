package config

import (
	"os"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesystem(t *testing.T) {
	category.Set(t, category.File)

	tests := []struct {
		name string
		f    SaveFunc
	}{
		{
			name: "autoconnect data is saved",
			f: func(c Config) Config {
				c.AutoConnectData.ThreatProtectionLite = true
				return c
			},
		},
		{
			name: "tokens data is saved",
			f: func(c Config) Config {
				c.TokensData[1337] = TokenData{
					NordLynxPrivateKey: "nobody-is-going-to-guess-this",
				}
				return c
			},
		},
		{
			name: "allowlist is saved",
			f: func(c Config) Config {
				c.AutoConnectData.Allowlist.Ports.TCP = map[int64]bool{
					443: true,
				}
				c.AutoConnectData.Allowlist.Ports.UDP = map[int64]bool{
					53: true,
				}
				c.AutoConnectData.Allowlist.Subnets = []string{
					"1.1.1.1/32",
				}
				return c
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configLocation := "testdata/config"
			vaultLocation := "testdata/vault"
			fs := NewFilesystemConfigManager("testdata/config", "testdata/vault", "", NewMachineID(os.ReadFile, os.Hostname), StdFilesystemHandle{}, nil)
			defer os.Remove(configLocation)
			defer os.Remove(vaultLocation)

			err := fs.SaveWith(test.f)
			assert.NoError(t, err)

			var cfg Config
			err = fs.Load(&cfg)
			assert.NoError(t, err)
			var cfg2 Config
			assert.NotEqual(t, cfg2, cfg)
		})
	}
}

// Deprecated: since 3.10.0
func TestConfigIsBackwardsCompatible(t *testing.T) {
	category.Set(t, category.File)

	salt, ok := os.LookupEnv("SALT")
	require.True(t, ok)

	tests := []struct {
		settingsFile string
		installFile  string
	}{
		{
			settingsFile: "testdata/settings_3.10.0.dat",
			installFile:  "testdata/install_3.10.0.dat",
		},
		{
			settingsFile: "testdata/settings_3.12.0.dat",
			installFile:  "testdata/install_3.12.0.dat",
		},
		{
			settingsFile: "testdata/settings_3.13.0.dat",
			installFile:  "testdata/install_3.13.0.dat",
		},
		{
			settingsFile: "testdata/settings_3.14.0.dat",
			installFile:  "testdata/install_3.14.0.dat",
		},
	}

	for _, test := range tests {
		t.Run(test.settingsFile, func(t *testing.T) {
			fs := NewFilesystemConfigManager(test.settingsFile, test.installFile, salt, NewMachineID(os.ReadFile, os.Hostname), StdFilesystemHandle{}, nil)
			var cfg Config
			err := fs.Load(&cfg)
			require.NoError(t, err)
			var cfg2 Config
			assert.NotEqual(t, cfg2, cfg)
		})
	}
}

func TestConfigDefaultValues(t *testing.T) {
	category.Set(t, category.File)

	salt, ok := os.LookupEnv("SALT")
	require.True(t, ok)

	tests := []struct {
		settingsFile string
		installFile  string
		autoconnect  bool
		technology   Technology
	}{
		{
			settingsFile: "testdata/settings_3.10.0.dat",
			installFile:  "testdata/install_3.10.0.dat",
			autoconnect:  true,
			technology:   Technology_NORDLYNX,
		},
		{
			settingsFile: "testdata/settings_3.12.0.dat",
			installFile:  "testdata/install_3.12.0.dat",
			technology:   Technology_OPENVPN,
		},
		{
			settingsFile: "testdata/settings_3.13.0.dat",
			installFile:  "testdata/install_3.13.0.dat",
			technology:   Technology_NORDLYNX,
		},
		{
			settingsFile: "testdata/settings_3.14.0.dat",
			installFile:  "testdata/install_3.14.0.dat",
			technology:   Technology_NORDLYNX,
		},
	}

	for _, test := range tests {
		t.Run(test.settingsFile, func(t *testing.T) {
			fs := NewFilesystemConfigManager(test.settingsFile, test.installFile, salt, NewMachineID(os.ReadFile, os.Hostname), StdFilesystemHandle{}, nil)
			var cfg Config
			err := fs.Load(&cfg)
			require.NoError(t, err)
			assert.Equal(t, defaultFWMarkValue, cfg.FirewallMark)
			assert.Equal(t, test.technology, cfg.Technology)
			assert.True(t, cfg.Firewall)
			assert.True(t, cfg.Routing.Get())
			assert.False(t, cfg.Mesh)
			assert.False(t, cfg.KillSwitch)
			assert.Equal(t, test.autoconnect, cfg.AutoConnect)
			assert.False(t, cfg.IPv6)
			assert.True(t, cfg.VirtualLocation.Get())
		})
	}
}
