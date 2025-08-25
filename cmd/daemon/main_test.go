package main

import (
	"net/http"
	"testing"

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
