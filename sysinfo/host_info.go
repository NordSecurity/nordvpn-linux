package sysinfo

import "github.com/NordSecurity/nordvpn-linux/log"

type HostInfo struct {
	OSName       string
	OSPrettyName string
	Architecture string
	DeviceType   SystemDeviceType
}

type hostInfoSources struct {
	osName       func() (string, error)
	osPrettyName func() (string, error)
	architecture func() string
	deviceType   func() SystemDeviceType
}

func NewHostInfo() HostInfo {
	return newHostInfo(hostInfoSources{
		osName:       func() (string, error) { return readOSReleaseTag("NAME") },
		osPrettyName: getHostOSPrettyName,
		architecture: GetHostArchitecture,
		deviceType:   getDeviceType,
	})
}

func newHostInfo(src hostInfoSources) HostInfo {
	osName, err := src.osName()
	if err != nil {
		log.Sysinfo.Warnf("reading OS name from %s: %s", etcOSReleaseFile, err)
	}

	osPrettyName, err := src.osPrettyName()
	if err != nil || osPrettyName == "" {
		log.Sysinfo.Warnf(
			"retrieving OS pretty name: err '%v', name '%s', fallback to osName %s",
			err, osPrettyName, osName,
		)
		osPrettyName = osName
	}

	architecture := src.architecture()
	if architecture == "" {
		log.Sysinfo.Warn("host architecture is unknown")
	}

	deviceType := src.deviceType()
	if deviceType == SystemDeviceTypeUnknown {
		log.Sysinfo.Warn("device type is unknown")
	}

	return HostInfo{
		OSName:       osName,
		OSPrettyName: osPrettyName,
		Architecture: architecture,
		DeviceType:   deviceType,
	}
}
