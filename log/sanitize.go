package log

import "sync/atomic"

var stringsToSanitize atomic.Pointer[[]string]

// SetRestrictedStrings sets strings to be sanitized out of the logs
func SetRestrictedStrings(restrictedStrings ...string) {
	stringsToSanitize.Store(&restrictedStrings)
}
