package mapper

import (
	"io"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mapper struct {
	value *mesh.MachineMap
	err   error
}

func (m *mapper) Map(_ string, _ uuid.UUID) (*mesh.MachineMap, error) {
	return m.value, m.err
}

// TestCachingMapper_Map simply checks whether inner mapper is used in the implementation without
// checking the CachedValue internals.
func TestCachingMapper_Map(t *testing.T) {
	for _, tt := range []struct {
		name  string
		err   error
		inner mesh.Mapper
		mmap  *mesh.MachineMap
	}{
		{
			name:  "no error",
			err:   nil,
			inner: &mapper{value: &mesh.MachineMap{}},
			mmap:  &mesh.MachineMap{},
		},
		{
			name:  "error",
			err:   io.EOF,
			inner: &mapper{err: io.EOF},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewCachingMapper(tt.inner, time.Second)
			mmap, err := mapper.Map("any", uuid.New(), false)
			assert.ErrorIs(t, tt.err, err)
			assert.EqualValues(t, tt.mmap, mmap)
			assert.Equal(t, tt.err == nil, mapper.mmap.cachedDate != time.Time{})
			assert.Equal(t, tt.err == nil, mapper.mmap.value != nil)
		})
	}
}

func TestCachedValue_Get(t *testing.T) {
	category.Set(t, category.Unit)
	fnRetOne := func(_ int) (int, error) {
		return 1, nil
	}
	for _, tt := range []struct {
		name        string
		key         int
		forceUpdate bool
		err         error
		expValue    int
		updated     bool
		cv          *CachedValue[int, int]
	}{
		{
			name:     "return func res",
			expValue: 1,
			key:      1,
			updated:  true,
			cv: NewCachedValue(time.Second, func(key int) (int, error) {
				return key, nil
			}),
		},
		{
			name:     "return cached value on nil GetFn",
			expValue: 0,
			key:      1,
			updated:  false,
			cv:       &CachedValue[int, int]{},
		},
		{
			name:     "get while cache still valid",
			expValue: 2,
			key:      2,
			updated:  false,
			cv: func() *CachedValue[int, int] {
				cv := NewCachedValue(time.Second, fnRetOne)
				cv.cachedDate = time.Now()
				cv.value = 2
				cv.key = 2
				return cv
			}(),
		},
		{
			name:        "forceUpdate causes get fn call",
			expValue:    1,
			key:         2,
			forceUpdate: true,
			updated:     true,
			cv: func() *CachedValue[int, int] {
				cv := NewCachedValue(time.Second, fnRetOne)
				cv.cachedDate = time.Now()
				cv.value = 2
				cv.key = 2
				return cv
			}(),
		},
		{
			name:     "outdated value causes get fn call",
			expValue: 1,
			key:      2,
			updated:  true,
			cv: func() *CachedValue[int, int] {
				cv := NewCachedValue(time.Second, fnRetOne)
				cv.value = 2
				cv.key = 2
				return cv
			}(),
		},
		{
			name:     "new key value causes get fn call",
			expValue: 1,
			key:      3,
			updated:  true,
			cv: func() *CachedValue[int, int] {
				cv := NewCachedValue(time.Second, fnRetOne)
				cv.cachedDate = time.Now()
				cv.value = 2
				cv.key = 2
				return cv
			}(),
		},
		{
			name:        "failing function does not update",
			expValue:    0,
			key:         2,
			err:         io.EOF,
			forceUpdate: true,
			updated:     false,
			cv: func() *CachedValue[int, int] {
				cv := NewCachedValue(time.Second, func(_ int) (int, error) {
					return 0, io.EOF
				})
				cv.cachedDate = time.Now()
				cv.value = 2
				cv.key = 2
				return cv
			}(),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cachedBefore := tt.cv.cachedDate
			val, err := tt.cv.Get(tt.key, tt.forceUpdate)
			assert.Equal(t, tt.expValue, val)
			assert.ErrorIs(t, tt.err, err)
			assert.Equal(t, tt.updated, cachedBefore.Before(tt.cv.cachedDate))
		})
	}
}
