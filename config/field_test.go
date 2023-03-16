package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

// Get the inner value.
func (f Field[T]) Get() T {
	if f.value != nil {
		return *f.value
	}

	var ret T
	return ret
}

// used to take a pointer to primitive types, such as bool
func pointer[T any](t *testing.T, to T) *T {
	t.Helper()
	return &to
}

func testUnsetFieldMarshalsToNull[T any](t *testing.T) {
	t.Run("unset field marshals to null", func(t *testing.T) {
		var field Field[T]
		actual, err := json.Marshal(&field)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Equal(t, []byte("null"), actual)
	})
}

func testNullUnmarshalsToUnsetField[T any](t *testing.T) {
	t.Run("null unmarshals to unset field", func(t *testing.T) {
		var (
			field    Field[T]
			expected T
		)
		err := json.Unmarshal([]byte("null"), &field)
		assert.NoError(t, err)
		assert.Nil(t, field.value)
		assert.Equal(t, expected, field.Get())
	})
}

func testSetFieldMarshalsToItsValue[T any](t *testing.T, value T, expected string) {
	t.Run("set field marshals to its value", func(t *testing.T) {
		field := Field[T]{value: pointer(t, value)}
		actual, err := json.Marshal(&field)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Equal(t, []byte(expected), actual)
	})
}

func testValueUnmarshalsToSetField[T any](t *testing.T, expected T, value string) {
	t.Run("value unmarshals to set field", func(t *testing.T) {
		var field Field[T]
		err := json.Unmarshal([]byte(value), &field)
		assert.NoError(t, err)
		assert.NotNil(t, field.value)
		assert.Equal(t, expected, field.Get())
	})
}

type testStruct[T any] struct {
	Visible   T        `json:"visible"`
	Invisible Field[T] `json:"invisible"`
}

func testStructField[T any](t *testing.T, zero T) {
	t.Run("unset struct field marshals to null", func(t *testing.T) {
		actual, err := json.Marshal(&testStruct[T]{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		expected := fmt.Sprintf(`{ "visible": %#v, "invisible": null }`, zero)
		assert.JSONEq(t, expected, string(actual))
	})

	t.Run("set struct field marshals to its value", func(t *testing.T) {
		actual, err := json.Marshal(&testStruct[T]{
			Invisible: Field[T]{value: pointer(t, zero)},
		})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		expected := fmt.Sprintf(`{ "visible": %#v, "invisible": %#v }`, zero, zero)
		assert.JSONEq(t, expected, string(actual))
	})

	t.Run("new struct field unmarshals to nil", func(t *testing.T) {
		var test struct {
			Old Field[T] `json:"old"`
			New Field[T] `json:"new"`
		}
		data := fmt.Sprintf(`{"old": %#v}`, zero)
		err := json.Unmarshal([]byte(data), &test)
		assert.NoError(t, err)
		assert.NotNil(t, test.Old.value)
		assert.Nil(t, test.New.value)
	})
}

func TestFieldBool(t *testing.T) {
	category.Set(t, category.Unit)

	testUnsetFieldMarshalsToNull[bool](t)
	testSetFieldMarshalsToItsValue(t, true, "true")
	testNullUnmarshalsToUnsetField[bool](t)
	testValueUnmarshalsToSetField(t, true, "true")
	testStructField(t, false)
}

func TestFieldInt(t *testing.T) {
	category.Set(t, category.Unit)

	testUnsetFieldMarshalsToNull[int](t)
	testSetFieldMarshalsToItsValue(t, 1337, "1337")
	testNullUnmarshalsToUnsetField[int](t)
	testValueUnmarshalsToSetField(t, 1337, "1337")
	testStructField(t, 0)
}

func TestFieldString(t *testing.T) {
	category.Set(t, category.Unit)

	testUnsetFieldMarshalsToNull[string](t)
	testSetFieldMarshalsToItsValue(t, "1337", `"1337"`)
	testNullUnmarshalsToUnsetField[string](t)
	testValueUnmarshalsToSetField(t, "1337", `"1337"`)
	testStructField(t, "")
}

func TestTrueField(t *testing.T) {
	category.Set(t, category.Unit)

	testUnsetFieldMarshalsToNull[TrueField](t)
	testSetFieldMarshalsToItsValue(t,
		TrueField{Field: Field[bool]{value: pointer(t, false)}},
		"false",
	)
	testNullUnmarshalsToUnsetField[TrueField](t)
	testValueUnmarshalsToSetField(t,
		TrueField{Field: Field[bool]{value: pointer(t, true)}},
		"true",
	)
}
