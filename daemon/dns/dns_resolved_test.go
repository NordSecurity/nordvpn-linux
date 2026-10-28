package dns

import (
	"net"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/require"
)

type dbusMethod struct {
	name  string
	flags dbus.Flags
	args  []interface{}
}

func newDbusMethod(member string, flags dbus.Flags, args ...any) dbusMethod {
	return dbusMethod{
		name:  member,
		flags: flags,
		args:  args,
	}
}

// Implements the dbus.Object interface, but only for the required function
type dbusObject struct {
	calls []dbusMethod
	dbus.BusObject
}

func (d *dbusObject) Call(method string, flags dbus.Flags, args ...any) *dbus.Call {
	d.calls = append(d.calls, newDbusMethod(method, flags, args...))
	return &dbus.Call{Err: nil}
}

func TestDbusCallsForSystemd(t *testing.T) {
	category.Set(t, category.Unit)

	ifIndex := 100
	iface := net.Interface{
		Index: ifIndex,
		Name:  "test",
	}
	dbusObj := &dbusObject{}
	resolver := Resolved{}
	resolver.configure(dbusObj, &iface, []string{"1.2.3.4"})

	expectedAtConfig := []dbusMethod{
		{
			name:  "org.freedesktop.resolve1.Manager.SetLinkDNS",
			flags: 0x0,
			args:  []interface{}{ifIndex, []dnsServer{{Family: 2, Address: []uint8{0x1, 0x2, 0x3, 0x4}}}},
		},

		{
			name:  "org.freedesktop.resolve1.Manager.SetLinkDomains",
			flags: 0x0,
			args:  []interface{}{ifIndex, []linkDomain{{Domain: ".", RoutingOnly: true}}},
		},

		{
			name:  "org.freedesktop.resolve1.Manager.SetLinkDNSSEC",
			flags: 0x0,
			args:  []interface{}{ifIndex, ""},
		},

		{
			name:  "org.freedesktop.resolve1.Manager.FlushCaches",
			flags: 0x0,
			args:  []interface{}(nil),
		},
	}

	require.Equal(t, dbusObj.calls, expectedAtConfig)

	dbusObj.calls = nil

	resolver.unset(dbusObj, &iface)

	expectedUnset := []dbusMethod{
		{
			name:  "org.freedesktop.resolve1.Manager.RevertLink",
			flags: 0x0,
			args:  []interface{}{ifIndex},
		},
		{
			name:  "org.freedesktop.resolve1.Manager.FlushCaches",
			flags: 0x0,
			args:  []interface{}(nil),
		},
	}

	require.Equal(t, dbusObj.calls, expectedUnset)
}
