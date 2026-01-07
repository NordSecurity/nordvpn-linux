package core

import (
	"net/http"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core"
)

// testNewCDNAPI returns a pointer to initialized and
// ready for use in tests CDNAPI
// url can be obtained by creating a NewHTTPTestServer()
func NewTestingCDNAPI(t *testing.T, url string) *core.CDNAPI {
	t.Helper()

	cdn := core.NewCDNAPI(
		"",
		url,
		http.DefaultClient,
		nil,
	)

	return cdn
}
