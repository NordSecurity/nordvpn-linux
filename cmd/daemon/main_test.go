package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildClientAPIAndSessionStores(t *testing.T) {
	category.Set(t, category.Unit)

	clientAPI, sessionBuilder := buildClientAPIAndSessionStores(
		"test-agent",
		&http.Client{},
		response.NoopValidator{},
		mock.NewMockConfigManager(),
	)

	assert.NotNil(t, clientAPI)
	assert.NotNil(t, sessionBuilder)

	stores := sessionBuilder.GetStores()
	assert.Len(t, stores, 4)

	for i, store := range stores {
		assert.NotNil(t, store, "Store at index %d should not be nil", i)
	}
}

// Test that TP nameservers and resolver are build
func TestBuildTpServersAndResolver(t *testing.T) {
	category.Set(t, category.Unit)

	var wg sync.WaitGroup
	wg.Add(1)

	var fetched atomic.Bool

	serversList := []string{"1.2.3.4", "4.4.5.6"}

	server := mock.NewHTTPTestServer(t, []mock.Handler{
		{
			Pattern: core.ThreatProtectionLiteURL,
			Fn: func() ([]byte, *mock.HTTPError) {
				// simulate that fetching takes more time, also gives time to check fetched value
				time.Sleep(time.Millisecond * 20)
				assert.False(t, fetched.Load(), "must execute only once")
				fetched.Store(true)

				defer wg.Done()
				b, _ := json.Marshal(serversList)
				response := fmt.Sprintf("{\"servers\":%s}", string(b))
				return []byte(response), nil
			},
		},
	})

	server.Start()
	defer server.Close()

	tp, resolver := buildTpServersAndResolver(
		"test-agent",
		server.URL(),
		http.DefaultClient,
		response.NoopValidator{},
		&firewall.Firewall{},
		func(attempt int) time.Duration {
			assert.Fail(t, "this must be called only for error while fetching")
			return time.Minute
		},
	)

	assert.False(t, fetched.Load(), "fetcher must not be executed when building the objects")
	assert.NotNil(t, tp)
	assert.NotNil(t, resolver)

	wg.Wait()

	assert.True(t, fetched.Load(), "servers were fetched")

	// wait a few milliseconds to give time to set the data into the list
	time.Sleep(time.Millisecond * 5)
	assert.ElementsMatch(t, serversList, tp.Get(true))
}
