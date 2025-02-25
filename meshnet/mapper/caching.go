package mapper

import (
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/google/uuid"
)

// CachingMapper is an implementation of mesh.CachingMapper that wraps inner Mapper and caching
// logic from CachedValue.
type CachingMapper struct {
	mmap *CachedValue[retrievalKey, *mesh.MachineMap]
}

// NewCachingMapper returns a new instance of CachingMapper filled newly created CachedValue and
// inner.Map function as GetFn for CachedValue.
func NewCachingMapper(inner mesh.Mapper, cacheTTL time.Duration) *CachingMapper {
	mapFn := func(key retrievalKey) (*mesh.MachineMap, error) {
		return inner.Map(key.token, key.id)
	}
	return &CachingMapper{
		mmap: NewCachedValue(cacheTTL, mapFn),
	}
}

// Map uses CachedValue.Get function, where GetFn is inner.Map function.
func (r *CachingMapper) Map(
	token string,
	self uuid.UUID,
	forceUpdate bool,
) (*mesh.MachineMap, error) {
	return r.mmap.Get(retrievalKey{token: token, id: self}, forceUpdate)
}

// retrievalKey is a structure acting as a key to CachedValue which composes of fields that are
// arguments to mesh.Mapper.Map function.
type retrievalKey struct {
	token string
	id    uuid.UUID
}

// Thread safe data caching
type CachedValue[K comparable, V any] struct {
	key        K
	value      V
	cachedDate time.Time
	validity   time.Duration
	getFn      GetFn[K, V]
	mu         sync.Mutex
}

// GetFn is a function that returns data and error in case it failed.
type GetFn[K, V any] func(K) (V, error)

// NewCachedValue creates a new instance of CachedValue filled with the given parameters.
func NewCachedValue[K comparable, V any](
	validity time.Duration,
	getFn GetFn[K, V],
) *CachedValue[K, V] {
	return &CachedValue[K, V]{
		validity: validity,
		getFn:    getFn,
	}
}

// Get returns either latest cached value or updates it before returning it.
// Even though it can work with different keys, it only saves the latest one and update is forced
// if key changes.
//
// Considering this getFn may be a long running function (such as an HTTP call), cached value
// returning is instant and key is not updated frequently, whole function is put under a single
// mutex lock.
// Thread safe.
func (c *CachedValue[K, V]) Get(key K, forceUpdate bool) (V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// to prevent race condition store function value in lock
	if c.getFn != nil &&
		(forceUpdate || key != c.key || c.cachedDate.Add(c.validity).Before(time.Now())) {
		val, err := c.getFn(key)
		if err != nil {
			return val, err
		}
		c.key = key
		c.value = val
		c.cachedDate = time.Now()
	}
	return c.value, nil
}
