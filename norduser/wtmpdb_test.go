package norduser

import (
	"context"
	"errors"
	"testing"

	"github.com/godbus/dbus/v5"
)

// mockDBusConn implements dbusConn
type mockDBusConn struct {
	objects map[dbus.ObjectPath]dbus.BusObject
}

func (m *mockDBusConn) object(_ string, path dbus.ObjectPath) dbus.BusObject {
	if obj, ok := m.objects[path]; ok {
		return obj
	}
	return m.objects["*"] // fallback for any path
}

func (m *mockDBusConn) close() error { return nil }

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
	x11Path := dbus.ObjectPath("/org/freedesktop/login1/session/_31")
	waylandPath := dbus.ObjectPath("/org/freedesktop/login1/session/_32")
	otherPath := dbus.ObjectPath("/org/freedesktop/login1/session/_33")

	x11Obj := &mockBusObject{getPropFn: func(string) (dbus.Variant, error) {
		return dbus.MakeVariant("x11"), nil
	}}
	waylandObj := &mockBusObject{getPropFn: func(string) (dbus.Variant, error) {
		return dbus.MakeVariant("wayland"), nil
	}}
	ttyObj := &mockBusObject{getPropFn: func(string) (dbus.Variant, error) {
		return dbus.MakeVariant("tty"), nil
	}}
	failObj := &mockBusObject{getPropFn: func(string) (dbus.Variant, error) {
		return dbus.Variant{}, errors.New("no such property")
	}}

	tests := []struct {
		name           string
		sessions       [][]interface{}
		objects        map[dbus.ObjectPath]dbus.BusObject
		expectedResult bool
	}{
		{
			name:           "x11 session",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "alice", "seat0", x11Path)},
			objects:        map[dbus.ObjectPath]dbus.BusObject{x11Path: x11Obj},
			expectedResult: true,
		},
		{
			name:           "wayland session",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "alice", "seat0", waylandPath)},
			objects:        map[dbus.ObjectPath]dbus.BusObject{waylandPath: waylandObj},
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
			objects:        map[dbus.ObjectPath]dbus.BusObject{otherPath: ttyObj},
			expectedResult: false,
		},
		{
			name:           "different user",
			sessions:       [][]interface{}{sessionTuple("1", "1000", "bob", "seat0", x11Path)},
			objects:        map[dbus.ObjectPath]dbus.BusObject{x11Path: x11Obj},
			expectedResult: false,
		},
		{
			name:           "malformed short tuple",
			sessions:       [][]interface{}{{"1", "1000"}},
			expectedResult: false,
		},
		{
			name:           "malformed username",
			sessions:       [][]interface{}{{"1", "1000", 42, "seat0", x11Path}},
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
			objects:        map[dbus.ObjectPath]dbus.BusObject{x11Path: failObj},
			expectedResult: false,
		},
		{
			name: "second session is gui",
			sessions: [][]interface{}{
				sessionTuple("1", "1000", "alice", "seat0", otherPath),
				sessionTuple("2", "1000", "alice", "seat0", x11Path),
			},
			objects:        map[dbus.ObjectPath]dbus.BusObject{otherPath: ttyObj, x11Path: x11Obj},
			expectedResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manager := &mockBusObject{callFn: func(string, ...interface{}) *dbus.Call {
				return &dbus.Call{Body: []interface{}{test.sessions}}
			}}
			objects := map[dbus.ObjectPath]dbus.BusObject{"/org/freedesktop/login1": manager}
			for path, object := range test.objects {
				objects[path] = object
			}
			getter := &wtmpdbSessionGetter{dbusConn: &mockDBusConn{objects: objects}}

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
