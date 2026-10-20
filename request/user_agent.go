package request

import (
	"fmt"
)

const AppName = "NordApp Linux"

// GetUserAgentValue generates a User-Agent header value compliant with RFC 9110 section 10.1.5.
// It formats the user agent as "<application-name>/<version> (<platform-details>)" where:
// - application-name is application identifier
// - version is the application version
// - platform-details contains the distro name of the currently running kernel, or "unknown" if it is empty
//
// See: https://www.rfc-editor.org/rfc/rfc9110.html#section-10.1.5
func GetUserAgentValue(version string, distroName string) string {
	if distroName == "" {
		distroName = "unknown"
	}
	return fmt.Sprintf("%s/%s (%s)", AppName, version, distroName)
}
