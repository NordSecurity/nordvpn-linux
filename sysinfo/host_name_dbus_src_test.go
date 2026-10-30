package sysinfo

import (
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/sysinfo/dbusutil"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	dbusutil_mock "github.com/NordSecurity/nordvpn-linux/test/mock/sysinfo/dbusutil"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func Test_GetPropertyFromDBus(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		client    dbusutil.DBusObjectClient
		property  string
		want      string
		expectErr bool
	}{
		{
			name:     "Valid Property Retrieval",
			client:   dbusutil_mock.NewMockDbusObjectClient(nil, map[string]dbus.Variant{"TestProperty": dbus.MakeVariant("test-value")}, nil),
			property: "TestProperty",
			want:     "test-value",
		},
		{
			name:      "Property Not Found",
			client:    dbusutil_mock.NewMockDbusObjectClient(nil, map[string]dbus.Variant{}, nil),
			property:  "MissingProperty",
			want:      "",
			expectErr: true,
		},
		{
			name:      "DBus Client Error",
			client:    dbusutil_mock.NewMockDbusObjectClient(nil, nil, fmt.Errorf("DBus failure")),
			property:  "TestProperty",
			want:      "",
			expectErr: true,
		},
		{
			name:      "Empty Property Name",
			client:    dbusutil_mock.NewMockDbusObjectClient(nil, map[string]dbus.Variant{"ValidProperty": dbus.MakeVariant("valid-value")}, nil),
			property:  "",
			want:      "",
			expectErr: true,
		},
		{
			name:      "Different Data Type (int instead of string)",
			client:    dbusutil_mock.NewMockDbusObjectClient(nil, map[string]dbus.Variant{"NumericProperty": dbus.MakeVariant(42)}, nil),
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
			got, err := dbusutil.GetStringProperty(tt.client, tt.property)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
			if got != tt.want {
				t.Errorf("Expected result: %q, got: %q", tt.want, got)
			}
		})
	}
}

func Test_NewHostname1DBusPropertyClient(t *testing.T) {
	category.Set(t, category.Unit)

	client := dbusutil.NewObjectClient(
		nil,
		"org.freedesktop.hostname1",
	)
	assert.Nil(t, client, "must be nil")
}
