package session_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// raceConditionRenewalAPI simulates server behavior where using an old renewToken returns 404.
type raceConditionRenewalAPI struct {
	mu                sync.Mutex
	currentRenewToken string
	requestDelay      time.Duration
	callCount         atomic.Int32
	successCount      atomic.Int32
	failCount         atomic.Int32
}

func newRaceConditionRenewalAPI(initialRenewToken string) *raceConditionRenewalAPI {
	return &raceConditionRenewalAPI{
		currentRenewToken: initialRenewToken,
		requestDelay:      50 * time.Millisecond,
	}
}

var ErrRenewTokenNotFound = internal.NewCodedError(101202, "Renew token not found", core.ErrNotFound)

func (api *raceConditionRenewalAPI) Renew(token string, idempotencyKey uuid.UUID) (*session.AccessTokenResponse, error) {
	api.callCount.Add(1)
	time.Sleep(api.requestDelay)

	api.mu.Lock()
	defer api.mu.Unlock()

	if token != api.currentRenewToken {
		api.failCount.Add(1)
		return nil, ErrRenewTokenNotFound
	}

	api.successCount.Add(1)
	u := uuid.New()
	newRenewToken := fmt.Sprintf("%x%x", u[:], u[:4])[:40]
	api.currentRenewToken = newRenewToken

	return &session.AccessTokenResponse{
		Token:      "ab78bb36299d442fa0715fb53b5e3e57",
		RenewToken: newRenewToken,
		ExpiresAt:  time.Now().UTC().Add(24 * time.Hour).Format(internal.ServerDateFormat),
	}, nil
}

// TestAccessTokenSessionStore_RaceCondition_MutexFix verifies mutex prevents
// concurrent renewals from causing spurious logouts.
func TestAccessTokenSessionStore_RaceCondition_MutexFix(t *testing.T) {
	category.Set(t, category.Unit)

	initialRenewToken := "deadbeef1234567890abcdef1234567890abcdef"
	api := newRaceConditionRenewalAPI(initialRenewToken)
	idempotencyKey := uuid.New()
	expiredDate := time.Now().UTC().Add(-1 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: 1},
		TokensData: map[int64]config.TokenData{
			1: {
				Token:              "ab78bb36299d442fa0715fb53b5e3e57",
				TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
				RenewToken:         initialRenewToken,
				IdempotencyKey:     &idempotencyKey,
				NordLynxPrivateKey: "nordlynx-pkey",
				OpenVPNUsername:    "openvpn-username",
				OpenVPNPassword:    "openvpn-password",
				IsOAuth:            true,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}

	errorHandlerCalls := atomic.Int32{}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		errorHandlerCalls.Add(1)
		t.Logf("ERROR HANDLER CALLED (would trigger logout): %v", err)
	}, core.ErrNotFound)

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, api.Renew)

	const numGoroutines = 5
	var wg sync.WaitGroup
	renewErrors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			renewErrors[idx] = store.Renew()
		}(i)
	}

	wg.Wait()

	t.Logf("API call count: %d", api.callCount.Load())
	t.Logf("Success count: %d", api.successCount.Load())
	t.Logf("Fail count (404): %d", api.failCount.Load())
	t.Logf("Error handler calls: %d", errorHandlerCalls.Load())

	assert.Equal(t, int32(1), api.callCount.Load(), "Only one API call expected")
	assert.Equal(t, int32(1), api.successCount.Load(), "Single API call should succeed")
	assert.Equal(t, int32(0), api.failCount.Load(), "No 404 errors expected")
	assert.Equal(t, int32(0), errorHandlerCalls.Load(), "No spurious logouts")

	for i, err := range renewErrors {
		assert.NoError(t, err, "Goroutine %d should succeed", i)
	}
}

// TestAccessTokenSessionStore_IdempotencyKey_PreservedAfterRenewal verifies
// idempotency key is preserved after renewal for retry scenarios.
func TestAccessTokenSessionStore_IdempotencyKey_PreservedAfterRenewal(t *testing.T) {
	category.Set(t, category.Unit)

	var usedKey uuid.UUID
	renewAPI := func(token string, key uuid.UUID) (*session.AccessTokenResponse, error) {
		usedKey = key
		return &session.AccessTokenResponse{
			Token:      "ab78bb36299d442fa0715fb53b5e3e57",
			RenewToken: "deadbeef1234567890abcdef1234567890abcdef",
			ExpiresAt:  time.Now().UTC().Add(24 * time.Hour).Format(internal.ServerDateFormat),
		}, nil
	}

	initialKey := uuid.New()
	expiredDate := time.Now().UTC().Add(-1 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: 1},
		TokensData: map[int64]config.TokenData{
			1: {
				Token:              "ab78bb36299d442fa0715fb53b5e3e57",
				TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
				RenewToken:         "deadbeef1234567890abcdef1234567890abcdef",
				IdempotencyKey:     &initialKey,
				NordLynxPrivateKey: "nordlynx-pkey",
				OpenVPNUsername:    "openvpn-username",
				OpenVPNPassword:    "openvpn-password",
				IsOAuth:            true,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, renewAPI)

	err := store.Renew(session.ForceRenewal())
	assert.NoError(t, err, "Renewal should succeed")

	assert.Equal(t, initialKey, usedKey, "Initial idempotency key should be used")
	assert.NotNil(t, cfgManager.Cfg.TokensData[1].IdempotencyKey, "Key should be preserved")
	assert.Equal(t, initialKey, *cfgManager.Cfg.TokensData[1].IdempotencyKey, "Same key preserved")
}

// TestAccessTokenSessionStore_RaceCondition_WithExpiredToken tests race condition
// fix with expired token path (no ForceRenewal).
func TestAccessTokenSessionStore_RaceCondition_WithExpiredToken(t *testing.T) {
	category.Set(t, category.Unit)

	initialRenewToken := "deadbeef1234567890abcdef1234567890abcdef"
	api := newRaceConditionRenewalAPI(initialRenewToken)
	idempotencyKey := uuid.New()
	expiredDate := time.Now().UTC().Add(-1 * time.Hour)

	cfg := &config.Config{
		AutoConnectData: config.AutoConnectData{ID: 1},
		TokensData: map[int64]config.TokenData{
			1: {
				Token:              "ab78bb36299d442fa0715fb53b5e3e57",
				TokenExpiry:        expiredDate.Format(internal.ServerDateFormat),
				RenewToken:         initialRenewToken,
				IdempotencyKey:     &idempotencyKey,
				NordLynxPrivateKey: "nordlynx-pkey",
				OpenVPNUsername:    "openvpn-username",
				OpenVPNPassword:    "openvpn-password",
				IsOAuth:            true,
			},
		},
	}

	cfgManager := &mock.ConfigManager{Cfg: cfg}

	errorHandlerCalls := atomic.Int32{}
	errorRegistry := internal.NewErrorHandlingRegistry[error]()
	errorRegistry.Add(func(err error) {
		errorHandlerCalls.Add(1)
		t.Logf("ERROR HANDLER CALLED: %v", err)
	}, core.ErrNotFound)

	store := session.NewAccessTokenSessionStore(cfgManager, errorRegistry, api.Renew)

	const numGoroutines = 5
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(int) {
			defer wg.Done()
			_ = store.Renew()
		}(i)
	}

	wg.Wait()

	t.Logf("API call count: %d", api.callCount.Load())
	t.Logf("Success count: %d", api.successCount.Load())
	t.Logf("Fail count (404): %d", api.failCount.Load())
	t.Logf("Error handler calls: %d", errorHandlerCalls.Load())

	assert.Equal(t, int32(1), api.callCount.Load(), "Only one API call expected")
	assert.Equal(t, int32(0), errorHandlerCalls.Load(), "No spurious logouts")
}
