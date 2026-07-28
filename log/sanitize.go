package log

import "sync/atomic"

var stringsToSanitize atomic.Pointer[[]string]

func EnableSanitization(restrictedStrings ...string) {
	stringsToSanitize.Store(&restrictedStrings)
}

func DisableSanitization() {
	stringsToSanitize.Store(&[]string{})
}
