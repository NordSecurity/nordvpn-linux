package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/stretchr/testify/assert"
)

func TestBuildClientAPIAndSessionStores(t *testing.T) {
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

func TestBuildTpServersAndResolver(t *testing.T) {
	serversList := []string{"1.2.3.4", "4.4.5.6"}
	sort.Strings(serversList)
	server := mock.NewHTTPTestServer(t,
		[]mock.Handler{
			mock.Handler{
				Pattern: core.ThreatProtectionLiteURL,
				Fn: func() ([]byte, *mock.HTTPError) {
					// time.Sleep(10 * time.Second)
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
	)
	duration := time.Now().UnixMilli() - startPoint.UnixMilli()
	assert.Less(t, duration, time.Second.Milliseconds())

	assert.NotNil(t, tp)
	assert.NotNil(t, resolver)

	// get the servers list and sort it because internally is shuffled
	result := tp.Get(true)
	sort.Strings(result)
	assert.Equal(t, serversList, result)
}
