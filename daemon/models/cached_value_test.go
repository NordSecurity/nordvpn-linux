package models_test

import (
	"errors"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/daemon/models"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestCachedValue(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name         string
		initialValue int
		err          error
		validity     time.Duration
		updatedValue int
		shouldUpdate bool
	}{
		{
			name:         "works for initial value and update not called",
			initialValue: 1,
			err:          nil,
			validity:     1 * time.Minute,
			shouldUpdate: false,
		},
		{
			name:         "update callback function works",
			initialValue: 2,
			err:          nil,
			validity:     0 * time.Second,
			shouldUpdate: true,
			updatedValue: 10,
		},
		{
			name:         "initial value is error and update fixes it",
			initialValue: 0,
			err:          errors.New("failed"),
			validity:     0 * time.Second,
			shouldUpdate: true,
			updatedValue: 10,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updateCalled := false
			item := models.NewCachedValue(
				test.initialValue,
				test.err, time.Now(),
				test.validity,
				func(self *models.CachedValue[int]) {
					updateCalled = true
					self.Set(test.updatedValue, nil)
				},
			)
			value, err := item.Get()

			if err != nil {
				assert.Equal(t, value, test.initialValue)
			}
			assert.ErrorIs(t, err, test.err)

			assert.Equal(t, updateCalled, test.shouldUpdate)
			if test.shouldUpdate {
				value, err := item.Get()
				assert.Equal(t, value, test.updatedValue)
				assert.NoError(t, err)
			}
		})
	}
}
