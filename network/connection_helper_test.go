package network

import (
	"log"
	"net"
	"net/netip"
	"sync"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/miekg/dns"

	"github.com/stretchr/testify/assert"
)

func startTestDNSServer(t *testing.T, domainName string, ip netip.Addr) (addr string, shutdown func()) {
	t.Helper()

	mux := dns.NewServeMux()
	mux.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		msg := new(dns.Msg)
		msg.SetReply(r)

		for _, q := range r.Question {
			if q.Name == domainName+"." && q.Qtype == dns.TypeA {
				msg.Answer = append(msg.Answer, &dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    60,
					},
					A: ip.AsSlice(),
				})
			}
		}

		_ = w.WriteMsg(msg)
	})

	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}

	server := &dns.Server{
		PacketConn: pc,
		Handler:    mux,
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := server.ActivateAndServe(); err != nil {
			t.Fatalf("activate server: %v", err)
		}
	})

	log.Println("running DNS server on", pc.LocalAddr().String(), "for", domainName)

	return pc.LocalAddr().String(), func() {
		_ = server.Shutdown()
		_ = pc.Close()
	}
}

func TestResolveHostUsingLocalDnsServer(t *testing.T) {
	category.Set(t, category.Unit)

	domainName := "nordvpn.com"
	ipAddress := netip.MustParseAddr("1.2.3.4")

	dnsAddr, shutdown := startTestDNSServer(t, domainName, ipAddress)
	defer shutdown()

	result, err := LookupAddressNoFwmark(domainName, dnsAddr, "udp")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	if len(result) > 0 {
		assert.Equal(t, ipAddress, result[0])
	}
}

func TestResolveHost(t *testing.T) {
	category.Set(t, category.Unit)

	result, err := LookupAddressNoFwmark("google.com", "1.1.1.1", "udp")
	assert.NoError(t, err)
	assert.NotEmpty(t, result[0])
	assert.NotEmpty(t, result[1])
}
