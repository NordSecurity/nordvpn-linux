package mock

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
)

type ConfigManager struct {
	mu      sync.RWMutex
	Cfg     *config.Config
	SaveErr error
	LoadErr error
	Saved   bool
}

func NewMockConfigManager() *ConfigManager {
	m := ConfigManager{}
	m.Cfg = &config.Config{}
	m.Cfg.MeshDevice = &mesh.Machine{}

	return &m
}

// deepCopyConfig creates a complete deep copy of the config to prevent race conditions
// when accessing maps and pointers.
func deepCopyConfig(src config.Config) config.Config {
	// Shallow copy all value fields
	dst := src

	// Deep copy TokensData map to prevent concurrent map access
	if src.TokensData != nil {
		dst.TokensData = make(map[int64]config.TokenData, len(src.TokensData))
		for userID, tokenData := range src.TokensData {
			tokenCopy := tokenData
			if tokenData.IdempotencyKey != nil {
				keyCopy := *tokenData.IdempotencyKey
				tokenCopy.IdempotencyKey = &keyCopy
			}
			dst.TokensData[userID] = tokenCopy
		}
	}

	// Deep copy UsersData to prevent concurrent access to nested maps
	if src.UsersData != nil {
		usersCopy := *src.UsersData
		if src.UsersData.NotifyOff != nil {
			usersCopy.NotifyOff = make(config.UidBoolMap, len(src.UsersData.NotifyOff))
			for uid, value := range src.UsersData.NotifyOff {
				usersCopy.NotifyOff[uid] = value
			}
		}

		if src.UsersData.TrayOff != nil {
			usersCopy.TrayOff = make(config.UidBoolMap, len(src.UsersData.TrayOff))
			for uid, value := range src.UsersData.TrayOff {
				usersCopy.TrayOff[uid] = value
			}
		}

		dst.UsersData = &usersCopy
	}

	// Deep copy MeshDevice pointer to prevent concurrent access
	if src.MeshDevice != nil {
		meshCopy := *src.MeshDevice
		dst.MeshDevice = &meshCopy
	}

	return dst
}

func (m *ConfigManager) SaveWith(fn config.SaveFunc) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SaveErr != nil {
		return m.SaveErr
	}

	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}

	copyCfg := deepCopyConfig(*m.Cfg)
	updatedCfg := fn(copyCfg)
	*m.Cfg = updatedCfg

	m.Saved = true
	return nil
}

func (m *ConfigManager) Load(c *config.Config) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.LoadErr != nil {
		return m.LoadErr
	}

	if m.Cfg == nil {
		m.Cfg = &config.Config{}
	}

	*c = deepCopyConfig(*m.Cfg)
	return nil
}

func (m *ConfigManager) Reset(bool, bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	*m.Cfg = config.Config{}
	return nil
}
