package norduser

import (
	"database/sql"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/sysinfo/dbusutil"
	"github.com/godbus/dbus/v5"
	_ "github.com/mattn/go-sqlite3"
)

const wtmpdbPath = "/var/lib/wtmpdb/wtmp.db"

type wtmpdbSessionGetter struct {
	database        *sql.DB
	dbusConn        *dbus.Conn
	dbusLoginClient dbusutil.DBusObjectClient
}

func newWtmpdb() *wtmpdbSessionGetter {
	return &wtmpdbSessionGetter{}
}

func (w *wtmpdbSessionGetter) init() error {
	db, err := sql.Open("sqlite3", wtmpdbPath)
	if err != nil {
		return err
	}
	w.database = db

	dbusConn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.ProcessMonitor.Warn("failed to create dbus connection for wtmpdb monitor:", err)
		w.dbusConn = nil
		w.dbusLoginClient = nil
	} else {
		w.dbusConn = dbusConn
		w.dbusLoginClient = dbusutil.NewObjectClient(dbusConn, "/org/freedesktop/login1", "")
	}

	return nil
}

func (w *wtmpdbSessionGetter) getActiveUsers() (userData, error) {
	if w.database == nil {
		return userData{}, fmt.Errorf("database is not initialized")
	}

	const query = `SELECT DISTINCT User FROM wtmp
	          WHERE Type = 3 AND Logout IS NULL
	          ORDER BY ID`
	rows, err := w.database.Query(query)
	if err != nil {
		return userData{}, fmt.Errorf("querying wtmpdb: %w", err)
	}

	defer rows.Close()

	users := make(userData)

	for rows.Next() {
		var user string
		if err := rows.Scan(&user); err != nil {
			return userData{}, fmt.Errorf("scanning wtmpdb row: %w", err)
		}

		users[user] = loginText
		if hasGUISession, err := w.hasGUISession(user); err != nil {
			log.ProcessMonitor.Warn("failed to determine if user has GUI session:", err)
		} else if hasGUISession {
			users[user] = loginGUI
		}
	}

	if err := rows.Err(); err != nil {
		return userData{}, fmt.Errorf("reading wtmpdb row: %w", err)
	}

	return users, nil
}

func (w *wtmpdbSessionGetter) hasGUISession(username string) (bool, error) {
	if w.dbusLoginClient == nil {
		return false, nil
	}

	loginManager := w.dbusLoginClient.GetObject("/org/freedesktop/login1")

	var sessions [][]interface{}
	if err := loginManager.Call("org.freedesktop.login1.Manager.ListSessions", 0).Store(&sessions); err != nil {
		return false, fmt.Errorf("calling login manager: %w", err)
	}

	for _, session := range sessions {
		if len(session) < 5 {
			log.ProcessMonitor.Warn("malformed dbus tuple, invalid length")
			continue
		}

		user, ok := session[2].(string)
		if !ok {
			log.ProcessMonitor.Warn("malformed dbus tuple, unexpected type of username")
			continue
		}

		if user != username {
			continue
		}

		path, ok := session[4].(dbus.ObjectPath)
		if !ok {
			log.ProcessMonitor.Warn("malformed dbus tuple, unexpected type of object path")
			continue
		}

		sessionType, err := dbusutil.GetStringProperty(w.dbusLoginClient, string(path))
		if err != nil {
			log.ProcessMonitor.Warn("failed to get session type:", err)
			continue
		}

		if sessionType == "x11" || sessionType == "wayland" {
			return true, nil
		}
	}

	return false, nil
}

func (w *wtmpdbSessionGetter) getDatabasePath() string {
	return wtmpdbPath
}

func (w *wtmpdbSessionGetter) close() {
	if w.database != nil {
		if err := w.database.Close(); err != nil {
			log.ProcessMonitor.Warn("failed to close database connection:", err)
		}
	}

	if w.dbusConn != nil {
		if err := w.dbusConn.Close(); err != nil {
			log.ProcessMonitor.Warn("failed to close dbus connection:", err)
		}
	}
}
