package service

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestReturnsNilByDefault(t *testing.T) {
	category.Set(t, category.Unit)

	m := NoopFileshare{}
	assert.Nil(t, m.Enable(0, 0))
	assert.Nil(t, m.Disable(0, 0))
	assert.Nil(t, m.Stop(0, 0))
}
