package notables

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/firewall"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestAgentInterface(t *testing.T) {
	category.Set(t, category.Unit)
	assert.Implements(t, (*firewall.Agent)(nil), &Facade{})
}
