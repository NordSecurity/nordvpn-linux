package request

import (
	"encoding/json"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/features"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/sysinfo"
)

const (
	AppName           = "NordApp Linux"
	nordAgentApp      = "NordVPN"
	nordAgentPlatform = "Linux"
	unknownDistroName = "unknown"
)

type nordAgent struct {
	App            string `json:"App"`
	Platform       string `json:"Platform"`
	AppVersion     string `json:"AppVersion"`
	OS             string `json:"OS,omitempty"`
	DeviceInfo     string `json:"DeviceInfo,omitempty"`
	AdditionalInfo string `json:"AdditionalInfo,omitempty"`
}

type ClientIdentity struct {
	version   string
	buildType features.TypeOfBuild
	hostInfo  sysinfo.HostInfo
}

func NewClientIdentity(
	version string,
	buildType features.TypeOfBuild,
	hostInfo sysinfo.HostInfo,
) ClientIdentity {
	return ClientIdentity{
		version:   version,
		buildType: buildType,
		hostInfo:  hostInfo,
	}
}

// UserAgent generates a User-Agent header value compliant with RFC 9110 section 10.1.5.
// It formats the user agent as "<application-name>/<version> (<platform-details>)" where:
// - application-name is application identifier
// - version is the application version
// - platform-details contains the distro name of the currently running kernel, or "unknown" if it is empty
//
// See: https://www.rfc-editor.org/rfc/rfc9110.html#section-10.1.5
func (c ClientIdentity) UserAgent() string {
	distroName := c.hostInfo.OSPrettyName
	if distroName == "" {
		distroName = unknownDistroName
	}
	return fmt.Sprintf("%s/%s (%s)", AppName, c.version, distroName)
}

func (c ClientIdentity) NordAgent() string {
	agent := nordAgent{
		App:            nordAgentApp,
		Platform:       nordAgentPlatform,
		AppVersion:     c.version,
		OS:             c.hostInfo.OSPrettyName,
		DeviceInfo:     c.hostInfo.Architecture,
		AdditionalInfo: string(c.buildType),
	}

	value, err := json.Marshal(agent)
	if err != nil {
		log.Error("failed to marshal nord agent header value", err)
		return ""
	}
	return string(value)
}
