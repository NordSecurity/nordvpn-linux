//go:build quench

package quench

type Spec struct {
	TlsDomain string `json:"tls_domain"`
}

type Protocol struct {
	Addr string `json:"addr"`
	Spec Spec   `json:"spec"`
}

type Config struct {
	Protocol Protocol `json:"protocol"`
}
