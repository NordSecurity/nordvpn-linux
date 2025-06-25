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
	"github.com/NordSecurity/nordvpn-linux/request"
)

func TestCDN(t *testing.T) {
	var err error
	var validator response.Validator
	if true {
		validator = response.NoopValidator{}
	} else {
		validator, err = response.NewNordValidator()
		if err != nil {
			log.Fatalln("Error on creating validator:", err)
		}
	}

	userAgent := fmt.Sprintf("NordApp Linux %s %s", "3.33.3", "distro.KernelName")

	httpGlobalCtx, httpCancel := context.WithCancel(context.Background())
	httpCallsSubject := &subs.Subject[events.DataRequestAPI]{}

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

	cdnAPI.GetRemoteFile("nordvpn")

	httpCancel()
}
