package request

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestNordAgentRoundTripper(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "header set",
			value:    `{"App":"NordVPN"}`,
			expected: `{"App":"NordVPN"}`,
		},
		{
			name:     "empty value",
			value:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sent *http.Request
			inner := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				sent = req
				return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("User-Agent", "test-agent")

			resp, err := NewNordAgentRoundTripper(inner, tt.value).RoundTrip(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.expected, sent.Header.Get(NordAgentHeader))
			assert.Equal(t, "test-agent", sent.Header.Get("User-Agent"))
			assert.Empty(t, req.Header.Get(NordAgentHeader))
		})
	}
}

func TestNordAgentRoundTripper_NilInnerFallsBackToDefault(t *testing.T) {
	category.Set(t, category.Unit)

	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(NordAgentHeader)
	}))
	defer srv.Close()

	client := &http.Client{Transport: NewNordAgentRoundTripper(nil, `{"App":"NordVPN"}`)}
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, `{"App":"NordVPN"}`, got)
}
