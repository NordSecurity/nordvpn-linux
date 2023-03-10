package kernel

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestGetParams(t *testing.T) {
	category.Set(t, category.Unit)

	some := []byte("net.ipv4.ip_forward")
	rc, err := parametersFrom(some)

	assert.Error(t, err)
	assert.NotEqual(t, len(rc), 1)

	some = []byte("net.ipv4.ip_forward = 1")
	rc, err = parametersFrom(some)

	assert.NoError(t, err)
	assert.Equal(t, len(rc), 1)

	some = []byte("net.ipv4.ip_forward = 1 \n net.ipv4.ip_forward = 2 ")
	rc, err = parametersFrom(some)

	assert.NoError(t, err)
	assert.Equal(t, len(rc), 2)
}
