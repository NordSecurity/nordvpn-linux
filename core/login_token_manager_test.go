package core

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigManager struct {
	c config.Config
}

func (m *mockConfigManager) SaveWith(fn config.SaveFunc) error {
	m.c = fn(m.c)
	return nil
}

func (m *mockConfigManager) Load(cfg *config.Config) error {
	*cfg = m.c
	return nil
}

func (m *mockConfigManager) Reset(preserveLoginData bool, disableKillswitch bool) error {
	if preserveLoginData {
		newConfig := &config.Config{
			AutoConnectData: config.AutoConnectData{
				ID: m.c.AutoConnectData.ID,
			},
			TokensData: m.c.TokensData,
		}
		m.c = *newConfig
	}

	m.c = config.Config{
		TokensData: make(map[int64]config.TokenData),
	}

	return nil
}

func NewMockConfigManager() config.Manager {
	mgr := &mockConfigManager{}
	mgr.Reset(false, false)
	return mgr
}

func mockTokenRenewAPICall(token string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error) {
	return nil, nil
}

type mockTokenValidator struct {
	TokenValidator
	ValidatorFunc validatorFunc
}

type validatorFunc = func(token string, expiryDate string) error

func (m *mockTokenValidator) Validate(token string, expiryDate string) error {
	return m.ValidatorFunc(token, expiryDate)
}

func Test_LoginTokenManager_Token(t *testing.T) {
	category.Set(t, category.Unit)

	var mockErrRegsitry ErrorHandlingRegistry[int64]
	mockedCfgManager := NewMockConfigManager()
	tokenman := NewLoginTokenManager(
		mockedCfgManager,
		mockTokenRenewAPICall,
		&mockErrRegsitry,
		&mockTokenValidator{},
	)

	token, err := tokenman.Token()
	assert.Empty(t, token)
	assert.NotNil(t, err)

	mockedCfgManager.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.ID = 1
		tokenCfg := c.TokensData[c.AutoConnectData.ID]
		tokenCfg.Token = "nice-token"
		c.TokensData[c.AutoConnectData.ID] = tokenCfg
		return c
	})

	token, err = tokenman.Token()
	assert.Equal(t, "nice-token", token)
	assert.Nil(t, err)
}

func TestLoginTokenManager_Renew_NotExpiredTokenNotRenewed(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	mockCfg := NewMockConfigManager().(*mockConfigManager)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				RenewToken:     "current-renew-token",
				TokenExpiry:    now.Add(1 * time.Hour).Format(time.RFC3339),
				IdempotencyKey: nil,
			},
		},
	}

	mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
		return &TokenRenewResponse{
			Token:      "new-access-token",
			RenewToken: "new-renew-token",
			ExpiresAt:  now.Add(2 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()

	validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
		return nil
	}}
	manager := NewLoginTokenManager(
		mockCfg,
		mockRenewAPICall,
		mockErrHandlerRegistry,
		validator,
	)

	err := manager.Renew()
	require.NoError(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "current-access-token", updated.Token)
	assert.Equal(t, "current-renew-token", updated.RenewToken)
	assert.Nil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

func TestLoginTokenManager_Renew_ExpiredTokenGetsRenewedWithoutIdempotencyKey(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	mockCfg := NewMockConfigManager().(*mockConfigManager)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				TokenExpiry:    now.Add(-1 * time.Hour).Format(time.RFC3339),
				RenewToken:     "current-renew-token",
				IdempotencyKey: nil,
			},
		},
	}

	mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
		return &TokenRenewResponse{
			Token:      "new-access-token",
			RenewToken: "new-renew-token",
			ExpiresAt:  now.Add(2 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()

	validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
		return errors.New("expired")
	}}
	manager := NewLoginTokenManager(
		mockCfg,
		mockRenewAPICall,
		mockErrHandlerRegistry,
		validator,
	)

	err := manager.Renew()
	require.NoError(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "new-access-token", updated.Token)
	assert.Equal(t, "new-renew-token", updated.RenewToken)
	assert.NotNil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

func TestLoginTokenManager_Renew_ExpiredTokenGetsRenewedWithIdempotencyKey(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	newIdempotencyKey := func() *uuid.UUID { key := uuid.New(); return &key }

	mockCfg := NewMockConfigManager().(*mockConfigManager)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				TokenExpiry:    now.Add(-1 * time.Hour).Format(time.RFC3339),
				RenewToken:     "current-renew-token",
				IdempotencyKey: newIdempotencyKey(),
			},
		},
	}

	mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
		return &TokenRenewResponse{
			Token:      "new-access-token",
			RenewToken: "new-renew-token",
			ExpiresAt:  now.Add(2 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()

	validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
		return errors.New("expired")
	}}
	manager := NewLoginTokenManager(
		mockCfg,
		mockRenewAPICall,
		mockErrHandlerRegistry,
		validator,
	)

	err := manager.Renew()
	require.NoError(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "new-access-token", updated.Token)
	assert.Equal(t, "new-renew-token", updated.RenewToken)
	assert.NotNil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

func TestLoginTokenManager_Renew_ExpiredTokenNotRenewedWithUnknownUnhandledAPICallError(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	mockCfg := NewMockConfigManager().(*mockConfigManager)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-token",
				RenewToken:     "current-renew-token",
				TokenExpiry:    now.Add(-1 * time.Hour).Format(time.RFC3339),
				IdempotencyKey: nil,
			},
		},
	}

	mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
		return nil, errors.New("api call error")
	}

	mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()

	validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
		return errors.New("expired")
	}}
	manager := NewLoginTokenManager(
		mockCfg,
		mockRenewAPICall,
		mockErrHandlerRegistry,
		validator,
	)

	copyToken := strings.Clone(mockCfg.c.TokensData[uid].Token)
	copyRenewToken := strings.Clone(mockCfg.c.TokensData[uid].RenewToken)
	copyTokenExpiry := strings.Clone(mockCfg.c.TokensData[uid].TokenExpiry)

	err := manager.Renew()
	assert.NoError(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, copyToken, updated.Token)
	assert.Equal(t, copyRenewToken, updated.RenewToken)
	assert.NotNil(t, updated.IdempotencyKey)
	assert.Equal(t, copyTokenExpiry, updated.TokenExpiry)
}

func TestLoginTokenManager_Renew_ExpiredTokenGetsRenewedWithKnownAPICallError(t *testing.T) {
	category.Set(t, category.Unit)

	errs := []error{ErrUnauthorized, ErrNotFound, ErrBadRequest}

	for _, errIterator := range errs {
		now := time.Now()
		uid := int64(42)

		mockCfg := NewMockConfigManager().(*mockConfigManager)
		mockCfg.c = config.Config{
			TokensData: map[int64]config.TokenData{
				uid: {
					Token:          "current-token",
					RenewToken:     "current-renew-token",
					TokenExpiry:    now.Add(-1 * time.Hour).Format(time.RFC3339),
					IdempotencyKey: nil,
				},
			},
		}

		mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
			return nil, errIterator
		}

		mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()
		validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
			return errors.New("expired")
		}}
		manager := NewLoginTokenManager(
			mockCfg,
			mockRenewAPICall,
			mockErrHandlerRegistry,
			validator,
		)

		err := manager.Renew()
		assert.NoError(t, err)

		updated := mockCfg.c.TokensData[uid]
		assert.Empty(t, updated.Token)
		assert.Empty(t, updated.RenewToken)
		assert.Nil(t, updated.IdempotencyKey)
		assert.Empty(t, updated.TokenExpiry)
	}
}

func Test_LoginTokenManager_Store(t *testing.T) {
	category.Set(t, category.Unit)

	var mockErrRegsitry ErrorHandlingRegistry[int64]
	mockedCfgManager := NewMockConfigManager()
	tokenman := NewLoginTokenManager(
		mockedCfgManager,
		mockTokenRenewAPICall,
		&mockErrRegsitry,
		&mockTokenValidator{},
	)

	err := tokenman.Store("nice-token")
	assert.NotNil(t, err)

	const uid = 1
	mockedCfgManager.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.ID = uid
		tokenCfg := c.TokensData[c.AutoConnectData.ID]
		tokenCfg.Token = "random-stuff"
		c.TokensData[c.AutoConnectData.ID] = tokenCfg
		return c
	})

	err = tokenman.Store("nice-token-again")
	assert.Nil(t, err)

	var cfg config.Config
	err = mockedCfgManager.Load(&cfg)
	assert.Nil(t, err, "internal mocking error")
	assert.Equal(t, "nice-token-again", cfg.TokensData[uid].Token)
}

func TestLoginTokenManager_Invalidate(t *testing.T) {
	category.Set(t, category.Unit)

	mockErrRegsitry := NewErrorHandlingRegistry[int64]()
	mockedCfgManager := NewMockConfigManager()
	tokenman := NewLoginTokenManager(
		mockedCfgManager,
		mockTokenRenewAPICall,
		mockErrRegsitry,
		&mockTokenValidator{},
	)

	errDummy := errors.New("dummy")
	err := tokenman.Invalidate(errDummy)
	assert.Nil(t, err)

	dummyErrHandlerCalledCnt := 0
	dummyErrHandler := func(int64) {
		dummyErrHandlerCalledCnt++
	}

	const dummyErrHandlerAddCnt = 4
	for i := 0; i < dummyErrHandlerAddCnt; i++ {
		mockErrRegsitry.Add(dummyErrHandler, errDummy)
	}

	err = tokenman.Invalidate(errDummy)
	assert.Nil(t, err)
	assert.NotEqual(t, dummyErrHandlerCalledCnt, dummyErrHandlerAddCnt)

	const uid = 1
	mockedCfgManager.SaveWith(func(c config.Config) config.Config {
		c.AutoConnectData.ID = uid
		tokenCfg := c.TokensData[c.AutoConnectData.ID]
		tokenCfg.Token = "random-stuff"
		c.TokensData[c.AutoConnectData.ID] = tokenCfg
		return c
	})

	err = tokenman.Invalidate(errDummy)
	assert.Nil(t, err)
	assert.Equal(t, dummyErrHandlerAddCnt, dummyErrHandlerCalledCnt)

	testErr := errors.New("fancy error we have here")
	externalErrHandled := false
	externalErrUserMatching := false
	mockUid := int64(1)
	mockErrRegsitry.Add(func(int64) {
		if uid == mockUid {
			externalErrUserMatching = true
		}
		externalErrHandled = true
	}, testErr)

	err = tokenman.Invalidate(testErr)
	assert.Nil(t, err)
	assert.True(t, externalErrHandled)
	assert.True(t, externalErrUserMatching)
}

type mockConfigManagerWithBadSave struct {
	c config.Config
}

func (m *mockConfigManagerWithBadSave) SaveWith(fn config.SaveFunc) error {
	return errors.New("successfully failed")
}

func (m *mockConfigManagerWithBadSave) Load(cfg *config.Config) error {
	*cfg = m.c
	return nil
}

func (m *mockConfigManagerWithBadSave) Reset(preserveLoginData bool, disableKillswitch bool) error {
	if preserveLoginData {
		newConfig := &config.Config{
			AutoConnectData: config.AutoConnectData{
				ID: m.c.AutoConnectData.ID,
			},
			TokensData: m.c.TokensData,
		}
		m.c = *newConfig
	}

	m.c = config.Config{
		TokensData: make(map[int64]config.TokenData),
	}

	return nil
}

func NewMockConfigManagerWithBadSave() config.Manager {
	mgr := &mockConfigManagerWithBadSave{}
	mgr.Reset(false, false)
	return mgr
}

func TestLoginTokenManager_Renew_ExpiredTokenGetsRenewedWithBadConfigSaving(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	mockCfg := NewMockConfigManagerWithBadSave().(*mockConfigManagerWithBadSave)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				TokenExpiry:    now.Add(-1 * time.Hour).Format(time.RFC3339),
				RenewToken:     "current-renew-token",
				IdempotencyKey: nil,
			},
		},
	}

	mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
		return &TokenRenewResponse{
			Token:      "new-access-token",
			RenewToken: "new-renew-token",
			ExpiresAt:  now.Add(2 * time.Hour).Format(time.RFC3339),
		}, nil
	}

	mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()
	validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
		return errors.New("expired")
	}}
	manager := NewLoginTokenManager(
		mockCfg,
		mockRenewAPICall,
		mockErrHandlerRegistry,
		validator,
	)

	err := manager.Renew()
	require.Error(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "current-access-token", updated.Token)
	assert.Equal(t, "current-renew-token", updated.RenewToken)
	assert.Nil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

func TestLoginTokenManager_Renew_ExpiredTokenGetsRenewedWithIdempotencyKeyANdBadConfigSaving(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	newIdempotencyKey := func() *uuid.UUID { key := uuid.New(); return &key }

	mockCfg := NewMockConfigManagerWithBadSave().(*mockConfigManagerWithBadSave)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				TokenExpiry:    now.Add(1 * time.Hour).Format(time.RFC3339),
				RenewToken:     "current-renew-token",
				IdempotencyKey: newIdempotencyKey(),
			},
		},
	}

	mockRenewAPICall := func(token string, idKey uuid.UUID) (*TokenRenewResponse, error) {
		return nil, ErrUnauthorized
	}

	mockErrHandlerRegistry := NewErrorHandlingRegistry[int64]()
	validator := &mockTokenValidator{ValidatorFunc: func(token string, expiryDate string) error {
		return errors.New("expired")
	}}
	manager := NewLoginTokenManager(
		mockCfg,
		mockRenewAPICall,
		mockErrHandlerRegistry,
		validator,
	)

	err := manager.Renew()
	require.Error(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "current-access-token", updated.Token)
	assert.Equal(t, "current-renew-token", updated.RenewToken)
	assert.NotNil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

type mockConfigManagerWithBadLoad struct {
	c config.Config
}

func (m *mockConfigManagerWithBadLoad) SaveWith(fn config.SaveFunc) error {
	m.c = fn(m.c)
	return nil
}

func (m *mockConfigManagerWithBadLoad) Load(cfg *config.Config) error {
	return errors.New("successfully failed")
}

func (m *mockConfigManagerWithBadLoad) Reset(preserveLoginData bool, disableKillswitch bool) error {
	if preserveLoginData {
		newConfig := &config.Config{
			AutoConnectData: config.AutoConnectData{
				ID: m.c.AutoConnectData.ID,
			},
			TokensData: m.c.TokensData,
		}
		m.c = *newConfig
	}

	m.c = config.Config{
		TokensData: make(map[int64]config.TokenData),
	}

	return nil
}

func NewMockConfigManagerWithBadLoad() config.Manager {
	mgr := &mockConfigManagerWithBadLoad{}
	mgr.Reset(false, false)
	return mgr
}

func Test_LoginTokenManager_TokenWithBadLoad(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	var mockErrRegsitry ErrorHandlingRegistry[int64]
	mockCfg := NewMockConfigManagerWithBadLoad().(*mockConfigManagerWithBadLoad)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				RenewToken:     "current-renew-token",
				TokenExpiry:    now.Add(1 * time.Hour).Format(time.RFC3339),
				IdempotencyKey: nil,
			},
		},
	}

	tokenman := NewLoginTokenManager(
		mockCfg,
		mockTokenRenewAPICall,
		&mockErrRegsitry,
		&mockTokenValidator{},
	)

	token, err := tokenman.Token()
	updated := mockCfg.c.TokensData[uid]
	assert.Empty(t, token)
	assert.NotEqual(t, updated.Token, token)
	assert.Error(t, err)
}

func Test_LoginTokenManager_RenewWithBadLoad(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	var mockErrRegsitry ErrorHandlingRegistry[int64]

	mockCfg := NewMockConfigManagerWithBadLoad().(*mockConfigManagerWithBadLoad)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				TokenExpiry:    now.Add(1 * time.Hour).Format(time.RFC3339),
				RenewToken:     "current-renew-token",
				IdempotencyKey: nil,
			},
		},
	}

	tokenman := NewLoginTokenManager(
		mockCfg,
		mockTokenRenewAPICall,
		&mockErrRegsitry,
		&mockTokenValidator{},
	)

	err := tokenman.Renew()
	assert.Error(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "current-access-token", updated.Token)
	assert.Equal(t, "current-renew-token", updated.RenewToken)
	assert.Nil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

func Test_LoginTokenManager_StoreWithBadLoad(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	var mockErrRegsitry ErrorHandlingRegistry[int64]
	mockCfg := NewMockConfigManagerWithBadLoad().(*mockConfigManagerWithBadLoad)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				RenewToken:     "current-renew-token",
				TokenExpiry:    now.Add(1 * time.Hour).Format(time.RFC3339),
				IdempotencyKey: nil,
			},
		},
	}

	tokenman := NewLoginTokenManager(
		mockCfg,
		mockTokenRenewAPICall,
		&mockErrRegsitry,
		&mockTokenValidator{},
	)

	err := tokenman.Store("some token")
	assert.Error(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "current-access-token", updated.Token)
	assert.Equal(t, "current-renew-token", updated.RenewToken)
	assert.Nil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}

func Test_LoginTokenManager_InvalidateWithBadLoad(t *testing.T) {
	category.Set(t, category.Unit)

	now := time.Now()
	uid := int64(42)

	var mockErrRegsitry ErrorHandlingRegistry[int64]
	mockCfg := NewMockConfigManagerWithBadLoad().(*mockConfigManagerWithBadLoad)
	mockCfg.c = config.Config{
		TokensData: map[int64]config.TokenData{
			uid: {
				Token:          "current-access-token",
				RenewToken:     "current-renew-token",
				TokenExpiry:    now.Add(1 * time.Hour).Format(time.RFC3339),
				IdempotencyKey: nil,
			},
		},
	}

	tokenman := NewLoginTokenManager(
		mockCfg,
		mockTokenRenewAPICall,
		&mockErrRegsitry,
		&mockTokenValidator{},
	)

	err := tokenman.Invalidate(errors.New("some error"))
	assert.Error(t, err)

	updated := mockCfg.c.TokensData[uid]
	assert.Equal(t, "current-access-token", updated.Token)
	assert.Equal(t, "current-renew-token", updated.RenewToken)
	assert.Nil(t, updated.IdempotencyKey)
	assert.NotEmpty(t, updated.TokenExpiry)
}
