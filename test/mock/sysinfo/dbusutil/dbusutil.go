package dbusutil

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

type MockDbusObjectClient struct {
	GetPropertyErr error
	Properties     map[string]dbus.Variant
	Objects        map[dbus.ObjectPath]dbus.BusObject
}

func NewMockDbusObjectClient(objects map[dbus.ObjectPath]dbus.BusObject,
	properties map[string]dbus.Variant,
	getPropertyErr error) *MockDbusObjectClient {
	return &MockDbusObjectClient{
		Properties:     properties,
		GetPropertyErr: getPropertyErr,
	}
}

func (m *MockDbusObjectClient) GetObject(name dbus.ObjectPath) dbus.BusObject {
	return m.Objects[name]
}

func (m *MockDbusObjectClient) GetProperty(name string) (dbus.Variant, error) {
	if m.GetPropertyErr != nil {
		return dbus.Variant{}, m.GetPropertyErr
	}

	val, exists := m.Properties[name]
	if !exists {
		return dbus.Variant{}, fmt.Errorf("property %q not found", name)
	}

	return val, nil
}
