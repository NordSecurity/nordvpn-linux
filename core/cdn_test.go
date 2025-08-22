package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestCdnApi(t *testing.T) {
	category.Set(t, category.Integration)

	cdnApi, cancel := setupCdnApi()
	assert.NotNil(t, cdnApi)

	nameservers, err := cdnApi.ThreatProtectionLite()
	assert.NoError(t, err)
	assert.NotNil(t, nameservers)

	_, fileBytes, err := cdnApi.ConfigTemplate(false, http.MethodGet)
	assert.NoError(t, err)
	assert.NotZero(t, len(fileBytes))

	fileBytes, err = cdnApi.GetRemoteFile("/configs/templates/ovpn/1.0/template.xslt")
	assert.NoError(t, err)
	assert.NotZero(t, len(fileBytes))

	cancel()
}

func setupCdnApi() (*CDNAPI, context.CancelFunc) {
	Environment := "dev"
	Version := "3.3.3"
	httpCallsSubject := &subs.Subject[events.DataRequestAPI]{}

	// API
	var err error
	var validator response.Validator
	if !internal.IsProdEnv(Environment) {
		validator = response.NoopValidator{}
	} else {
		validator, err = response.NewNordValidator()
		if err != nil {
			log.Fatalln("Error on creating validator:", err)
		}
	}

	userAgent, err := request.GetUserAgentValue(Version, sysinfo.GetHostOSPrettyName)
	if err != nil {
		userAgent = fmt.Sprintf("%s/%s (unknown)", request.AppName, Version)
		log.Printf("Error while constructing UA value: %s. Falls back to default: %s\n", err, userAgent)
	}

	httpGlobalCtx, httpCancel := context.WithCancel(context.Background())

	// simple standard http client with dialer wrapped inside
	httpClientSimple := request.NewStdHTTP()
	httpClientSimple.Transport = request.NewHTTPReTransport(
		1, 1, "HTTP/1.1", func() http.RoundTripper {
			return request.NewPublishingRoundTripper(
				request.NewContextRoundTripper(request.NewStdTransport(), httpGlobalCtx),
				httpCallsSubject,
			)
		}, nil)

	cdnAPI := NewCDNAPI(
		userAgent,
		CDNURL,
		httpClientSimple,
		validator,
	)

	return cdnAPI, httpCancel
}
