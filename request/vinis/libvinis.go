//go:build vinis

package vinis

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	vinisBindings "vinis"
)

const (
	// PinURL is the url for certificate pins
	PinURL = "https://"
)

// PinningTransport implements certificate pinning for HTTP requests
type PinningTransport struct {
	inner     http.RoundTripper
	pdpClient *vinisBindings.PdpUrlConnectionClient
	pdpCache  *vinisBindings.PdpCache
}

func New(inner http.RoundTripper) http.RoundTripper {
	//TODO/FIXME: remove after debug
	return inner
	// if inner == nil {
	// 	inner = http.DefaultTransport
	// }

	// transport, ok := inner.(*http.Transport)
	// if !ok {
	// 	panic("wrapped transport must be *http.Transport")
	// }

	// pt := &PinningTransport{
	// 	inner:     inner,
	// 	pdpClient: &vinisBindings.PdpUrlConnectionClient{Origin: PinURL},
	// 	pdpCache:  vinisBindings.NewMemoryCache(),
	// }

	// if transport.TLSClientConfig == nil {
	// 	transport.TLSClientConfig = &tls.Config{}
	// }

	// transport.TLSClientConfig.VerifyPeerCertificate = pt.verifyCertificates

	// return pt
}

// RoundTrip implements the http.RoundTripper interface
func (t *PinningTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.inner.RoundTrip(req)
}

func (t *PinningTransport) verifyCertificates(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	// when a TLS connection is established, the server presents a certificate chain.
	// verifiedChains represents the validated certificate chains after the standard TLS verification process
	if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
		return fmt.Errorf("no verified certificate chain")
	}

	// leaf certificate
	cert := verifiedChains[0][0]

	// Pins must match exactly the certificates Subject Common Name or Subject Alternative Name
	domains := NewDomainSet(cert.Subject.CommonName, cert.DNSNames...)

	//fmt.Println("~~~LEAF CERT CN:", domain)
	// fmt.Println("~~~CERT PUB KEY b64:", calculateEncodeB64Hash(cert.RawSubjectPublicKeyInfo))
	// fmt.Println("~~~CERT PUB KEY hex:", calculateEncodeHexHash(cert.RawSubjectPublicKeyInfo))

	dmns1 := domains.getDomains()

	fmt.Println("~~~vinisBindings.FindPins domains:", dmns1)

	pins, err := vinisBindings.FindPins(t.pdpCache, t.pdpClient, vinisBindings.PinQuery{Domains: dmns1})
	if err != nil {
		return fmt.Errorf("find pins: %w", err)
	}

	// RawSubjectPublicKeyInfo - DER encoded SubjectPublicKeyInfo.
	certPubKeyHash := hash(cert.RawSubjectPublicKeyInfo)

	for _, pin := range pins.Domains {
		for _, fp := range pin.PubKeyFingerprints {
			if comparePins(certPubKeyHash, fp) {
				return nil
			}
		}
	}

	return fmt.Errorf("certificate pin not found")
}

func hash(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

func comparePins(crrPin, pinnedPin []byte) bool {
	return bytes.Equal(crrPin, pinnedPin)
}

type domainSet struct {
	names map[string]struct{}
}

func NewDomainSet(name string, names ...string) *domainSet { //TODO/FIXME: add unit tests
	ds := &domainSet{
		names: map[string]struct{}{},
	}
	ds.addName(name)
	for _, nm := range names {
		ds.addName(nm)
	}
	return ds
}

func (ds *domainSet) getDomains() []string {
	rc := []string{}
	for k := range ds.names {
		rc = append(rc, k)
	}
	return rc
}

func (ds *domainSet) addName(nm string) { //TODO/FIXME: add unit tests
	nmParts := strings.Split(nm, ".")
	if len(nmParts) == 0 {
		return
	}
	if nmParts[0] == "*" && len(nmParts) > 2 { // avoid `*.com` case, but `*.nordvpn.com` is ok
		nm1 := strings.Join(nmParts[1:], ".")
		ds.names[nm1] = struct{}{}
	}
	if nmParts[0] != "*" && len(nmParts) >= 2 {
		nm1 := strings.Join(nmParts[:], ".")
		ds.names[nm1] = struct{}{}
	}
}
