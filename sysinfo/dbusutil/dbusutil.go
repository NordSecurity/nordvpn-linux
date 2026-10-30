package dbusutil

import (
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/godbus/dbus/v5"
)

// DBusObjectClient defines the interface for interacting with DBus objects.
type DBusObjectClient interface {
	GetProperty(name string) (dbus.Variant, error)
	GetObject(objectPath dbus.ObjectPath) dbus.BusObject
}

// genericDBusObjectClient provides shared DBus functionality.
type genericDBusObjectClient struct {
	conn    *dbus.Conn
	service string
}

func (c *genericDBusObjectClient) GetObject(objectPath dbus.ObjectPath) dbus.BusObject {
	return c.conn.Object(c.service, objectPath)
}

// GetProperty retrieves a specified DBus property.
func (c *genericDBusObjectClient) GetProperty(name string) (dbus.Variant, error) {
	return c.GetObject(dbus.ObjectPath(name)).GetProperty(c.service + "." + name)
}

// NewObjectClient returns a generic D-Bus object client.
// This can be used for any service and object path.
func NewObjectClient(conn *dbus.Conn, service string) DBusObjectClient {
	if conn == nil {
		log.DBusUtil.Errorf("D-Bus connection is nil; cannot create property client")
		return nil
	}

	return &genericDBusObjectClient{
		conn:    conn,
		service: service,
	}
}

// GetStringProperty retrieves a property from the DBus client.
func GetStringProperty(client DBusObjectClient, property string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("invalid DBus client")
	}

	out, err := client.GetProperty(property)
	if err != nil {
		return "", fmt.Errorf("failed to get DBus property %q: %w", property, err)
	}

	result, ok := out.Value().(string)
	if !ok {
		return "", fmt.Errorf("unexpected type for DBus property %q", property)
	}

	return result, nil
}
