package session

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateExpiry(t *testing.T) {
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

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "valid token",
			token:   "valid-token",
			wantErr: nil,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: ErrInvalidToken,
		},
		{
			name:    "whitespace only token",
			token:   "   ",
			wantErr: nil, // ValidateToken doesn't trim whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToken(tt.token)
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
			wantErr:    ErrMissingRenewToken,
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

func TestValidateExpiry_TimeBoundary(t *testing.T) {
	// Test with time very close to now
	almostNow := time.Now().Add(1 * time.Millisecond)
	err := ValidateExpiry(almostNow)
	// This might pass or fail depending on timing, but should not panic
	if err != nil {
		assert.Equal(t, ErrSessionExpired, err)
	}
}

func TestValidateAccessTokenFormat_ExactLengthBoundaries(t *testing.T) {
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
