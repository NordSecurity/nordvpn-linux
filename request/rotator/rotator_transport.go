// Package rotator is responsible for api request transport rotation.
package rotator

import (
	"log"
	"net/url"

	"github.com/NordSecurity/nordvpn-linux/request"
)

// TransportRotator handles the HTTPClient transports rotation
// It rotates the active transport when needed.
// The selected transport is changed among the available transports in the transports slice.
type TransportRotator struct {
	client     *request.HTTPClient
	transports []request.MetaTransport
	baseURL    *url.URL
	index      int
}

func NewTransportRotator(client *request.HTTPClient, transports []request.MetaTransport) *TransportRotator {
	if len(transports) > 0 {
		client.SetTransport(transports[0])
	}
	url, _ := url.Parse(client.BaseURL)
	return &TransportRotator{
		client:     client,
		transports: transports,
		baseURL:    url,
	}
}

// Rotate changes the active HTTPClient transport or the URL domain when needed.
func (r *TransportRotator) Rotate() error {
	lastTransport := r.index+1 >= len(r.transports)
	if lastTransport {
		return request.ErrNothingMoreToRotate
	} else {
		r.index++
		r.client.SetTransport(r.transports[r.index])
		log.Printf("rotated api transport to: %s\n", r.client.SelectedTransport.Name)
	}
	return nil
}

// Restart sets the active transport to the first one available on the transports slice
func (r *TransportRotator) Restart() {
	if len(r.transports) > 0 {
		r.index = 0
		r.client.SetTransport(r.transports[r.index])
	}
}
