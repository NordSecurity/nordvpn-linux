package request

import (
	"fmt"
)

const AppName = "NordApp Linux"

// osPrettyNameGetter provides definition to a OS pretty name getter function.
type osPrettyNameGetter func() (string, error)

// GetUserAgentValue generates a User-Agent header value compliant with RFC 9110 section 10.1.5.
// It formats the user agent as "<application-name>/<version> (<platform-details>)" where:
// - application-name is application identifier
// - version is the application version
// - platform-details contains the distro name of the currently running kernel
//
// Returns:
//   - string: The formatted User-Agent string
//   - error: An error if distribution information cannot be retrieved
//
// See: https://www.rfc-editor.org/rfc/rfc9110.html#section-10.1.5
func GetUserAgentValue(version string, nameGetter osPrettyNameGetter) (string, error) {
	distro_name, err := nameGetter()
	if err != nil {
		return "", fmt.Errorf("determining device os: %w", err)
	}
	return fmt.Sprintf("%s/%s (%s)", AppName, version, distro_name), nil
}
