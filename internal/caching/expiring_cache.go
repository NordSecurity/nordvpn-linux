package caching

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

var (
	ErrNoCacheData = errors.New("no cache data available")
	ErrStaleData   = errors.New("stale cache data")
)

type cachedEntry[T any] struct {
	data      T
	timestamp time.Time
	stale     bool
}

// Cache provides a generic caching mechanism with validity checking
type Cache[T any] struct {
	entry     *cachedEntry[T]
	ttl       time.Duration
	fetchFunc func() (T, error)
	mutex     sync.RWMutex
}

// NewCacheWithTTL creates a new cache with a custom validity period and fetch function
func NewCacheWithTTL[T any](validity time.Duration, fetchFunc func() (T, error)) *Cache[T] {
	return &Cache[T]{
		ttl:       validity,
		fetchFunc: fetchFunc,
	}
}

// Get returns the data either from cache or fetches from origin depending on validity period
func (c *Cache[T]) Get() (T, error) {
	c.mutex.RLock()
	if c.isValid() {
		data := c.entry.data
		c.mutex.RUnlock()
		return data, nil
	}
	c.mutex.RUnlock()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// check again under the write lock to avoid double fetching
	if c.isValid() {
		return c.entry.data, nil
	}

	return c.fetchLocked()
}

// GetEvenIfStale returns data regardless of staleness but with a bool indicating freshness
func (c *Cache[T]) GetEvenIfStale() (data T, fresh bool, err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var zero T
	if c.entry == nil {
		return zero, false, ErrNoCacheData
	}

	return c.entry.data, c.isValid(), nil
}

// Set explicitly sets data in the cache
func (c *Cache[T]) Set(data T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entry = &cachedEntry[T]{
		data:      data,
		timestamp: time.Now(),
		stale:     false,
	}
}

// Invalidate marks the cache as stale
func (c *Cache[T]) Invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.entry != nil {
		c.entry.stale = true
	}
}

// Clear completely removes any cached data
func (c *Cache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entry = nil
}

// Fetch forces a refresh of the data regardless of current validity
func (c *Cache[T]) Fetch() (T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.fetchLocked()
}

// Caller must hold a read lock
func (c *Cache[T]) isValid() bool {
	if c.entry == nil || c.entry.stale {
		return false
	}

	expirationTime := c.entry.timestamp.Add(c.ttl)
	return time.Now().Before(expirationTime)
}

// Caller must hold a write lock
func (c *Cache[T]) fetchLocked() (T, error) {
	var zero T

	if c.fetchFunc == nil {
		if c.entry == nil {
			return zero, ErrNoCacheData
		}

		v := reflect.ValueOf(c.entry.data)
		if v.Kind() == reflect.Ptr && v.IsNil() {
			return zero, ErrNoCacheData
		}

		return c.entry.data, ErrStaleData
	}

	newData, err := c.fetchFunc()
	if err != nil {
		// if fetch fails but we have existing data, return stale data with error
		if c.entry != nil {
			return c.entry.data, ErrStaleData
		}
		return zero, err
	}

	c.entry = &cachedEntry[T]{
		data:      newData,
		timestamp: time.Now(),
		stale:     false,
	}
	return newData, nil
}
