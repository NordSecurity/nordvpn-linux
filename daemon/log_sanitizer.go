package daemon

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
)

type LogSanitizer struct {
	logSanitizationUpdate events.Publisher[*pb.LogSanitizationEvent]
}

func NewLogSanitizer(logSanitizationUpdate events.Publisher[*pb.LogSanitizationEvent]) LogSanitizer {
	return LogSanitizer{
		logSanitizationUpdate,
	}
}

// EnableSanitization enables log sanitization and publishes log sanitization update.
func (l *LogSanitizer) EnableSanitization(restrictedStrings ...string) {
	log.SetRestrictedStrings(restrictedStrings...)
	l.logSanitizationUpdate.Publish(&pb.LogSanitizationEvent{RestrictedStrings: restrictedStrings})
}

// DisableSanitization disables log sanitization and publishes log sanitization update.
func (l *LogSanitizer) DisableSanitization() {
	log.SetRestrictedStrings()
	l.logSanitizationUpdate.Publish(&pb.LogSanitizationEvent{})
}
