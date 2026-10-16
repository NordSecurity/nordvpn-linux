package norduser

import (
	"fmt"
	"os"
)

type sessionGetter interface {
	getActiveUsers() (userData, error)
	getDatabasePath() string
	close()
}

func newSessionGetter() (sessionGetter, error) {
	_, err := os.Stat(wtmpdbPath)
	if err == nil {
		wtmpdb := newWtmpdb()
		if err := wtmpdb.init(); err != nil {
			return nil, fmt.Errorf("failed to stat wtmpdb: %w", err)
		}

		return wtmpdb, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat wtmpdb file: %w", err)
	}

	return &utmpSessionGetter{}, nil
}
