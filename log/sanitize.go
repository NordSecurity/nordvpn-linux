package log

import (
	"fmt"
	"regexp"
	"strings"
)

var sanitizeRegex = regexp.MustCompile(
	// IPv4 address: matches valid IPv4 addresses from 0.0.0.0 to 255.255.255.255
	`\b(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}\b|` +
		// Hostname: matches only .nordvpn.com domains
		`\b[a-zA-Z0-9]+\.nordvpn\.com\b`,
)

var ipv6Regex = regexp.MustCompile(
	`(?:(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|` + // 1:2:3:4:5:6:7:8
		`fe80:(?::[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|` + // fe80::7:8%eth0 (link-local with zone index)
		`::(?:ffff(?::0{1,4})?:)?` + // ::255.255.255.255  ::ffff:255.255.255.255 (IPv4-mapped/translated)
		`(?:(?:25[0-5]|(?:2[0-4]|1?[0-9])?[0-9])\.){3}` +
		`(?:25[0-5]|(?:2[0-4]|1?[0-9])?[0-9])|` +
		`(?:[0-9a-fA-F]{1,4}:){1,4}:` + // 2001:db8:3:4::192.0.2.33 (IPv4-embedded)
		`(?:(?:25[0-5]|(?:2[0-4]|1?[0-9])?[0-9])\.){3}` +
		`(?:25[0-5]|(?:2[0-4]|1?[0-9])?[0-9])|` +
		`[0-9a-fA-F]{1,4}:(?::[0-9a-fA-F]{1,4}){1,6}|` + // 1::3:4:5:6:7:8
		`(?:[0-9a-fA-F]{1,4}:){1,2}(?::[0-9a-fA-F]{1,4}){1,5}|` + // 1:2::4:5:6:7:8
		`(?:[0-9a-fA-F]{1,4}:){1,3}(?::[0-9a-fA-F]{1,4}){1,4}|` + // 1:2:3::5:6:7:8
		`(?:[0-9a-fA-F]{1,4}:){1,4}(?::[0-9a-fA-F]{1,4}){1,3}|` + // 1:2:3:4::6:7:8
		`(?:[0-9a-fA-F]{1,4}:){1,5}(?::[0-9a-fA-F]{1,4}){1,2}|` + // 1:2:3:4:5::7:8
		`(?:[0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|` + // 1:2:3:4:5:6::8
		`(?:[0-9a-fA-F]{1,4}:){1,7}:|` + // 1::  1:2:3:4:5:6:7::
		`:(?::[0-9a-fA-F]{1,4}){1,7}|` + // ::2:3:4:5:6:7:8  ::8
		`::)`, // ::
)

func maskString(s string) string {
	if strings.ContainsRune(s, ':') {
		s = ipv6Regex.ReplaceAllString(s, "***")
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
