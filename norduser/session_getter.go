package norduser

import (
	"os"

	"github.com/NordSecurity/nordvpn-linux/log"
)

// sessionGetter provides access to users sessions
type sessionGetter interface {
	// getActiveUsers returns a set of users who have active sessions at the moment.
	getActiveUsers() (userData, error)
	// getDatabasePath returns path to the database of the utility used for monitoring the sessions.
	getDatabasePath() string
	// close cleanly closes any database connections used for session monitoring.
	close()
}

// newSessionGetter returns wtmpdbSessionGetter if wtmpdb database file is found and utmpSessionGetter if it isn't.
func newSessionGetter() sessionGetter {
	_, err := os.Stat(wtmpdbPath)
	if err == nil {
		wtmpdb := newWtmpdb()
		if err := wtmpdb.init(); err != nil {
			log.ProcessMonitor.Warn("failed to init wtmpdb:", err)
		} else {
			return wtmpdb
		}
	}
	if !os.IsNotExist(err) {
		log.Warn("failed to stat wtmpdb file: %w", err)
	}

	return &utmpSessionGetter{}
}
