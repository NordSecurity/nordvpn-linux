package norduser

import (
	"fmt"
	"os"
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
func newSessionGetter() (sessionGetter, error) {
	_, err := os.Stat(wtmpdbPath)
	if err == nil {
		wtmpdb := newWtmpdb()
		if err := wtmpdb.init(); err != nil {
			return nil, fmt.Errorf("failed to init wtmpdb: %w", err)
		}

		return wtmpdb, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat wtmpdb file: %w", err)
	}

	return &utmpSessionGetter{}, nil
}
