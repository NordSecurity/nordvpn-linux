package dbusutil

import (
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/godbus/dbus/v5"
)

const logTag = "[dbusutil]"

// DBusPropertyClient defines the interface for interacting with DBus properties.
type DBusPropertyClient interface {
	GetProperty(name string) (dbus.Variant, error)
}

// genericDBusPropertyClient provides shared DBus functionality.
type genericDBusPropertyClient struct {
	conn       *dbus.Conn
	service    string
	objectPath dbus.ObjectPath
}

// GetProperty retrieves a specified DBus property.
func (c *genericDBusPropertyClient) GetProperty(name string) (dbus.Variant, error) {
	return c.conn.Object(c.service, c.objectPath).GetProperty(c.service + "." + name)
}

// NewPropertyClient returns a generic D-Bus property client.
// This can be used for any service and object path.
func NewPropertyClient(conn *dbus.Conn, service string, objectPath dbus.ObjectPath) DBusPropertyClient {
	if conn == nil {
		log.Printf("%s %s D-Bus connection is nil; cannot create property client",
			logTag, internal.ErrorPrefix)
		return nil
	}

	return &genericDBusPropertyClient{
		conn:       conn,
		service:    service,
		objectPath: objectPath,
	}
}

// GetStringProperty retrieves a property from the DBus client.
func GetStringProperty(client DBusPropertyClient, property string) (string, error) {
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
