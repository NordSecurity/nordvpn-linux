package session

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionTokenProvider struct {
	token string
}

func (m *mockSessionTokenProvider) GetToken() string {
	return m.token
}

func Test_TokenValidator_Validate(t *testing.T) {
	t.Run("Valid token", func(t *testing.T) {
		validToken := "valid-token"
		session := &mockSessionTokenProvider{token: validToken}

		validateCalled := false
		validator := NewTokenValidator(func(token string) error {
			validateCalled = true
			assert.Equal(t, validToken, token)
			return nil
		})

		err := validator.Validate(session)

		assert.NoError(t, err)
		assert.True(t, validateCalled)
	})

	t.Run("Invalid token", func(t *testing.T) {
		invalidToken := "invalid-token"
		session := &mockSessionTokenProvider{token: invalidToken}
		expectedError := errors.New("token validation failed")

		validator := NewTokenValidator(func(token string) error {
			assert.Equal(t, invalidToken, token)
			return expectedError
		})

		err := validator.Validate(session)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Non-SessionTokenProvider session", func(t *testing.T) {
		nonTokenSession := "not-a-token-provider"

		validator := NewTokenValidator(func(token string) error {
			t.Fatal("This should not be called")
			return nil
		})

		err := validator.Validate(nonTokenSession)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported store")
	})

	t.Run("Validation returns error", func(t *testing.T) {
		session := &mockSessionTokenProvider{token: "some-token"}
		expectedError := errors.New("validation error")

		validator := NewTokenValidator(func(token string) error {
			return expectedError
		})

		err := validator.Validate(session)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Validation returns expected data", func(t *testing.T) {
		expectedToken := "expected-data-token"
		session := &mockSessionTokenProvider{token: expectedToken}

		validator := NewTokenValidator(func(token string) error {
			assert.Equal(t, expectedToken, token)
			return nil
		})

		err := validator.Validate(session)

		assert.NoError(t, err)
	})
}

func Test_NewTokenValidator(t *testing.T) {
	t.Run("Creates validator with provided function", func(t *testing.T) {
		validateFn := func(token string) error {
			return nil
		}

		validator := NewTokenValidator(validateFn)

		require.NotNil(t, validator)

		tokenValidator, ok := validator.(*TokenValidator)
		require.True(t, ok)

		assert.NotNil(t, tokenValidator.ValidateFunc)

		err := tokenValidator.ValidateFunc("test")
		assert.NoError(t, err)
	})
}

type mockExpirableSession struct {
	expired bool
}

func (m *mockExpirableSession) IsExpired() bool {
	return m.expired
}

func Test_ExpiryValidator_Validate(t *testing.T) {
	t.Run("Non-expired session", func(t *testing.T) {
		session := &mockExpirableSession{expired: false}

		validator := NewExpiryValidator()

		err := validator.Validate(session)

		assert.NoError(t, err)
	})

	t.Run("Expired session", func(t *testing.T) {
		session := &mockExpirableSession{expired: true}

		validator := NewExpiryValidator()

		err := validator.Validate(session)

		assert.Error(t, err)
		assert.Equal(t, ErrSessionExpired, err)
	})

	t.Run("Non-ExpirableSession type", func(t *testing.T) {
		nonExpirableSession := "not-an-expirable-session"

		validator := NewExpiryValidator()

		err := validator.Validate(nonExpirableSession)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported store")
	})

	t.Run("Validate returns error for expired session", func(t *testing.T) {
		session := &mockExpirableSession{expired: true}

		validator := NewExpiryValidator()

		err := validator.Validate(session)

		assert.Error(t, err)
		assert.Equal(t, ErrSessionExpired, err)
	})

	t.Run("Validate returns no error for valid session", func(t *testing.T) {
		session := &mockExpirableSession{expired: false}

		validator := NewExpiryValidator()

		err := validator.Validate(session)

		assert.NoError(t, err)
	})
}

func Test_NewExpiryValidator(t *testing.T) {
	t.Run("Creates validator correctly", func(t *testing.T) {
		validator := NewExpiryValidator()

		require.NotNil(t, validator)

		expiryValidator, ok := validator.(*ExpiryValidator)
		require.True(t, ok)

		assert.NotNil(t, expiryValidator)
	})
}

type mockCredentialsBasedSession struct {
	username string
	password string
}

func (m *mockCredentialsBasedSession) GetUsername() string {
	return m.username
}

func (m *mockCredentialsBasedSession) GetPassword() string {
	return m.password
}

func Test_CredentialValidator_Validate(t *testing.T) {
	t.Run("Valid credentials", func(t *testing.T) {
		validUsername := "valid-user"
		validPassword := "valid-pass"
		session := &mockCredentialsBasedSession{
			username: validUsername,
			password: validPassword,
		}

		validateCalled := false
		validator := NewCredentialValidator(func(username, password string) error {
			validateCalled = true
			assert.Equal(t, validUsername, username)
			assert.Equal(t, validPassword, password)
			return nil
		})

		err := validator.Validate(session)

		assert.NoError(t, err)
		assert.True(t, validateCalled)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		invalidUsername := "invalid-user"
		invalidPassword := "invalid-pass"
		session := &mockCredentialsBasedSession{
			username: invalidUsername,
			password: invalidPassword,
		}
		expectedError := errors.New("credentials validation failed")

		validator := NewCredentialValidator(func(username, password string) error {
			assert.Equal(t, invalidUsername, username)
			assert.Equal(t, invalidPassword, password)
			return expectedError
		})

		err := validator.Validate(session)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Non-CredentialsBasedSession session", func(t *testing.T) {
		nonCredentialsSession := "not-a-credentials-provider"

		validator := NewCredentialValidator(func(username, password string) error {
			t.Fatal("This should not be called")
			return nil
		})

		err := validator.Validate(nonCredentialsSession)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported store")
	})

	t.Run("Validation returns error", func(t *testing.T) {
		session := &mockCredentialsBasedSession{
			username: "some-user",
			password: "some-pass",
		}
		expectedError := errors.New("validation error")

		validator := NewCredentialValidator(func(username, password string) error {
			return expectedError
		})

		err := validator.Validate(session)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Validation returns expected data", func(t *testing.T) {
		expectedUsername := "expected-username"
		expectedPassword := "expected-password"
		session := &mockCredentialsBasedSession{
			username: expectedUsername,
			password: expectedPassword,
		}

		validator := NewCredentialValidator(func(username, password string) error {
			assert.Equal(t, expectedUsername, username)
			assert.Equal(t, expectedPassword, password)
			return nil
		})

		err := validator.Validate(session)

		assert.NoError(t, err)
	})
}

func Test_NewCredentialValidator(t *testing.T) {
	t.Run("Creates validator with provided function", func(t *testing.T) {
		validateFn := func(username, password string) error {
			return nil
		}

		validator := NewCredentialValidator(validateFn)

		require.NotNil(t, validator)

		credentialValidator, ok := validator.(*CredentialValidator)
		require.True(t, ok)

		assert.NotNil(t, credentialValidator.ValidateFunc)

		err := credentialValidator.ValidateFunc("testuser", "testpass")
		assert.NoError(t, err)
	})
}

type mockValidator struct {
	validateFunc func(session any) error
}

func (m *mockValidator) Validate(session any) error {
	if m.validateFunc != nil {
		return m.validateFunc(session)
	}
	return nil
}

type mockCombinedSession struct {
	token   string
	expired bool
}

func (m *mockCombinedSession) GetToken() string {
	return m.token
}

func (m *mockCombinedSession) IsExpired() bool {
	return m.expired
}

func Test_CompositeValidator_Validate(t *testing.T) {
	t.Run("All validators pass", func(t *testing.T) {
		mockSession := struct{}{}

		validator1Called := false
		validator1 := &mockValidator{
			validateFunc: func(s any) error {
				validator1Called = true
				return nil
			},
		}

		validator2Called := false
		validator2 := &mockValidator{
			validateFunc: func(s any) error {
				validator2Called = true
				return nil
			},
		}

		composite := NewCompositeValidator(validator1, validator2)

		err := composite.Validate(mockSession)

		assert.NoError(t, err)
		assert.True(t, validator1Called)
		assert.True(t, validator2Called)
	})

	t.Run("First validator fails", func(t *testing.T) {
		mockSession := struct{}{}
		expectedError := errors.New("validator 1 failed")

		validator1 := &mockValidator{
			validateFunc: func(s any) error {
				return expectedError
			},
		}

		validator2Called := false
		validator2 := &mockValidator{
			validateFunc: func(s any) error {
				validator2Called = true
				return nil
			},
		}

		composite := NewCompositeValidator(validator1, validator2)

		err := composite.Validate(mockSession)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.False(t, validator2Called, "Second validator should not be called after first fails")
	})

	t.Run("Second validator fails", func(t *testing.T) {
		mockSession := struct{}{}
		expectedError := errors.New("validator 2 failed")

		validator1Called := false
		validator1 := &mockValidator{
			validateFunc: func(s any) error {
				validator1Called = true
				return nil
			},
		}

		validator2 := &mockValidator{
			validateFunc: func(s any) error {
				return expectedError
			},
		}

		composite := NewCompositeValidator(validator1, validator2)

		err := composite.Validate(mockSession)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.True(t, validator1Called, "First validator should be called")
	})

	t.Run("No validators", func(t *testing.T) {
		mockSession := struct{}{}

		composite := NewCompositeValidator()

		err := composite.Validate(mockSession)

		assert.NoError(t, err)
	})

	t.Run("Validators are executed in order", func(t *testing.T) {
		mockSession := struct{}{}
		executionOrder := []int{}

		validator1 := &mockValidator{
			validateFunc: func(s any) error {
				executionOrder = append(executionOrder, 1)
				return nil
			},
		}

		validator2 := &mockValidator{
			validateFunc: func(s any) error {
				executionOrder = append(executionOrder, 2)
				return nil
			},
		}

		validator3 := &mockValidator{
			validateFunc: func(s any) error {
				executionOrder = append(executionOrder, 3)
				return nil
			},
		}

		composite := NewCompositeValidator(validator1, validator2, validator3)

		err := composite.Validate(mockSession)

		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, executionOrder)
	})

	t.Run("Integration with real validators", func(t *testing.T) {
		session := &mockCombinedSession{
			token:   "valid-token",
			expired: false,
		}

		tokenValidator := NewTokenValidator(func(token string) error {
			if token != "valid-token" {
				return errors.New("invalid token")
			}
			return nil
		})

		expiryValidator := NewExpiryValidator()

		composite := NewCompositeValidator(tokenValidator, expiryValidator)

		err := composite.Validate(session)
		assert.NoError(t, err)

		session.token = "invalid-token"
		err = composite.Validate(session)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")

		session.token = "valid-token"
		session.expired = true
		err = composite.Validate(session)
		assert.Error(t, err)
		assert.Equal(t, ErrSessionExpired, err)
	})
}

func Test_NewCompositeValidator(t *testing.T) {
	t.Run("Creates validator with provided validators", func(t *testing.T) {
		validator1 := &mockValidator{}
		validator2 := &mockValidator{}

		composite := NewCompositeValidator(validator1, validator2)

		require.NotNil(t, composite)

		compositeValidator, ok := composite.(*CompositeValidator)
		require.True(t, ok)

		assert.Len(t, compositeValidator.Validators, 2)
		assert.Equal(t, validator1, compositeValidator.Validators[0])
		assert.Equal(t, validator2, compositeValidator.Validators[1])
	})

	t.Run("Creates validator with no validators", func(t *testing.T) {
		composite := NewCompositeValidator()

		require.NotNil(t, composite)

		compositeValidator, ok := composite.(*CompositeValidator)
		require.True(t, ok)

		assert.Empty(t, compositeValidator.Validators)
	})
}
