//go:build (linux || freebsd || openbsd || netbsd) && !android

package systray

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/systray/internal/generated/notifier"
)

func TestPropSpecCoversIntrospection(t *testing.T) {
	served := instance.createPropSpec()["org.kde.StatusNotifierItem"]
	for _, p := range notifier.IntrospectDataStatusNotifierItem.Properties {
		if _, ok := served[p.Name]; !ok {
			t.Errorf("property %q is advertised via introspection but not served", p.Name)
		}
	}
}
