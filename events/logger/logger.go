package logger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// Subscriber is a subscriber for logging debug messages, info messages
// and error messages
type Subscriber struct{}

// NotifyMessage logs data with a debug prefix only in dev builds.
func (Subscriber) NotifyMessage(data string) error {
	log.Println(internal.DebugPrefix, data)
	return nil
}

// NotifyInfo logs data with an info prefix in production and dev
// builds
func (Subscriber) NotifyInfo(data string) error {
	log.Println(internal.InfoPrefix, data)
	return nil
}

// NotifyError logs an error with an error prefix in production and
// dev builds
func (Subscriber) NotifyError(err error) error {
	log.Println(internal.ErrorPrefix, err)
	return nil
}

func (Subscriber) NotifyRequestAPI(data events.DataRequestAPI) error {
	log.Printf("%s HTTP CALL %s",
		internal.InfoPrefix,
		dataRequestAPIToString(data, nil, nil, true),
	)
	return nil
}

func (Subscriber) NotifyRequestAPIVerbose(data events.DataRequestAPI) error {
	var reqBodyBytes []byte
	// Additional read of request body. Do not use in production builds
	if data.Request != nil && data.Request.GetBody != nil {
		body, _ := data.Request.GetBody()
		reqBodyBytes, _ = io.ReadAll(body)
	}

	// Additional read of response body. Do not use in production builds
	var respBodyBytes []byte
	if data.Response != nil {
		rawRespBodyBytes, _ := io.ReadAll(data.Response.Body)
		_ = data.Response.Body.Close()
		var reader io.Reader = bytes.NewBuffer(bytes.Clone(rawRespBodyBytes))
		if data.Response.Header.Get("Content-Encoding") == "gzip" {
			gReader, err := gzip.NewReader(io.NopCloser(reader))
			if err == nil {
				reader = gReader
			}
		}
		data.Response.Body = io.NopCloser(bytes.NewBuffer(rawRespBodyBytes))

		respBodyBytes, _ = io.ReadAll(reader)
	}
	log.Printf("%s HTTP CALL %s",
		internal.InfoPrefix,
		dataRequestAPIToString(data, reqBodyBytes, respBodyBytes, false),
	)
	return nil
}

func dataRequestAPIToString(
	data events.DataRequestAPI,
	reqBody []byte,
	respBody []byte,
	hideSensitiveHeaders bool,
) string {
	b := strings.Builder{}
	headers := processHeaders(hideSensitiveHeaders, data.Request.Header)
	b.WriteString(fmt.Sprintf("Duration: %s\n", data.Duration))
	if data.Request != nil {
		b.WriteString(fmt.Sprintf("Request: %s %s %s %s %s\n",
			data.Request.Proto,
			data.Request.Method,
			data.Request.URL,
			headers,
			string(reqBody),
		))
	}
	if data.Error != nil {
		b.WriteString(fmt.Sprintf("Error: %s\n", data.Error))
	}
	if data.Response != nil {
		b.WriteString(fmt.Sprintf("Response: %s %d - %s %s\n",
			data.Response.Proto,
			data.Response.StatusCode,
			data.Response.Header,
			string(respBody),
		))
	}

	return b.String()
}

func processHeaders(hide bool, headers http.Header) http.Header {
	if !hide {
		return headers
	}
	headers = headers.Clone()
	sensitiveHeaders := []string{
		"Authorization",
	}
	for _, header := range sensitiveHeaders {
		if headers.Get(header) != "" {
			headers.Set(header, "hidden")
		}
	}
	return headers
}
