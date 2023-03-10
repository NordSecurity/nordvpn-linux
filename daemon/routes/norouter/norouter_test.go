package norouter

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/routes"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestImplementsAgent(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Implements(t, (*routes.Agent)(nil), &Facade{})
}
