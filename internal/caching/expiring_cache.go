package caching

import (
	"sync"
	"time"
)

type cachedEntry[T any] struct {
	data      T
	timestamp time.Time
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

// Set explicitly sets data in the cache
func (c *Cache[T]) Set(data T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entry = &cachedEntry[T]{
		data:      data,
		timestamp: time.Now(),
	}
}

// Invalidate forces the cache to be considered invalid
func (c *Cache[T]) Invalidate() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entry = nil
}

// Caller must hold a read lock
func (c *Cache[T]) isValid() bool {
	if c.entry == nil {
		return false
	}

	expirationTime := c.entry.timestamp.Add(c.ttl)
	return time.Now().Before(expirationTime)
}

// Caller must hold a write lock
func (c *Cache[T]) fetchLocked() (T, error) {
	newData, err := c.fetchFunc()
	if err != nil {
		var zero T
		return zero, err
	}

	c.entry = &cachedEntry[T]{data: newData, timestamp: time.Now()}
	return newData, nil
}

func (c *Cache[T]) fetch() (T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.fetchLocked()
}
