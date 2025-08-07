package session

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestValidateExpiry(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		expiry  time.Time
		wantErr error
	}{
		{
			name:    "valid future expiry",
			expiry:  time.Now().Add(time.Hour),
			wantErr: nil,
		},
		{
			name:    "expired time",
			expiry:  time.Now().Add(-time.Hour),
			wantErr: ErrSessionExpired,
		},
		{
			name:    "expiry exactly now",
			expiry:  time.Now(),
			wantErr: ErrSessionExpired,
		},
		{
			name:    "far future expiry",
			expiry:  time.Now().Add(365 * 24 * time.Hour),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExpiry(tt.expiry)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAccessTokenFormat(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid hex token",
			token:   "ab78bb36299d442fa0715fb53b5e3e57",
			wantErr: nil,
		},
		{
			name:    "invalid hex token uppercase",
			token:   "AB78BB36299D442FA0715FB53B5E3E57",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "invalid hex token mixed case",
			token:   "Ab78Bb36299d442fA0715fB53b5e3e57",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "invalid format - not hex",
			token:   "not-a-hex-token",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "valid format - short hex",
			token:   "ab78",
			wantErr: nil,
		},
		{
			name:    "invalid format - contains special chars",
			token:   "ab78bb36-299d-442f-a071-5fb53b5e3e57",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAccessTokenFormat(tt.token)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTrustedPassTokenFormat(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid token with letters",
			token:   "validToken",
			wantErr: nil,
		},
		{
			name:    "valid token with numbers",
			token:   "token123",
			wantErr: nil,
		},
		{
			name:    "valid token with underscore",
			token:   "valid_token",
			wantErr: nil,
		},
		{
			name:    "valid token with hyphen",
			token:   "valid-token",
			wantErr: nil,
		},
		{
			name:    "valid token mixed",
			token:   "Valid-Token_123",
			wantErr: nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "invalid token with spaces",
			token:   "invalid token",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "invalid token with special chars",
			token:   "invalid@token",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "invalid token with dot",
			token:   "invalid.token",
			wantErr: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTrustedPassTokenFormat(tt.token)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTrustedPassOwnerID(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		ownerID string
		wantErr error
	}{
		{
			name:    "valid owner ID",
			ownerID: TrustedPassOwnerID,
			wantErr: nil,
		},
		{
			name:    "empty owner ID",
			ownerID: "",
			wantErr: ErrInvalidOwnerID,
		},
		{
			name:    "invalid owner ID",
			ownerID: "invalid",
			wantErr: ErrInvalidOwnerID,
		},
		{
			name:    "wrong case owner ID",
			ownerID: "NORDVPN",
			wantErr: ErrInvalidOwnerID,
		},
		{
			name:    "owner ID with spaces",
			ownerID: " nordvpn ",
			wantErr: ErrInvalidOwnerID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTrustedPassOwnerID(tt.ownerID)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRenewToken(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name       string
		renewToken string
		wantErr    error
	}{
		{
			name:       "valid hex token lowercase",
			renewToken: "ab78bb36299d442fa0715fb53b5e3e57",
			wantErr:    nil,
		},
		{
			name:       "invalid hex token uppercase",
			renewToken: "AB78BB36299D442FA0715FB53B5E3E57",
			wantErr:    ErrInvalidRenewToken,
		},
		{
			name:       "invalid hex token mixed case",
			renewToken: "aB78bB36299D442Fa0715Fb53B5E3e57",
			wantErr:    ErrInvalidRenewToken,
		},
		{
			name:       "empty token",
			renewToken: "",
			wantErr:    ErrInvalidRenewToken,
		},
		{
			name:       "invalid format - not hex",
			renewToken: "not-a-hex-token",
			wantErr:    ErrInvalidRenewToken,
		},
		{
			name:       "invalid format - contains special chars",
			renewToken: "ab78bb36-299d-442f-a071-5fb53b5e3e57",
			wantErr:    ErrInvalidRenewToken,
		},
		{
			name:       "valid short hex",
			renewToken: "ab78",
			wantErr:    nil,
		},
		{
			name:       "valid hex - deadbeef",
			renewToken: "deadbeef",
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRenewToken(tt.renewToken)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrustedPassExternalValidator(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		token     string
		ownerID   string
		validator TrustedPassExternalValidator
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "nil validator",
			token:     "token",
			ownerID:   "owner",
			validator: nil,
			wantErr:   false,
		},
		{
			name:    "validator returns nil",
			token:   "token",
			ownerID: "owner",
			validator: func(token, ownerID string) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "validator returns error",
			token:   "token",
			ownerID: "owner",
			validator: func(token, ownerID string) error {
				return errors.New("validation failed")
			},
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:    "validator checks token",
			token:   "expected-token",
			ownerID: "owner",
			validator: func(token, ownerID string) error {
				if token != "expected-token" {
					return errors.New("unexpected token")
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:    "validator checks ownerID",
			token:   "token",
			ownerID: "expected-owner",
			validator: func(token, ownerID string) error {
				if ownerID != "expected-owner" {
					return errors.New("unexpected owner")
				}
				return nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.validator != nil {
				err = tt.validator(tt.token, tt.ownerID)
			}

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOpenVPNCredentials(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{
			name:     "valid credentials",
			username: "user123",
			password: "pass456",
			wantErr:  nil,
		},
		{
			name:     "empty username",
			username: "",
			password: "pass456",
			wantErr:  ErrMissingVPNCredentials,
		},
		{
			name:     "empty password",
			username: "user123",
			password: "",
			wantErr:  ErrMissingVPNCredentials,
		},
		{
			name:     "both empty",
			username: "",
			password: "",
			wantErr:  ErrMissingVPNCredentials,
		},
		{
			name:     "whitespace username",
			username: "   ",
			password: "pass456",
			wantErr:  nil, // whitespace is considered valid
		},
		{
			name:     "whitespace password",
			username: "user123",
			password: "   ",
			wantErr:  nil, // whitespace is considered valid
		},
		{
			name:     "username with newline",
			username: "user\n123",
			password: "pass456",
			wantErr:  nil,
		},
		{
			name:     "password with tab",
			username: "user123",
			password: "pass\t456",
			wantErr:  nil,
		},
		{
			name:     "very long credentials",
			username: "verylongusernamethatexceedsnormallengthbutshouldbefine",
			password: "verylongpasswordthatexceedsnormallengthbutshouldbefine",
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOpenVPNCredentials(tt.username, tt.password)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNordLynxPrivateKey(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "valid key",
			key:     "abcdef123456789",
			wantErr: nil,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: ErrMissingNordLynxPrivateKey,
		},
		{
			name:    "whitespace key",
			key:     "   ",
			wantErr: nil, // whitespace is considered valid
		},
		{
			name:    "key with special characters",
			key:     "key-with-special-chars!@#",
			wantErr: nil, // any non-empty string is valid
		},
		{
			name:    "base64-like key",
			key:     "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=",
			wantErr: nil,
		},
		{
			name:    "hex-like key",
			key:     "deadbeef1234567890abcdef",
			wantErr: nil,
		},
		{
			name:    "key with equals padding",
			key:     "somekey==",
			wantErr: nil,
		},
		{
			name:    "key with forward slash",
			key:     "some/key/value",
			wantErr: nil,
		},
		{
			name:    "very long key",
			key:     "verylongkeythatexceedsnormallengthbutshouldbefineforprivatekeyusage1234567890",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNordLynxPrivateKey(tt.key)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExpiry_TimeBoundary(t *testing.T) {
	category.Set(t, category.Unit)

	// Test edge cases with deterministic time boundaries
	tests := []struct {
		name        string
		expiry      time.Time
		wantErr     error
		description string
	}{
		{
			name:        "far past expiry",
			expiry:      time.Now().Add(-365 * 24 * time.Hour), // 1 year ago
			wantErr:     ErrSessionExpired,
			description: "Should fail for time far in the past",
		},
		{
			name:        "just expired",
			expiry:      time.Now().Add(-1 * time.Second), // 1 second ago
			wantErr:     ErrSessionExpired,
			description: "Should fail for recently expired time",
		},
		{
			name:        "future expiry short duration",
			expiry:      time.Now().Add(1 * time.Second), // 1 second from now
			wantErr:     nil,
			description: "Should pass for near future time",
		},
		{
			name:        "future expiry medium duration",
			expiry:      time.Now().Add(1 * time.Hour), // 1 hour from now
			wantErr:     nil,
			description: "Should pass for medium future time",
		},
		{
			name:        "zero time",
			expiry:      time.Time{}, // Zero value of time.Time
			wantErr:     ErrSessionExpired,
			description: "Should fail for zero time value",
		},
		{
			name:        "unix epoch",
			expiry:      time.Unix(0, 0), // January 1, 1970
			wantErr:     ErrSessionExpired,
			description: "Should fail for Unix epoch time",
		},
		{
			name:        "max time",
			expiry:      time.Date(9999, 12, 31, 23, 59, 59, 999999999, time.UTC),
			wantErr:     nil,
			description: "Should pass for maximum representable time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExpiry(tt.expiry)
			if tt.wantErr != nil {
				assert.Error(t, err, tt.description)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}

	// Test concurrent access to ensure thread safety
	t.Run("concurrent validation", func(t *testing.T) {
		futureTime := time.Now().Add(time.Hour)
		pastTime := time.Now().Add(-time.Hour)

		// Run multiple goroutines to test concurrent access
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(i int) {
				if i%2 == 0 {
					err := ValidateExpiry(futureTime)
					assert.NoError(t, err)
				} else {
					err := ValidateExpiry(pastTime)
					assert.Equal(t, ErrSessionExpired, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestValidateAccessTokenFormat_ExactLengthBoundaries(t *testing.T) {
	category.Set(t, category.Unit)

	// Test with various hex string lengths
	testCases := []struct {
		length  int
		wantErr bool
	}{
		{0, true},   // empty
		{1, false},  // valid hex of any length
		{8, false},  // valid hex of any length
		{16, false}, // valid hex of any length
		{32, false}, // typical MD5 length
		{40, false}, // typical SHA1 length
		{64, false}, // typical SHA256 length
	}

	for _, tc := range testCases {
		token := ""
		for i := 0; i < tc.length; i++ {
			token += "a"
		}

		err := ValidateAccessTokenFormat(token)
		if tc.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestValidateTrustedPassTokenFormat_EdgeCharacters(t *testing.T) {
	category.Set(t, category.Unit)

	// Test boundary characters
	testCases := []struct {
		token   string
		wantErr bool
	}{
		{"0", false}, // start of numbers
		{"9", false}, // end of numbers
		{"A", false}, // start of uppercase
		{"Z", false}, // end of uppercase
		{"a", false}, // start of lowercase
		{"z", false}, // end of lowercase
		{"_", false}, // underscore
		{"-", false}, // hyphen
		{"@", true},  // before uppercase A
		{"[", true},  // after uppercase Z
		{"`", true},  // before lowercase a
		{"{", true},  // after lowercase z
		{"/", true},  // before 0
		{":", true},  // after 9
	}

	for _, tc := range testCases {
		err := ValidateTrustedPassTokenFormat(tc.token)
		if tc.wantErr {
			assert.Error(t, err, "token: %s", tc.token)
		} else {
			assert.NoError(t, err, "token: %s", tc.token)
		}
	}
}
