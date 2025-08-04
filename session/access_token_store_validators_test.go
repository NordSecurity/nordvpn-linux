package session_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/session"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionWithTokenAndExpiry struct {
	token  string
	expiry time.Time
}

func (m *mockSessionWithTokenAndExpiry) GetToken() string {
	return m.token
}

func (m *mockSessionWithTokenAndExpiry) GetExpiry() time.Time {
	return m.expiry
}

type mockSessionWithTokenOnly struct {
	token string
}

func (m *mockSessionWithTokenOnly) GetToken() string {
	return m.token
}

type mockSessionWithExpiryOnly struct {
	expiry time.Time
}

func (m *mockSessionWithExpiryOnly) GetExpiry() time.Time {
	return m.expiry
}

func Test_ManualAccessTokenValidator_Validate(t *testing.T) {
	category.Set(t, category.Unit)

	t.Run("Non-token provider session", func(t *testing.T) {
		expiryOnlySession := &mockSessionWithExpiryOnly{expiry: session.ManualAccessTokenExpiryDate}

		apiCalled := false
		api := func(token string) error {
			apiCalled = true
			return nil
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(expiryOnlySession)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported store for token validation")
		assert.False(t, apiCalled)
	})

	t.Run("Non-expiry provider session", func(t *testing.T) {
		tokenOnlySession := &mockSessionWithTokenOnly{token: "valid-token"}

		apiCalled := false
		api := func(token string) error {
			apiCalled = true
			return nil
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(tokenOnlySession)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported store for expiry validation")
		assert.False(t, apiCalled)
	})

	t.Run("Not a manual access token (different expiry date)", func(t *testing.T) {
		regularExpiry := time.Now().Add(24 * time.Hour)
		store := &mockSessionWithTokenAndExpiry{
			token:  "valid-token",
			expiry: regularExpiry,
		}

		apiCalled := false
		api := func(token string) error {
			apiCalled = true
			return nil
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(store)

		assert.NoError(t, err)
		assert.False(t, apiCalled, "API should not be called for non-manual tokens")
	})

	t.Run("Manual token with invalid format", func(t *testing.T) {
		originalValidator := internal.AccessTokenFormatValidator
		internal.AccessTokenFormatValidator = func(token string) bool { return false }
		// internal workaround
		defer func() { internal.AccessTokenFormatValidator = originalValidator }()

		store := &mockSessionWithTokenAndExpiry{
			token:  "invalid-format-token",
			expiry: session.ManualAccessTokenExpiryDate,
		}

		apiCalled := false
		api := func(token string) error {
			apiCalled = true
			return nil
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(store)

		assert.Error(t, err)
		assert.ErrorIs(t, err, session.ErrAccessTokenExpired)
		assert.False(t, apiCalled, "API should not be called for invalid format tokens")
	})

	t.Run("Manual token with valid format but revoked", func(t *testing.T) {
		originalValidator := internal.AccessTokenFormatValidator
		internal.AccessTokenFormatValidator = func(token string) bool { return true }
		// internal workaround
		defer func() { internal.AccessTokenFormatValidator = originalValidator }()

		store := &mockSessionWithTokenAndExpiry{
			token:  "valid-format-but-revoked",
			expiry: session.ManualAccessTokenExpiryDate,
		}

		api := func(token string) error {
			assert.Equal(t, "valid-format-but-revoked", token)
			return session.ErrAccessTokenRevoked
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(store)

		assert.Error(t, err)
		assert.ErrorIs(t, err, session.ErrAccessTokenRevoked)
	})

	t.Run("Manual token with valid format and active", func(t *testing.T) {
		originalValidator := internal.AccessTokenFormatValidator
		internal.AccessTokenFormatValidator = func(token string) bool { return true }
		// internal workaround
		defer func() { internal.AccessTokenFormatValidator = originalValidator }()

		store := &mockSessionWithTokenAndExpiry{
			token:  "valid-active-token",
			expiry: session.ManualAccessTokenExpiryDate,
		}

		apiCalled := false
		api := func(token string) error {
			apiCalled = true
			assert.Equal(t, "valid-active-token", token)
			return nil
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(store)

		assert.NoError(t, err)
		assert.True(t, apiCalled)
	})

	t.Run("Manual token with valid format. API error is passed through", func(t *testing.T) {
		originalValidator := internal.AccessTokenFormatValidator
		internal.AccessTokenFormatValidator = func(token string) bool { return true }
		// internal workaround
		defer func() { internal.AccessTokenFormatValidator = originalValidator }()

		store := &mockSessionWithTokenAndExpiry{
			token:  "valid-token",
			expiry: session.ManualAccessTokenExpiryDate,
		}

		apiError := errors.New("some other API error")
		api := func(token string) error {
			return apiError
		}

		validator := session.NewManualAccessTokenValidator(api)

		err := validator.Validate(store)
		fmt.Println("err:", err)
		assert.ErrorIs(t, err, apiError, "Input/Output error must match")
	})
}

func Test_NewManualAccessTokenValidator(t *testing.T) {
	category.Set(t, category.Unit)

	t.Run("Creates validator with provided API function", func(t *testing.T) {
		apiCalled := false
		api := func(token string) error {
			apiCalled = true
			return nil
		}

		validator := session.NewManualAccessTokenValidator(api)

		require.NotNil(t, validator)

		originalValidator := internal.AccessTokenFormatValidator
		internal.AccessTokenFormatValidator = func(token string) bool { return true }
		// internal workaround
		defer func() { internal.AccessTokenFormatValidator = originalValidator }()

		store := &mockSessionWithTokenAndExpiry{
			token:  "test-token",
			expiry: session.ManualAccessTokenExpiryDate,
		}

		err := validator.Validate(store)

		assert.NoError(t, err)
		assert.True(t, apiCalled, "API function should be called")
	})
}
