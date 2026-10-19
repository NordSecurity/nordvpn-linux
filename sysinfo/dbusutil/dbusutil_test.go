package dbusutil_test

import (
	"errors"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/sysinfo/dbusutil"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	dbusutil_mock "github.com/NordSecurity/nordvpn-linux/test/mock/sysinfo/dbusutil"
	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func Test_GetStringProperty_Success(t *testing.T) {
	category.Set(t, category.Unit)

	client := dbusutil_mock.NewMockDbusObjectClient(nil,
		map[string]dbus.Variant{"SomeProp": dbus.MakeVariant("hello")},
		nil)

	result, err := dbusutil.GetStringProperty(client, "SomeProp")
	assert.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func Test_GetStringProperty_InvalidType(t *testing.T) {
	category.Set(t, category.Unit)

	client := dbusutil_mock.NewMockDbusObjectClient(nil,
		map[string]dbus.Variant{"SomeProp": dbus.MakeVariant(1234)},
		nil)

	result, err := dbusutil.GetStringProperty(client, "SomeProp")
	assert.Error(t, err)
	assert.Empty(t, result)
}

func Test_GetStringProperty_ErrorFromClient(t *testing.T) {
	category.Set(t, category.Unit)

	client := dbusutil_mock.NewMockDbusObjectClient(nil,
		map[string]dbus.Variant{},
		errors.New("DBus failure"))

	result, err := dbusutil.GetStringProperty(client, "SomeProp")
	assert.Error(t, err)
	assert.Empty(t, result)
}

func Test_GetStringProperty_NilClient(t *testing.T) {
	category.Set(t, category.Unit)

	result, err := dbusutil.GetStringProperty(nil, "AnyProp")
	assert.Error(t, err)
	assert.Empty(t, result)
}

func Test_NewPropertyClient_NilClient(t *testing.T) {
	category.Set(t, category.Unit)

	client := dbusutil.NewObjectClient(nil, "Service", "AnyProp")
	assert.Nil(t, client)
}

func Test_NewPropertyClient_DummyClient(t *testing.T) {
	category.Set(t, category.Unit)

	client := dbusutil.NewObjectClient(&dbus.Conn{}, "Service", "AnyProp")
	assert.NotNil(t, client)
}
