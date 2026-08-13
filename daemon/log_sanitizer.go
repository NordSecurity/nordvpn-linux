package daemon

import (
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/log"
)

type LogSanitizer struct {
	mu                    sync.Mutex
	logSanitizationUpdate events.Publisher[*pb.LogSanitizationEvent]
}

func NewLogSanitizer(logSanitizationUpdate events.Publisher[*pb.LogSanitizationEvent]) LogSanitizer {
	return LogSanitizer{
		logSanitizationUpdate: logSanitizationUpdate,
	}
}

// EnableSanitization enables log sanitization and publishes log sanitization update.
func (l *LogSanitizer) EnableSanitization(restrictedStrings ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	log.SetRestrictedStrings(restrictedStrings...)
	l.logSanitizationUpdate.Publish(&pb.LogSanitizationEvent{RestrictedStrings: restrictedStrings})
}

// DisableSanitization disables log sanitization and publishes log sanitization update.
func (l *LogSanitizer) DisableSanitization() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(log.GetRestrictedStrings()) > 0 {
		log.SetRestrictedStrings()
		l.logSanitizationUpdate.Publish(&pb.LogSanitizationEvent{})
	}
}

func (l *LogSanitizer) GetRestrictedLogStrings() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return log.GetRestrictedStrings()
}
