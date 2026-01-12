package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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

// Test that TP nameservers and resolver are build, without blocking until servers list is downloaded
func TestBuildTpServersAndResolver(t *testing.T) {
	category.Set(t, category.Unit)

	var wg sync.WaitGroup
	wg.Add(1)
	serversList := []string{"1.2.3.4", "4.4.5.6"}
	server := mock.NewHTTPTestServer(t,
		[]mock.Handler{
			mock.Handler{
				Pattern: core.ThreatProtectionLiteURL,
				Fn: func() ([]byte, *mock.HTTPError) {
					defer wg.Done()

					b, _ := json.Marshal(serversList)
					response := fmt.Sprintf("{\"servers\":%s}", string(b))
					return []byte(response), nil
				},
			},
		},
	)

	server.Start()
	defer server.Close()

	startPoint := time.Now()
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
	duration := time.Now().UnixMilli() - startPoint.UnixMilli()
	assert.Less(t, duration, 100*time.Millisecond.Milliseconds())

	assert.NotNil(t, tp)
	assert.NotNil(t, resolver)

	wg.Wait()

	// wait a few milliseconds to give time to set the data into the list
	time.Sleep(time.Millisecond * 10)
	assert.ElementsMatch(t, serversList, tp.Get(true))
}
