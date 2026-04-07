package network

import (
	"fmt"
	"log"
	"net/netip"
	"os/exec"
	"strings"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/daemon/dns"
	daemonevents "github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	mockevents "github.com/NordSecurity/nordvpn-linux/test/mock/events"
	"github.com/stretchr/testify/assert"
)

func TestResolverUpdatesVpnStatus(t *testing.T) {
	category.Set(t, category.Unit)

	const fwmark uint32 = 0x1234

	ev := &daemonevents.ServiceEvents{
		Connect:    mockevents.NewMockPublisherSubscriber[events.DataConnect](),
		Disconnect: mockevents.NewMockPublisherSubscriber[events.DataDisconnect](),
	}
	resolver := NewResolver(&dns.NameServers{}, fwmark, ev)

	r, ok := resolver.(*Resolver)
	assert.True(t, ok)

	assert.Equal(t, fwmark, r.fwmark)
	assert.False(t, r.isVpnConnected.Load())

	ev.Connect.Publish(events.DataConnect{EventStatus: events.StatusAttempt})
	assert.False(t, r.isVpnConnected.Load())

	ev.Connect.Publish(events.DataConnect{EventStatus: events.StatusCanceled})
	assert.False(t, r.isVpnConnected.Load())

	ev.Connect.Publish(events.DataConnect{EventStatus: events.StatusFailure})
	assert.False(t, r.isVpnConnected.Load())

	ev.Connect.Publish(events.DataConnect{EventStatus: events.StatusSuccess})
	assert.True(t, ok)

	ev.Disconnect.Publish(events.DataDisconnect{})
	assert.False(t, r.isVpnConnected.Load())
}

const nftBlockAll = `
table inet nordvpn_test
delete table inet nordvpn_test
table inet nordvpn_test {
  chain input {
    type filter hook input priority 0; policy drop;
  }

  chain output {
    type filter hook output priority 0; policy drop;
  }
}`

const nftAllowFwMark = `
table inet nordvpn_test
delete table inet nordvpn_test
table inet nordvpn_test {
  chain input {
    type filter hook input priority 0; policy drop;
	ct mark 0xe1f1 accept
  }

  chain output {
    type filter hook output priority 0; policy drop;
	meta mark 0xe1f1 ct mark set meta mark accept
	ct state established,related ct mark 0xe1f1 accept
  }
}`

func TestResolver(t *testing.T) {
	category.Set(t, category.Root)

	domainName := "nordvpn.com"
	ipAddress := netip.MustParseAddr("1.2.3.4")
	dnsAddr, shutdown := startTestDNSServer(t, domainName, ipAddress)
	assert.NotNil(t, dnsAddr)
	defer shutdown()

	const fwmark uint32 = 0xe1f1

	ev := &daemonevents.ServiceEvents{
		Connect:    mockevents.NewMockPublisherSubscriber[events.DataConnect](),
		Disconnect: mockevents.NewMockPublisherSubscriber[events.DataDisconnect](),
	}
	r := NewResolver(&mock.DNSGetter{Names: []string{}}, fwmark, ev)
	resolver, ok := r.(*Resolver)
	assert.True(t, ok)

	dnsRequestAreBlocked := func() {
		_, err := resolver.resolveWithNameservers(domainName, []string{dnsAddr}, "udp")
		assert.Error(t, err, "check that nft blocks all")
	}

	dnsRequestWork := func() {
		ips, err := resolver.resolveWithNameservers(domainName, []string{dnsAddr}, "udp")
		assert.NoError(t, err, "response received when fwmark is used")
		assert.Equal(t, ipAddress, ips[0])
	}

	t.Cleanup(func() {
		configureNft("delete table inet nordvpn_test")
	})

	assert.NoError(t, configureNft(nftBlockAll), "failed to block entire traffic")

	dnsRequestAreBlocked()

	assert.NoError(t, configureNft(nftAllowFwMark), "failed to allow only fwmark connections")
	dnsRequestWork()

	ev.Connect.Publish(events.DataConnect{EventStatus: events.StatusSuccess})
	dnsRequestAreBlocked()

	ev.Disconnect.Publish(events.DataDisconnect{})
	dnsRequestWork()
}

func configureNft(cmd string) error {
	c := exec.Command("nft", "-f", "-")
	c.Stdin = strings.NewReader(cmd)

	out, err := c.CombinedOutput()
	if err != nil {
		log.Println("failed to execute nft", cmd, err)

		return fmt.Errorf("nft failed: %v\noutput:\n%s", err, string(out))
	}
	return nil
}
