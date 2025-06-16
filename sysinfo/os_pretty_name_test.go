package sysinfo

import (
	"fmt"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

type mockDBusPropertyClient struct {
	props map[string]dbus.Variant
	err   error
}

func (m *mockDBusPropertyClient) GetProperty(name string) (dbus.Variant, error) {
	if m.err != nil {
		return dbus.Variant{}, m.err
	}

	val, exists := m.props[name]
	if !exists {
		return dbus.Variant{}, fmt.Errorf("property %q not found", name)
	}

	return val, nil
}

func newMockDBusPropertyClient(props map[string]dbus.Variant, err error) DBusPropertyClient {
	return &mockDBusPropertyClient{
		props: props,
		err:   err,
	}
}

func Test_GetPropertyFromDBus(t *testing.T) {
	tests := []struct {
		name      string
		client    DBusPropertyClient
		property  string
		want      string
		expectErr bool
	}{
		{
			name:     "Valid Property Retrieval",
			client:   newMockDBusPropertyClient(map[string]dbus.Variant{"TestProperty": dbus.MakeVariant("test-value")}, nil),
			property: "TestProperty",
			want:     "test-value",
		},
		{
			name:      "Property Not Found",
			client:    newMockDBusPropertyClient(map[string]dbus.Variant{}, nil),
			property:  "MissingProperty",
			want:      "",
			expectErr: true,
		},
		{
			name:      "DBus Client Error",
			client:    newMockDBusPropertyClient(nil, fmt.Errorf("DBus failure")),
			property:  "TestProperty",
			want:      "",
			expectErr: true,
		},
		{
			name:      "Empty Property Name",
			client:    newMockDBusPropertyClient(map[string]dbus.Variant{"ValidProperty": dbus.MakeVariant("valid-value")}, nil),
			property:  "",
			want:      "",
			expectErr: true,
		},
		{
			name:      "Different Data Type (int instead of string)",
			client:    newMockDBusPropertyClient(map[string]dbus.Variant{"NumericProperty": dbus.MakeVariant(42)}, nil),
			property:  "NumericProperty",
			want:      "",
			expectErr: true,
		},
		{
			name:      "Nil Client",
			client:    nil,
			property:  "SomeProperty",
			want:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPropertyFromDBus(tt.client, tt.property)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
			if got != tt.want {
				t.Errorf("Expected result: %q, got: %q", tt.want, got)
			}
		})
	}
}

func Test_GetOSPrettyName(t *testing.T) {
	mockClient := newMockDBusPropertyClient(
		map[string]dbus.Variant{"OperatingSystemPrettyName": dbus.MakeVariant("FancyOS 1.2.3 LTS")},
		nil)

	got, err := getOSPrettyName(mockClient)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	want := "FancyOS 1.2.3 LTS"
	if got != want {
		t.Errorf("Expected: %q, got: %q", want, got)
	}
}

func TestNewHostname1DBusPropertyClient(t *testing.T) {
	client := NewHostname1DBusPropertyClient(&dbus.Conn{})
	assert.NotNil(t, client, "must not be nil")

	hostname1Client, ok := client.(*genericDBusPropertyClient)
	assert.True(t, ok, "interface should be convertible to generic dbus client")
	assert.Equal(t, hostname1Client.service, "org.freedesktop.hostname1")
	assert.Equal(t, hostname1Client.objectPath, dbus.ObjectPath("/org/freedesktop/hostname1"))
}
