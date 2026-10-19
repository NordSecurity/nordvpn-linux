package log

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"
)

var sanitizeRegex = regexp.MustCompile(
	// IPv4 address: matches valid IPv4 addresses from 0.0.0.0 to 255.255.255.255
	`\b(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}\b|` +
		// Hostname: matches only .nordvpn.com domains
		`\b[a-zA-Z0-9]+\.nordvpn\.com\b`,
)

var ipv6Regex = regexp.MustCompile(`[0-9a-fA-F:.%a-zA-Z]+`)

func maskString(s string) string {
	if strings.ContainsRune(s, ':') {
		s = ipv6Regex.ReplaceAllStringFunc(s, func(s string) string {
			if addr, err := netip.ParseAddr(s); err == nil && addr.Is6() {
				return "***"
			}
			return s
		})
	}
	if strings.ContainsRune(s, '.') { // cheap bail-out
		s = sanitizeRegex.ReplaceAllString(s, "***")
	}
	return s
}

type LogSafe struct {
	inner any
}

func DoNotSanitize(arg any) LogSafe {
	return LogSafe{
		inner: arg,
	}
}

func (l *LogSafe) Get() any {
	return l.inner
}

func sanitize(args ...any) []any {
	sanitizedArgs := make([]any, len(args))
	for pos, arg := range args {
		var sanitizedArg any
		switch v := arg.(type) {
		case LogSafe:
			sanitizedArg = v.Get()
		case error:
			sanitizedArg = maskString(v.Error())
		case string:
			sanitizedArg = maskString(v)
		case bool, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64, float32, float64:
			sanitizedArg = v
		default:
			sanitizedArg = maskString(fmt.Sprint(v))
		}
		sanitizedArgs[pos] = sanitizedArg
	}
	return sanitizedArgs
}

func unwrapLogSafe(args ...any) []any {
	unwrappedArgs := make([]any, len(args))
	for pos, arg := range args {
		if a, ok := arg.(LogSafe); ok {
			unwrappedArgs[pos] = a.Get()
		} else {
			unwrappedArgs[pos] = arg
		}
	}
	return unwrappedArgs
}
