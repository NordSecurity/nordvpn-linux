package session

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/stretchr/testify/assert"
)

type mockConfigManager struct {
	config    config.Config
	loadError error
	saveError error
}

func (m *mockConfigManager) Load(cfg *config.Config) error {
	if m.loadError != nil {
		return m.loadError
	}
	*cfg = m.config
	return nil
}

func (m *mockConfigManager) SaveWith(f config.SaveFunc) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.config = f(m.config)
	return nil
}

func (m *mockConfigManager) Reset(preserveLoginData bool, disableKillswitch bool) error {
	return nil
}

func TestAccessTokenSession_Get(t *testing.T) {
	expiryTime := time.Now().Add(time.Hour)
	expiryTimeStr := expiryTime.Format(internal.ServerDateFormat)

	t.Run("success", func(t *testing.T) {
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						Token:       "test-token",
						RenewToken:  "test-renew-token",
						TokenExpiry: expiryTimeStr,
					},
				},
			},
		}

		session := newAccessTokenSession(cm)
		cfg, err := session.get()

		assert.NoError(t, err)
		assert.Equal(t, "test-token", cfg.Token)
		assert.Equal(t, "test-renew-token", cfg.RenewToken)
		parsed, _ := time.Parse(internal.ServerDateFormat, expiryTimeStr)
		assert.Equal(t, parsed, cfg.ExpiresAt)
	})

	t.Run("load error", func(t *testing.T) {
		cm := &mockConfigManager{
			loadError: errors.New("load error"),
		}

		session := newAccessTokenSession(cm)
		_, err := session.get()

		assert.Error(t, err)
		assert.Equal(t, "load error", err.Error())
	})

	t.Run("non existing data", func(t *testing.T) {
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData:      map[int64]config.TokenData{},
			},
		}

		session := newAccessTokenSession(cm)
		_, err := session.get()

		assert.Error(t, err)
		assert.Equal(t, "non existing data", err.Error())
	})
}

func TestAccessTokenSession_Set(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expiryTime := time.Now().Add(time.Hour)
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						Token:       "old-token",
						RenewToken:  "old-renew-token",
						TokenExpiry: "old-expiry",
					},
				},
			},
		}

		session := newAccessTokenSession(cm)
		err := session.set(accessTokenConfig{
			Token:      "new-token",
			RenewToken: "new-renew-token",
			ExpiresAt:  expiryTime,
		})

		assert.NoError(t, err)
		assert.Equal(t, "new-token", cm.config.TokensData[123].Token)
		assert.Equal(t, "new-renew-token", cm.config.TokensData[123].RenewToken)
		assert.Equal(t, expiryTime.Format(internal.ServerDateFormat), cm.config.TokensData[123].TokenExpiry)
	})

	t.Run("save error", func(t *testing.T) {
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {},
				},
			},
			saveError: errors.New("save error"),
		}

		session := newAccessTokenSession(cm)
		err := session.set(accessTokenConfig{})

		assert.Error(t, err)
		assert.Equal(t, "save error", err.Error())
	})
}

func TestAccessTokenSessionStore_SetToken(t *testing.T) {
	cm := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData: map[int64]config.TokenData{
				123: {
					Token:       "old-token",
					RenewToken:  "renew-token",
					TokenExpiry: "expiry",
				},
			},
		},
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetToken("new-token")

	assert.NoError(t, err)
	assert.Equal(t, "new-token", cm.config.TokensData[123].Token)
}

func TestAccessTokenSessionStore_SetRenewToken(t *testing.T) {
	cm := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData: map[int64]config.TokenData{
				123: {
					Token:       "token",
					RenewToken:  "old-renew-token",
					TokenExpiry: "expiry",
				},
			},
		},
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetRenewToken("new-renew-token")

	assert.NoError(t, err)
	assert.Equal(t, "new-renew-token", cm.config.TokensData[123].RenewToken)
}

func TestAccessTokenSessionStore_SetExpiry(t *testing.T) {
	cm := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData: map[int64]config.TokenData{
				123: {
					Token:       "token",
					RenewToken:  "renew-token",
					TokenExpiry: "old-expiry",
				},
			},
		},
	}

	newExpiry := time.Now().Add(time.Hour)
	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetExpiry(newExpiry)

	assert.NoError(t, err)
	assert.Equal(t, newExpiry.Format(internal.ServerDateFormat), cm.config.TokensData[123].TokenExpiry)
}

func TestAccessTokenSessionStore_GetToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						Token: "test-token",
					},
				},
			},
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		token := store.GetToken()

		assert.Equal(t, "test-token", token)
	})

	t.Run("error", func(t *testing.T) {
		cm := &mockConfigManager{
			loadError: errors.New("load error"),
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		token := store.GetToken()

		assert.Equal(t, "", token)
	})
}

func TestAccessTokenSessionStore_GetRenewalToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						RenewToken: "test-renew-token",
					},
				},
			},
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		token := store.GetRenewalToken()

		assert.Equal(t, "test-renew-token", token)
	})

	t.Run("error", func(t *testing.T) {
		cm := &mockConfigManager{
			loadError: errors.New("load error"),
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		token := store.GetRenewalToken()

		assert.Equal(t, "", token)
	})
}

func TestAccessTokenSessionStore_GetExpiry(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expiryTime := time.Now().Add(time.Hour)
		expiryTimeStr := expiryTime.Format(internal.ServerDateFormat)
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						TokenExpiry: expiryTimeStr,
					},
				},
			},
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		expiry := store.GetExpiry()

		parsed, _ := time.Parse(internal.ServerDateFormat, expiryTimeStr)
		assert.Equal(t, parsed, expiry)
	})

	t.Run("error", func(t *testing.T) {
		cm := &mockConfigManager{
			loadError: errors.New("load error"),
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		expiry := store.GetExpiry()

		assert.Equal(t, time.Time{}, expiry)
	})
}

func TestAccessTokenSessionStore_IsExpired(t *testing.T) {
	t.Run("not expired", func(t *testing.T) {
		expiryTime := time.Now().Add(time.Hour)
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						TokenExpiry: expiryTime.Format(internal.ServerDateFormat),
					},
				},
			},
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		expired := store.IsExpired()

		assert.False(t, expired)
	})

	t.Run("expired", func(t *testing.T) {
		expiryTime := time.Now().Add(-12 * time.Hour) // Past time
		cm := &mockConfigManager{
			config: config.Config{
				AutoConnectData: config.AutoConnectData{ID: 123},
				TokensData: map[int64]config.TokenData{
					123: {
						TokenExpiry: expiryTime.Format(internal.ServerDateFormat),
					},
				},
			},
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		expired := store.IsExpired()

		assert.True(t, expired)
	})

	t.Run("error", func(t *testing.T) {
		cm := &mockConfigManager{
			loadError: errors.New("load error"),
		}

		store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
		expired := store.IsExpired()

		assert.True(t, expired)
	})
}

func TestAccessTokenSessionStore_SetToken_GetError(t *testing.T) {
	cm := &mockConfigManager{
		loadError: errors.New("load error"),
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetToken("new-token")

	assert.Error(t, err)
	assert.Equal(t, "load error", err.Error())
}

func TestAccessTokenSessionStore_SetRenewToken_GetError(t *testing.T) {
	cm := &mockConfigManager{
		loadError: errors.New("load error"),
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetRenewToken("new-renew-token")

	assert.Error(t, err)
	assert.Equal(t, "load error", err.Error())
}

func TestAccessTokenSessionStore_SetExpiry_GetError(t *testing.T) {
	cm := &mockConfigManager{
		loadError: errors.New("load error"),
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetExpiry(time.Now())

	assert.Error(t, err)
	assert.Equal(t, "load error", err.Error())
}

func TestAccessTokenSessionStore_SetToken_NonExistingData(t *testing.T) {
	cm := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData:      map[int64]config.TokenData{},
		},
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetToken("new-token")

	assert.Error(t, err)
	assert.Equal(t, "non existing data", err.Error())
}

func TestAccessTokenSessionStore_SetRenewToken_NonExistingData(t *testing.T) {
	cm := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData:      map[int64]config.TokenData{},
		},
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetRenewToken("new-renew-token")

	assert.Error(t, err)
	assert.Equal(t, "non existing data", err.Error())
}

func TestAccessTokenSessionStore_SetExpiry_NonExistingData(t *testing.T) {
	cm := &mockConfigManager{
		config: config.Config{
			AutoConnectData: config.AutoConnectData{ID: 123},
			TokensData:      map[int64]config.TokenData{},
		},
	}

	store := AccessTokenSessionStore{session: newAccessTokenSession(cm)}
	err := store.SetExpiry(time.Now())

	assert.Error(t, err)
	assert.Equal(t, "non existing data", err.Error())
}
