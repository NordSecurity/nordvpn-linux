package sysinfo

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// DBusPropertyClient defines the interface for interacting with DBus properties.
type DBusPropertyClient interface {
	GetProperty(name string) (dbus.Variant, error)
}

// getPropertyFromDBus retrieves a generic property from any DBus client.
func getPropertyFromDBus(client DBusPropertyClient, property string) (string, error) {
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

// getOSPrettyName retrieves the OS pretty name using a DBusPropertyClient.
func getOSPrettyName(client DBusPropertyClient) (string, error) {
	return getPropertyFromDBus(client, "OperatingSystemPrettyName")
}

// genericDBusPropertyClient provides shared DBus functionality.
type genericDBusPropertyClient struct {
	conn       *dbus.Conn
	service    string
	objectPath dbus.ObjectPath
}

// GetProperty retrieves a specified DBus property.
func (c *genericDBusPropertyClient) GetProperty(name string) (dbus.Variant, error) {
	return c.conn.Object(c.service, c.objectPath).GetProperty(name)
}

// NewHostname1DBusPropertyClient initializes a DBus client for hostname1.
func NewHostname1DBusPropertyClient(conn *dbus.Conn) DBusPropertyClient {
	return &genericDBusPropertyClient{
		conn:       conn,
		service:    "org.freedesktop.hostname1",
		objectPath: "/org/freedesktop/hostname1",
	}
}
