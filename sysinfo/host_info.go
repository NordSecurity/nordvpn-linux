package sysinfo

import "github.com/NordSecurity/nordvpn-linux/log"

type HostInfo struct {
	OSName       string
	OSPrettyName string
	Architecture string
	DeviceType   SystemDeviceType
}

type hostInfoSources struct {
	osRelease    func() (osRelease, error)
	osPrettyName func() (string, error)
	architecture func() string
	deviceType   func() SystemDeviceType
}

func NewHostInfo() HostInfo {
	return newHostInfo(hostInfoSources{
		osRelease:    readOSRelease,
		osPrettyName: getHostOSPrettyName,
		architecture: GetHostArchitecture,
		deviceType:   getDeviceType,
	})
}

func newHostInfo(src hostInfoSources) HostInfo {
	osRelease, err := src.osRelease()
	if err != nil {
		log.Sysinfo.Warnf("reading %s: %s", etcOSReleaseFile, err)
	}

	osName, err := osRelease.tag("NAME")
	if err != nil {
		log.Sysinfo.Warnf("reading OS name from %s: %s", etcOSReleaseFile, err)
	}

	osPrettyName, err := src.osPrettyName()
	if err != nil || osPrettyName == "" {
		log.Sysinfo.Warnf(
			"retrieving OS pretty name: err '%v', name '%s', fallback to %s",
			err, osPrettyName, etcOSReleaseFile,
		)
		osPrettyName, err = osRelease.tag("PRETTY_NAME")
	}
	if err != nil || osPrettyName == "" {
		log.Sysinfo.Warnf(
			"reading OS pretty name from %s: err '%v', name '%s', fallback to osName %s",
			etcOSReleaseFile, err, osPrettyName, osName,
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
