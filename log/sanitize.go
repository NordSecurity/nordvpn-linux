package log

import (
	"fmt"
	"regexp"
)

type DoNotSanitize struct {
	inner string
}

func (d DoNotSanitize) Get() string {
	return d.inner
}

var ipRegex = regexp.MustCompile(`\b(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}\b`)

func maskString(s string) string {
	// if !strings.ContainsRune(s, '.') { // cheap bail-out
	// 	return s
	// }
	return ipRegex.ReplaceAllString(s, "***")
}

func sanitize(args ...any) []any {
	sanitizedArgs := make([]any, len(args))
	for idx, arg := range args {
		var sanitizedArg any
		switch v := arg.(type) {
		case DoNotSanitize:
			sanitizedArg = v.Get()
		case error:
			sanitizedArg = maskString(v.Error())
		case string:
			sanitizedArg = maskString(v)
		case bool, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64, float32, float64:
			sanitizedArg = v
		default:
			sanitizedArg = fmt.Sprintf("<%T>", v)
		}
		sanitizedArgs[idx] = sanitizedArg
	}
	return sanitizedArgs
}
