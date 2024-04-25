package notify

import (
	"path"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

// GetIconPath checks if it is executed under snap and returns a full path to a requested icon.
// If it is not run under snap, it returns the given name
func GetIconPath(name string) string {
	if snapconf.IsUnderSnap() {
		const iconPath = "/usr/share/icons/hicolor/scalable/apps"
		return internal.PrefixStaticPath(path.Join(iconPath, name+".svg"))
	}

	return name
}
