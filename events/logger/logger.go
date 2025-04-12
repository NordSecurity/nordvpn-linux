package logger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"regexp"
	"slices"
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

func (s Subscriber) NotifyConnect(data events.DataConnect) error {
	eventName := "POST_CONNECT"
	switch data.EventStatus {
	case events.StatusSuccess:
		log.Println(internal.InfoPrefix, "connected to", data.TargetServerDomain)
	case events.StatusFailure:
		log.Println(internal.ErrorPrefix, "failed to connect to", data.TargetServerDomain, ":", data.Error)
	case events.StatusCanceled:
		log.Println(internal.InfoPrefix, "connection to", data.TargetServerDomain, "was cancelled")
	case events.StatusAttempt:
		eventName = "PRE_CONNECT"
	}
	log.Printf("%s %s system info:\n%s\n%s\n", internal.InfoPrefix, eventName, getSystemInfo(), getNetworkInfo())
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
			headers, // Ensure headers are properly obfuscated
			"(request body hidden)", // Do not log request body
		))
	}
	if data.Error != nil {
		b.WriteString(fmt.Sprintf("Error: %s\n", data.Error))
	}
	if data.Response != nil {
		tmpBody := "(binary data)"
		// do not print binary data
		if !slices.Contains(data.Response.Header.Values("Content-Type"), "application/octet-stream") {
			tmpBody = string(respBody)
		}
		b.WriteString(fmt.Sprintf("Response: %s %d - %s %s\n",
			data.Response.Proto,
			data.Response.StatusCode,
			data.Response.Header,
			tmpBody,
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
		"Cookie",
		"Set-Cookie",
		"Proxy-Authorization",
		"WWW-Authenticate",
		"Proxy-Authenticate",
	}
	for _, header := range sensitiveHeaders {
		if headers.Get(header) != "" {
			headers.Set(header, "hidden")
		}
	}
	return headers
}

func getSystemInfo() string {
	builder := strings.Builder{}
	out, err := os.ReadFile("/etc/os-release")
	if err == nil {
		builder.WriteString("OS Info:\n" + string(out) + "\n")
	}
	out, err = exec.Command("uname", "-a").CombinedOutput()
	if err == nil {
		builder.WriteString("System Info:" + string(out) + "\n")
	}
	return builder.String()
}

// maskIPRouteOutput changes any non-local ip address in the output to ***
func maskIPRouteOutput(output string) string {
	expIPv4 := regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	expIPv6 := regexp.MustCompile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}` +
		`|([0-9a-fA-F]{1,4}:){1,7}:` +
		`|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}` +
		`|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}` +
		`|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}` +
		`|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}` +
		`|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}` +
		`|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})` +
		`|:((:[0-9a-fA-F]{1,4}){1,7}|:)` +
		`|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}` +
		`|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])` +
		`|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)

	ips := expIPv4.FindAllString(output, -1)
	ips = append(ips, expIPv6.FindAllString(output, -1)...)
	for _, ip := range ips {
		parsed, err := netip.ParseAddr(ip)
		if err != nil {
			log.Println(internal.WarningPrefix,
				"Failed to parse ip address %s for masking: %v", ip, err)
			continue
		}

		if !parsed.IsLinkLocalMulticast() && !parsed.IsLinkLocalUnicast() && !parsed.IsLoopback() && !parsed.IsPrivate() {
			output = strings.Replace(output, ip, "***", -1)
		}
	}

	return output
}

func getNetworkInfo() string {
	builder := strings.Builder{}
	for _, arg := range []string{"4", "6"} {
		// #nosec G204 -- arg values are known before even running the program
		out, err := exec.Command("ip", "-"+arg, "route", "show", "table", "all").CombinedOutput()
		if err != nil {
			continue
		}
		maskedOutput := maskIPRouteOutput(string(out))
		builder.WriteString("Routes for ipv" + arg + ":\n")
		builder.WriteString(maskedOutput)

		// #nosec G204 -- arg values are known before even running the program
		out, err = exec.Command("ip", "-"+arg, "rule").CombinedOutput()
		if err != nil {
			continue
		}
		builder.WriteString("IP rules for ipv" + arg + ":\n" + string(out) + "\n")
	}

	for _, iptableVersion := range internal.GetSupportedIPTables() {
		tableRules := ""
		for _, table := range []string{"filter", "nat", "mangle", "raw", "security"} {
			// #nosec G204 -- input is properly sanitized
			out, err := exec.Command(iptableVersion, "-S", "-t", table, "-w", internal.SecondsToWaitForIptablesLock).CombinedOutput()
			if err == nil {
				tableRules += table + ":\n" + string(out) + "\n"
			}
		}
		version := "4"
		if iptableVersion == "ip6tables" {
			version = "6"
		}
		builder.WriteString("IP tables for ipv" + version + ":\n" + tableRules)
	}

	return builder.String()
}
