package models

import (
	"sync"
	"time"
)

// Thread safe data caching
type CachedValue[T any] struct {
	value       T
	latestError error
	cachedDate  time.Time
	validity    time.Duration
	updaterFn   func(self *CachedValue[T])
	mu          *sync.Mutex
}

func NewCachedValue[T any](
	value T,
	latestError error,
	cachedDate time.Time,
	validity time.Duration,
	updaterFn func(*CachedValue[T]),
) *CachedValue[T] {
	return &CachedValue[T]{
		value:       value,
		latestError: latestError,
		cachedDate:  cachedDate,
		validity:    validity,
		updaterFn:   updaterFn,
		mu:          &sync.Mutex{},
	}
}

func (c *CachedValue[T]) Get() (T, error) {
	shouldUpdate := false
	c.mu.Lock()
	defer func() {
		// to prevent race condition store function value in lock
		updaterFn := c.updaterFn
		c.mu.Unlock()
		if shouldUpdate && updaterFn != nil {
			updaterFn(c)
		}
	}()
	shouldUpdate = c.cachedDate.Add(c.validity).Before(time.Now())
	return c.value, c.latestError
}

func (c *CachedValue[T]) Set(value T, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cachedDate = time.Now()
	c.value = value
	c.latestError = err
}

func (c *CachedValue[T]) ChangeUpdaterFn(updaterFn func(*CachedValue[T])) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updaterFn = updaterFn
}
