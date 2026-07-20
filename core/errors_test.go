//nolint:bodyclose
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func newMockJSONResponse(t *testing.T, statusCode int, payload any) *http.Response {
	t.Helper()

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal mock payload: %v", err)
	}

	h := make(http.Header)
	h.Set("Content-Type", "application/json")

	return &http.Response{
		StatusCode:    statusCode,
		Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Header:        h,
		Body:          io.NopCloser(bytes.NewReader(b)),
		ContentLength: int64(len(b)),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
	}
}

func TestExtractError(t *testing.T) {
	category.Set(t, category.Unit)
	tests := []struct {
		name          string
		resp          *http.Response
		expectedError error
	}{
		{
			name:          "no err",
			resp:          newMockJSONResponse(t, 200, nil),
			expectedError: nil,
		},
		{
			name:          "unauthorized",
			resp:          newMockJSONResponse(t, 401, nil),
			expectedError: ErrUnauthorized,
		},
		{
			name:          "forbidden",
			resp:          newMockJSONResponse(t, 403, nil),
			expectedError: ErrForbidden,
		},
		{
			name:          "not found",
			resp:          newMockJSONResponse(t, 404, nil),
			expectedError: ErrNotFound,
		},
		{
			name:          "status conflict",
			resp:          newMockJSONResponse(t, 409, nil),
			expectedError: ErrConflict,
		},
		{
			name:          "too many requests",
			resp:          newMockJSONResponse(t, 429, nil),
			expectedError: ErrTooManyRequests,
		},
		{
			name:          "internal server error",
			resp:          newMockJSONResponse(t, 500, nil),
			expectedError: ErrServerInternal,
		},
		{
			name:          "unhandled bad request",
			resp:          newMockJSONResponse(t, 400, nil),
			expectedError: ErrBadRequest,
		},
		// Login errors
		{
			name:          "login: invalid auth header",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "Invalid authorization header", "code": InvalidAuthorizationHeader}}),
			expectedError: ErrInvalidAuthHeader,
		},
		// Meshnet errors
		{
			name:          "meshnet: rate limit reached",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": RateLimitReachCode}}),
			expectedError: ErrRateLimitReach,
		},
		{
			name:          "meshnet: nickname too long",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": NicknameTooLongCode}}),
			expectedError: ErrNicknameTooLong,
		},
		{
			name:          "meshnet: duplicated nickname",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": DuplicateNicknameCode}}),
			expectedError: ErrDuplicateNickname,
		},
		{
			name:          "meshnet: forbidden word",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": ForbiddenWordCode}}),
			expectedError: ErrContainsForbiddenWord,
		},
		{
			name:          "meshnet: invalid prefix or suffix",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": InvalidPrefixOrSuffixCode}}),
			expectedError: ErrInvalidPrefixOrSuffix,
		},
		{
			name:          "meshnet: nickname double hyphen",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": NicknameHasDoubleHyphensCode}}),
			expectedError: ErrNicknameWithDoubleHyphens,
		},
		{
			name:          "meshnet: invalid chars",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": InvalidCharsCode}}),
			expectedError: ErrContainsInvalidChars,
		},
		{
			name:          "meshnet: max machine count",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": MaxMachineCountReached}}),
			expectedError: ErrMaximumDeviceCount,
		},
		{
			name:          "meshnet: max machine per peer count",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": MaxMachinePerPeerCountReached}}),
			expectedError: ErrMaximumDeviceCount,
		},
		{
			name:          "meshnet: max peer count on external machine",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": MaxPeerCountReachedOnExternalMachine}}),
			expectedError: ErrMaximumDeviceCount,
		},
		// Dedicated server errors
		{
			name:          "dedicated server: device not found",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": DeviceNotFound}}),
			expectedError: ErrDedicatedServersDeviceNotFound,
		},
		{
			name:          "dedicated server: device not registered",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": DeviceNotRegistered}}),
			expectedError: ErrDedicatedServersDeviceNotRegistered,
		},
		{
			name:          "dedicated server: invalid form data",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": InvalidFormData}}),
			expectedError: ErrDedicatedServersInvalidFormData,
		},
		{
			name:          "dedicated server: public key mismatch",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": PublicKeyMismatch}}),
			expectedError: ErrDedicatedServersPublicKeyMismatch,
		},
		{
			name:          "dedicated server: session limit hit",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": SessionLimitHit}}),
			expectedError: ErrDedicatedServersSessionMaxLimitReached,
		},
		{
			name:          "dedicated server: server offline",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": ServerOffline}}),
			expectedError: ErrDedicatedServersServerOffline,
		},
		{
			name:          "dedicated server: server not found",
			resp:          newMockJSONResponse(t, 400, map[string]map[string]any{"errors": {"message": "error message", "code": ServerNotFound}}),
			expectedError: ErrDedicatedServersServerNotFound,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.resp.Body.Close()
			err := ExtractError(test.resp)
			assert.ErrorIs(t, err, test.expectedError)
		})
	}
}
