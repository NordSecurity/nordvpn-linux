package sysinfo

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/godbus/dbus/v5"
)

const etcOSReleaseFile = "/etc/os-release"

// GetHostOSName retrieves the standard name of the currently running operating system.
func GetHostOSName() (string, error) {
	return readOSReleaseTag("NAME")
}

// HostOSPrettyName retrieves the human-readable OS name.
func GetHostOSPrettyNameFallback() (string, error) {
	return readOSReleaseTag("PRETTY_NAME")
}

// readOSReleaseTag opens the 'etcOSReleaseFile' file and retrieves the specified tag.
func readOSReleaseTag(tag string) (string, error) {
	file, err := os.Open(etcOSReleaseFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	return readTagFromOSRelease(file, tag)
}

// HostOSPrettyName retrieves user-friendly OS name using D-Bus communication with the system
// hostname service.
func GetHostOSPrettyName() string {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Println(internal.ErrorPrefix, "connecting to system dbus:", err)
		return ""
	}

	name, err := getOSPrettyName(NewHostname1DBusPropertyClient(conn))
	if err != nil {
		log.Println(internal.WarningPrefix, "retrieving OS pretty name:", err)
		return ""
	}

	return name
}

func readTagFromOSRelease(r io.Reader, tag string) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	for _, line := range bytes.Split(data, []byte("\n")) {
		key, value, ok := bytes.Cut(line, []byte("="))
		if !ok {
			continue
		}

		if string(key) == tag {
			return string(bytes.Trim(value, "\"")), nil
		}
	}

	return "", fmt.Errorf("unsupported tag: %v", tag)
}

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
