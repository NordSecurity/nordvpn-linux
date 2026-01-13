package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
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
			tmpDir := t.TempDir()
			configLocation := filepath.Join(tmpDir, "config")
			vaultLocation := filepath.Join(tmpDir, "vault")
			fs := NewFilesystemConfigManager(configLocation, vaultLocation, "", NewMachineID(os.ReadFile, os.Hostname), StdFilesystemHandle{}, nil)

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
	assert.True(t, ok)

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
			assert.NoError(t, err)
			var cfg2 Config
			assert.NotEqual(t, cfg2, cfg)
		})
	}
}

func TestConfigDefaultValues(t *testing.T) {
	category.Set(t, category.File)

	salt, ok := os.LookupEnv("SALT")
	assert.True(t, ok)

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
			assert.NoError(t, err)
			assert.Equal(t, defaultFWMarkValue, cfg.FirewallMark)
			assert.Equal(t, test.technology, cfg.Technology)
			assert.True(t, cfg.Firewall)
			assert.True(t, cfg.Routing.Get())
			assert.False(t, cfg.Mesh)
			assert.False(t, cfg.KillSwitch)
			assert.Equal(t, test.autoconnect, cfg.AutoConnect)
			assert.True(t, cfg.VirtualLocation.Get())
		})
	}
}

func TestSaveWithPublishesPreviousConfig(t *testing.T) {
	category.Set(t, category.File)

	tmpDir := t.TempDir()
	configLocation := filepath.Join(tmpDir, "config")
	vaultLocation := filepath.Join(tmpDir, "vault")

	var publishedChange DataConfigChange
	publisher := &mockConfigPublisher{
		onPublish: func(change DataConfigChange) {
			publishedChange = change
		},
	}

	fs := NewFilesystemConfigManager(
		configLocation,
		vaultLocation,
		"",
		NewMachineID(os.ReadFile, os.Hostname),
		StdFilesystemHandle{},
		publisher,
	)

	// First save - creates initial config
	err := fs.SaveWith(func(c Config) Config {
		c.AutoConnect = true
		return c
	})
	assert.NoError(t, err)

	// Verify first publish (new installation)
	assert.Nil(t, publishedChange.PreviousConfig)
	assert.NotNil(t, publishedChange.Config)
	assert.True(t, publishedChange.Config.AutoConnect)

	// Second save - should have PreviousConfig
	err = fs.SaveWith(func(c Config) Config {
		c.KillSwitch = true
		return c
	})
	assert.NoError(t, err)

	// Verify PreviousConfig contains the state before the change
	assert.NotNil(t, publishedChange.PreviousConfig)
	assert.True(t, publishedChange.PreviousConfig.AutoConnect, "PreviousConfig should have AutoConnect=true from first save")
	assert.False(t, publishedChange.PreviousConfig.KillSwitch, "PreviousConfig should have KillSwitch=false before change")

	// Verify Config contains the new state
	assert.NotNil(t, publishedChange.Config)
	assert.True(t, publishedChange.Config.AutoConnect)
	assert.True(t, publishedChange.Config.KillSwitch, "Config should have KillSwitch=true after change")
}

type mockConfigPublisher struct {
	onPublish func(DataConfigChange)
}

func (m *mockConfigPublisher) Publish(change DataConfigChange) {
	if m.onPublish != nil {
		m.onPublish(change)
	}
}

func TestLoadTwoCopiesReturnsIndependentCopies(t *testing.T) {
	category.Set(t, category.File)

	tmpDir := t.TempDir()
	configLocation := filepath.Join(tmpDir, "config")
	vaultLocation := filepath.Join(tmpDir, "vault")

	fs := NewFilesystemConfigManager(
		configLocation,
		vaultLocation,
		"",
		NewMachineID(os.ReadFile, os.Hostname),
		StdFilesystemHandle{},
		nil,
	)

	// Create initial config
	err := fs.SaveWith(func(c Config) Config {
		c.AutoConnect = true
		c.TokensData[123] = TokenData{Token: "test-token"}
		return c
	})
	assert.NoError(t, err)

	// Load two copies
	var first, second Config
	_, err = fs.loadTwoCopies(&first, &second)
	assert.NoError(t, err)

	// Verify both have the same initial values
	assert.Equal(t, first.AutoConnect, second.AutoConnect)
	assert.Equal(t, first.TokensData[123].Token, second.TokensData[123].Token)

	// Modify first copy
	first.AutoConnect = false
	first.TokensData[123] = TokenData{Token: "modified-token"}

	// Verify second copy is unchanged (independent)
	assert.NotEqual(t, first, second, "Modifications to the first copy should not affect the second copy")
}
