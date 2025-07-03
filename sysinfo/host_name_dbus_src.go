package sysinfo

import (
	"log"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/sysinfo/dbusutil"
	"github.com/godbus/dbus/v5"
)

// all available readonly properties from hostname1
// ref. https://www.freedesktop.org/software/systemd/man/latest/org.freedesktop.hostname1.html
const (
	dbusHostname1PropHostname                  string = "Hostname"
	dbusHostname1PropStaticHostname            string = "StaticHostname"
	dbusHostname1PropPrettyHostname            string = "PrettyHostname"
	dbusHostname1PropDefaultHostname           string = "DefaultHostname"
	dbusHostname1PropHostnameSource            string = "HostnameSource"
	dbusHostname1PropIconName                  string = "IconName"
	dbusHostname1PropChassis                   string = "Chassis"
	dbusHostname1PropDeployment                string = "Deployment"
	dbusHostname1PropLocation                  string = "Location"
	dbusHostname1PropKernelName                string = "KernelName"
	dbusHostname1PropKernelRelease             string = "KernelRelease"
	dbusHostname1PropKernelVersion             string = "KernelVersion"
	dbusHostname1PropOperatingSystemPrettyName string = "OperatingSystemPrettyName"
	dbusHostname1PropOperatingSystemCPEName    string = "OperatingSystemCPEName"
	dbusHostname1PropHomeURL                   string = "HomeURL"
	dbusHostname1PropHardwareVendor            string = "HardwareVendor"
	dbusHostname1PropHardwareModel             string = "HardwareModel"
	dbusHostname1PropFirmwareVersion           string = "FirmwareVersion"
	dbusHostname1PropFirmwareVendor            string = "FirmwareVendor"
)

type HostOsPrettyName interface {
	GetHostOSPrettyName() (string, error)
}
type HostOsPrettyNameImpl struct{}

func NewHostOsPrettyName() HostOsPrettyName {
	return HostOsPrettyNameImpl{}
}

// GetHostOSPrettyName retrieves user-friendly OS name using D-Bus communication with the system
// hostname service.
func (HostOsPrettyNameImpl) GetHostOSPrettyName() (string, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Printf("%s %s connecting to system dbus: %s", logTag, internal.ErrorPrefix, err)
		return "", err
	}
	defer conn.Close()

	client := dbusutil.NewPropertyClient(
		conn,
		"org.freedesktop.hostname1",
		"/org/freedesktop/hostname1",
	)

	name, err := dbusutil.GetStringProperty(client, dbusHostname1PropOperatingSystemPrettyName)
	if err != nil {
		log.Printf("%s %s retrieving OS pretty name: %s", logTag, internal.WarningPrefix, err)
		return "", err
	}

	return name, nil
}
