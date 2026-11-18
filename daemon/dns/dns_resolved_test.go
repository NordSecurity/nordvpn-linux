package dns

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/require"
)

type dbusMethod struct {
	name  string
	flags dbus.Flags
	args  []any
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

func (d *dbusObject) CallWithContext(ctx context.Context, method string, flags dbus.Flags, args ...any) *dbus.Call {
	if ctx.Err() != nil {
		return &dbus.Call{Err: ctx.Err()}
	}
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
	err := resolver.configure(context.Background(), dbusObj, &iface, []string{"1.2.3.4"})
	require.NoError(t, err)

	expectedAtConfig := []dbusMethod{
		{
			name:  "org.freedesktop.resolve1.Manager.SetLinkDNS",
			flags: 0x0,
			args:  []any{ifIndex, []dnsServer{{Family: 2, Address: []uint8{0x1, 0x2, 0x3, 0x4}}}},
		},

		{
			name:  "org.freedesktop.resolve1.Manager.SetLinkDomains",
			flags: 0x0,
			args:  []any{ifIndex, []linkDomain{{Domain: ".", RoutingOnly: true}}},
		},

		{
			name:  "org.freedesktop.resolve1.Manager.SetLinkDNSSEC",
			flags: 0x0,
			args:  []any{ifIndex, ""},
		},

		{
			name:  "org.freedesktop.resolve1.Manager.FlushCaches",
			flags: 0x0,
			args:  []any(nil),
		},
	}

	require.Equal(t, dbusObj.calls, expectedAtConfig)

	dbusObj.calls = nil
	err = resolver.unset(context.Background(), dbusObj, &iface)
	require.NoError(t, err)

	expectedUnset := []dbusMethod{
		{
			name:  "org.freedesktop.resolve1.Manager.RevertLink",
			flags: 0x0,
			args:  []any{ifIndex},
		},
		{
			name:  "org.freedesktop.resolve1.Manager.FlushCaches",
			flags: 0x0,
			args:  []any(nil),
		},
	}

	require.Equal(t, dbusObj.calls, expectedUnset)
}

func TestSystemdResolverTimeout(t *testing.T) {
	category.Set(t, category.Unit)

	ifIndex := 100
	iface := net.Interface{
		Index: ifIndex,
		Name:  "test",
	}
	dbusObj := &dbusObject{}
	resolver := Resolved{}
	ctx, cancelFn := context.WithTimeout(context.Background(), time.Microsecond*1)
	time.Sleep(time.Millisecond * 1) // give time for the context to expire

	errConfigure := resolver.configure(ctx, dbusObj, &iface, []string{"1.2.3.4"})
	require.Error(t, errConfigure)

	errUnset := resolver.unset(ctx, dbusObj, &iface)
	require.Error(t, errUnset)
	cancelFn()
}
