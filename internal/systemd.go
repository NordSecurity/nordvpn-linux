package internal

import (
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

// IsServiceActive check if given service is active
func IsServiceActive(service string) bool {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Println(ErrorPrefix, "getting system dbus:", err)
		return false
	}
	defer conn.Close()

	// obtain systemd dbus object
	dbusObj, ok := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1").(*dbus.Object)
	if !ok {
		log.Println(ErrorPrefix, "obtaining systemd dbus object:", err)
		return false
	}

	if !strings.HasSuffix(service, ".service") {
		service = service + ".service"
	}

	var unitPath dbus.ObjectPath

	// obtain service unit path in dbus naming system
	err = dbusObj.Call("org.freedesktop.systemd1.Manager.GetUnit", 0, service).Store(&unitPath)
	if err != nil {
		log.Println(ErrorPrefix, "invoking dbus object method:", err)
		return false
	}

	// obtain service unit dbus object
	dbusObj, ok = conn.Object("org.freedesktop.systemd1", unitPath).(*dbus.Object)
	if !ok {
		log.Println(ErrorPrefix, "obtaining service unit dbus object:", err)
		return false
	}

	// get dbus object property
	propVal, err := dbusObj.GetProperty("org.freedesktop.systemd1.Unit.ActiveState")
	if err != nil {
		log.Println(ErrorPrefix, "obtaining service unit dbus object property:", err)
		return false
	}

	// property values is surrounded with ""
	return strings.ToLower(strings.Trim(propVal.String(), "\"")) == "active"
}

// IsSystemShutdown detect if system is being shutdown
func IsSystemShutdown() bool {
	// https://www.freedesktop.org/software/systemd/man/latest/shutdown.html
	return FileExists("/run/nologin")
}

// IsSystemd detect if system is running systemd
func IsSystemd() bool {
	// https://www.freedesktop.org/software/systemd/man/latest/sd_booted.html
	return FileExists("/run/systemd/system")
}
