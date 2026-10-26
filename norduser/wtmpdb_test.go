package norduser

import (
	"context"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	dbusutil_mock "github.com/NordSecurity/nordvpn-linux/test/mock/sysinfo/dbusutil"

	"github.com/godbus/dbus/v5"
)

// mockBusObject implements dbus.BusObject
type mockBusObject struct {
	callFn       func(method string, args ...interface{}) *dbus.Call
	getPropFn    func(prop string) (dbus.Variant, error)
	getAllPropFn func(iface string) (map[string]dbus.Variant, error)
}

func (m *mockBusObject) Call(method string, _ dbus.Flags, args ...interface{}) *dbus.Call {
	return m.callFn(method, args...)
}

func (m *mockBusObject) CallWithContext(ctx context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return m.Call(method, flags, args...)
}

func (m *mockBusObject) Go(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	call := m.Call(method, flags, args...)
	ch <- call
	return call
}

func (m *mockBusObject) GoWithContext(ctx context.Context, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return m.Go(method, flags, ch, args...)
}

func (m *mockBusObject) AddMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}

func (m *mockBusObject) RemoveMatchSignal(iface, member string, options ...dbus.MatchOption) *dbus.Call {
	return &dbus.Call{}
}

func (m *mockBusObject) GetProperty(prop string) (dbus.Variant, error) {
	return m.getPropFn(prop)
}

func (m *mockBusObject) GetAllProperties(iface string) (map[string]dbus.Variant, error) {
	return m.getAllPropFn(iface)
}

func (m *mockBusObject) StoreProperty(p string, value any) error { return nil }

func (m *mockBusObject) SetProperty(_ string, _ interface{}) error { return nil }

func (m *mockBusObject) Destination() string { return "org.freedesktop.login1" }

func (m *mockBusObject) Path() dbus.ObjectPath { return "/" }

func sessionTuple(id, uid, user, seat string, path dbus.ObjectPath) []interface{} {
	return []interface{}{id, uid, user, seat, path}
}

func TestHasGUISession(t *testing.T) {
	category.Set(t, category.Unit)

	x11Path := dbus.ObjectPath("/org/freedesktop/login1/session/_31")
	waylandPath := dbus.ObjectPath("/org/freedesktop/login1/session/_32")
	otherPath := dbus.ObjectPath("/org/freedesktop/login1/session/_33")

	x11Prop := dbus.MakeVariant("x11")
	waylandProp := dbus.MakeVariant("wayland")
	ttyProp := dbus.MakeVariant("tty")
	failProp := dbus.Variant{}

	tests := []struct {
		name           string
		sessions       [][]interface{}
		properties     map[dbus.ObjectPath]dbus.Variant
		expectedResult bool
	}{
		{
			name:           "x11 session",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "alice", "seat0", x11Path)},
			properties:     map[dbus.ObjectPath]dbus.Variant{x11Path: x11Prop},
			expectedResult: true,
		},
		{
			name:           "wayland session",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "alice", "seat0", waylandPath)},
			properties:     map[dbus.ObjectPath]dbus.Variant{waylandPath: waylandProp},
			expectedResult: true,
		},
		{
			name:           "no sessions",
			sessions:       nil,
			expectedResult: false,
		},
		{
			name:           "tty session only",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "alice", "seat0", otherPath)},
			properties:     map[dbus.ObjectPath]dbus.Variant{otherPath: ttyProp},
			expectedResult: false,
		},
		{
			name:           "different user",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "bob", "seat0", x11Path)},
			properties:     map[dbus.ObjectPath]dbus.Variant{x11Path: x11Prop},
			expectedResult: false,
		},
		{
			name:           "malformed short tuple",
			sessions:       [][]interface{}{{"1", "1000"}},
			expectedResult: false,
		},
		{
			name:           "malformed username",
			sessions:       [][]interface{}{{"1", "1000", 42, "seat0", dbus.ObjectPath(x11Path)}},
			expectedResult: false,
		},
		{
			name:           "malformed path",
			sessions:       [][]interface{}{{"1", "1000", "alice", "seat0", "not-a-path"}},
			expectedResult: false,
		},
		{
			name:           "get property fails",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "alice", "seat0", x11Path)},
			properties:     map[dbus.ObjectPath]dbus.Variant{x11Path: failProp},
			expectedResult: false,
		},
		{
			name: "second session is gui",
			sessions: [][]interface{}{
				sessionTuple("1", "1000", "alice", "seat0", dbus.ObjectPath(otherPath)),
				sessionTuple("2", "1000", "alice", "seat0", dbus.ObjectPath(x11Path)),
			},
			properties:     map[dbus.ObjectPath]dbus.Variant{otherPath: ttyProp, x11Path: x11Prop},
			expectedResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := &mockBusObject{
				callFn: func(string, ...interface{}) *dbus.Call {
					return &dbus.Call{Body: []interface{}{test.sessions}}
				},
			}

			objects := make(map[dbus.ObjectPath]dbus.BusObject)

			objects["/org/freedesktop/login1"] = manager
			for path, object := range test.properties {
				objects[path] = &mockBusObject{
					getPropFn: func(prop string) (dbus.Variant, error) {
						return object, nil
					},
				}
			}

			dbusMock := dbusutil_mock.NewMockDbusObjectClient(objects, map[string]dbus.Variant{}, nil)
			dbusMock.Objects = objects
			getter := &wtmpdbSessionGetter{dbusLoginClient: dbusMock}

			got, err := getter.hasGUISession("alice")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.expectedResult {
				t.Fatalf("got %v, want %v", got, test.expectedResult)
			}
		})
	}
}
