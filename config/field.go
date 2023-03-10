package config

import "encoding/json"

// Field will unmarshal to null if unset.
type Field[T any] struct{ value *T }

// Set the inner value.
func (f *Field[T]) Set(value T) { f.value = &value }

// Get the inner value.
func (f Field[T]) Get() T {
	if f.value != nil {
		return *f.value
	}

	var ret T
	return ret
}

// MarshalJSON has to be a value receiver or else nil f.value will be marshaled as {}.
func (f Field[T]) MarshalJSON() ([]byte, error) { return json.Marshal(f.value) }

// UnmarshalJSON has to be a pointer receiver or else f.value will not update.
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var tmp T
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	f.value = &tmp
	return nil
}

// TrueField is a boolean, which is true by default.
type TrueField struct{ Field[bool] }

func (t TrueField) Get() bool {
	if t.value != nil {
		return *t.value
	}
	return true
}
