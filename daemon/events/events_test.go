package events

import (
	"math"
	"reflect"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestNewDaemonSubjects(t *testing.T) {
	category.Set(t, category.Unit)
	valid, _ := isValid(NewEventsEmpty())
	assert.True(t, valid)
}

// isValid returns true if given val is not nil. In case val is struct,
// it checks if any of exported fields are not nil
// Also, returns the minimum number of a private map field elements
func isValid(val interface{}) (bool, int) {
	return isValidGetMin(val, math.MaxInt32)
}

func isValidGetMin(val interface{}, min int) (bool, int) {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, -1
		}
		v = reflect.ValueOf(v.Elem().Interface())
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			subMin := min
			if !isPrivate(field) {
				var valid bool
				valid, subMin = isValidGetMin(field.Interface(), min)
				if !valid {
					return false, -1
				}
			} else if field.Kind() == reflect.Slice {
				subMin = field.Len()
			}
			if subMin < min {
				min = subMin
			}
		}
	}
	return true, min
}

func isPrivate(val reflect.Value) (private bool) {
	defer func() {
		if err := recover(); err != nil {
			private = true
		}
	}()
	val.Interface()
	return
}
