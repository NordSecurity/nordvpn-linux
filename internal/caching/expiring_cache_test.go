package caching

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func Test_CacheCreation(t *testing.T) {
	category.Set(t, category.Unit)

	fetchFunc := func() (string, error) {
		return "test", nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	assert.NotNil(t, cache, "cache should not be nil")
	assert.Equal(t, 1*time.Minute, cache.ttl, "expected ttl to be 1 minute")
	assert.Nil(t, cache.entry, "cache entry should be nil initially")
}

func Test_GetFromOrigin(t *testing.T) {
	category.Set(t, category.Unit)

	expectedData := "test data"
	fetchCount := 0

	fetchFunc := func() (string, error) {
		fetchCount++
		return expectedData, nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	data, err := cache.Get()

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, expectedData, data, "expected correct data to be returned")
	assert.Equal(t, 1, fetchCount, "expected fetch function to be called once")
}

func Test_GetFromCache(t *testing.T) {
	category.Set(t, category.Unit)

	expectedData := "test data"
	fetchCount := 0

	fetchFunc := func() (string, error) {
		fetchCount++
		return expectedData, nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	// first call will fetch from origin
	cache.Get()

	// second call should be from cache
	data, err := cache.Get()

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, expectedData, data, "expected correct data to be returned")
	assert.Equal(t, 1, fetchCount, "expected fetch function to be called only once")
}

func Test_CacheExpiration(t *testing.T) {
	category.Set(t, category.Unit)

	fetchCount := 0
	fetchFunc := func() (int, error) {
		fetchCount++
		return fetchCount, nil
	}

	// cache with very short TTL
	cache := NewCacheWithTTL(50*time.Millisecond, fetchFunc)

	// first call will fetch from origin
	data1, _ := cache.Get()
	assert.Equal(t, 1, data1, "expected first call to return 1")

	// second call should be from cache
	data2, _ := cache.Get()
	assert.Equal(t, 1, data2, "expected second call to return same data as first")
	assert.Equal(t, 1, fetchCount, "expected fetch function to be called only once at this point")

	// wait for cache to expire
	time.Sleep(100 * time.Millisecond)

	// third call after TTL should fetch from origin again
	data3, _ := cache.Get()
	assert.Equal(t, 2, data3, "expected third call to return new data")
	assert.Equal(t, 2, fetchCount, "expected fetch function to be called twice")
}

func Test_FetchError(t *testing.T) {
	category.Set(t, category.Unit)

	expectedErr := errors.New("fetch error")
	fetchFunc := func() (string, error) {
		return "", expectedErr
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	data, err := cache.Get()

	assert.Equal(t, expectedErr, err, "expected error to be returned")
	assert.Equal(t, "", data, "expected empty data when error occurs")
	assert.Nil(t, cache.entry, "cache should remain empty after fetch error")
}

func Test_ExplicitSet(t *testing.T) {
	category.Set(t, category.Unit)

	fetchCount := 0
	fetchFunc := func() (string, error) {
		fetchCount++
		return "origin data", nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	explicitData := "explicit data"
	cache.Set(explicitData)

	data, err := cache.Get()

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, explicitData, data, "expected explicitly set data to be returned")
	assert.Equal(t, 0, fetchCount, "expected fetch function to not be called")
}

func Test_Invalidate(t *testing.T) {
	category.Set(t, category.Unit)

	fetchCount := 0
	fetchFunc := func() (string, error) {
		fetchCount++
		return fmt.Sprintf("data %d", fetchCount), nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	data1, _ := cache.Get()
	assert.Equal(t, "data 1", data1, "expected first call to return 'data 1'")

	cache.Invalidate()

	// after invalidation, should fetch from origin again
	data2, _ := cache.Get()
	assert.Equal(t, "data 2", data2, "expected call after invalidation to return 'data 2'")
	assert.Equal(t, 2, fetchCount, "expected fetch function to be called twice")
}

func Test_ZeroTTL(t *testing.T) {
	category.Set(t, category.Unit)

	fetchCount := 0
	fetchFunc := func() (string, error) {
		fetchCount++
		return fmt.Sprintf("data %d", fetchCount), nil
	}

	cache := NewCacheWithTTL(0, fetchFunc)

	// first call will fetch from origin
	data1, _ := cache.Get()
	assert.Equal(t, "data 1", data1, "expected first call to return 'data 1'")

	// second call should also fetch from origin due to zero TTL
	data2, _ := cache.Get()
	assert.Equal(t, "data 2", data2, "expected second call to return 'data 2'")
	assert.Equal(t, 2, fetchCount, "expected fetch function to be called twice")
}

func Test_NegativeTTL(t *testing.T) {
	category.Set(t, category.Unit)

	fetchCount := 0
	fetchFunc := func() (string, error) {
		fetchCount++
		return fmt.Sprintf("data %d", fetchCount), nil
	}

	cache := NewCacheWithTTL(-1*time.Minute, fetchFunc)

	// first call will fetch from origin
	data1, _ := cache.Get()
	assert.Equal(t, "data 1", data1, "expected first call to return 'data 1'")

	// second call should also fetch from origin due to negative TTL
	data2, _ := cache.Get()
	assert.Equal(t, "data 2", data2, "expected second call to return 'data 2'")
	assert.Equal(t, 2, fetchCount, "expected fetch function to be called twice")
}

func Test_ComplexTypes(t *testing.T) {
	category.Set(t, category.Unit)

	type Complex struct {
		ID    int
		Name  string
		Items []string
	}

	expected := Complex{
		ID:    42,
		Name:  "Complex Object",
		Items: []string{"item1", "item2"},
	}

	fetchFunc := func() (Complex, error) {
		return expected, nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	data, err := cache.Get()

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, expected, data, "expected complex data to match")
}

func Test_DirectFetch(t *testing.T) {
	category.Set(t, category.Unit)

	fetchCount := 0
	fetchFunc := func() (int, error) {
		fetchCount++
		return fetchCount, nil
	}

	cache := NewCacheWithTTL(1*time.Minute, fetchFunc)

	// call to fill cache
	cache.Get()

	// call fetch directly to bypass cache
	data, err := cache.fetch()

	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 2, data, "expected data to be 2")
	assert.Equal(t, 2, fetchCount, "expected fetch function to be called twice")
}

func Test_ConcurrentAccess(t *testing.T) {
	category.Set(t, category.Unit)

	var fetchCount int
	fetchFunc := func() (int, error) {
		// simulate work
		time.Sleep(10 * time.Millisecond)
		fetchCount++
		return fetchCount, nil
	}

	cache := NewCacheWithTTL(100*time.Millisecond, fetchFunc)

	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for range numGoroutines {
		go func() {
			data, _ := cache.Get()
			assert.Equal(t, 1, data, "expected all goroutines to get the same data")
			done <- true
		}()
	}

	// wait for all goroutines to finish
	for range numGoroutines {
		<-done
	}

	// should only have fetched once despite concurrent access
	assert.Equal(t, 1, fetchCount, "expected fetch function to be called once despite concurrent access")
}
