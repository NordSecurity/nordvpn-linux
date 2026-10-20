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
		log.Sysinfo.Warnf("retrieving OS pretty name: %v. Falling back to OS name: %s", err, osName)
		osPrettyName = osName
	}

	return HostInfo{
		OSName:       osName,
		OSPrettyName: osPrettyName,
		Architecture: src.architecture(),
		DeviceType:   src.deviceType(),
	}
}
