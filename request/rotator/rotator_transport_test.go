package rotator

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/events/subs"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

type transportA struct{}

func (transportA) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("A")
}
func (transportA) NotifyConnect(events.DataConnect) error { return nil }

type transportB struct{}

func (transportB) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("B")
}
func (transportB) NotifyConnect(events.DataConnect) error { return nil }

func TestTransportRotator_Rotate(t *testing.T) {
	category.Set(t, category.Unit)

	bURL, _ := url.Parse("www.example.com")

	type fields struct {
		client     *http.Client
		transports []request.MetaTransport
		baseURL    *url.URL
	}
	tests := []struct {
		name     string
		fields   fields
		expected *url.URL
		hasError bool
	}{
		{
			name: "no transports",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{},
				baseURL:    bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "nil transports",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: nil,
				baseURL:    bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "no url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}, {Transport: transportB{}}},
				baseURL:    bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "nil rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}, {Transport: transportB{}}},
				baseURL:    bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "transport with url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}},
				baseURL:    bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "transport with failing url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}},
				baseURL:    bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "transports with url rotator",
			fields: fields{
				client: &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{
					{Transport: transportA{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
				},
				baseURL: bURL,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "transports with failing rotator",
			fields: fields{
				client: &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{
					{Transport: transportA{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
				},
				baseURL: bURL,
			},
			expected: nil,
			hasError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := request.NewHTTPClient(test.fields.client, test.fields.baseURL.String(), &subs.Subject[string]{}, nil)
			r := NewTransportRotator(c, test.fields.transports)
			c.CompleteRotator = r
			for i := 1; i < len(r.transports); i++ {
				err := r.Rotate()
				assert.Equal(t, c.SelectedTransport, r.transports[i])
				assert.NoError(t, err)
			}
			err := r.Rotate()
			assert.Equal(t, test.hasError, err != nil, "expected: %t, actual: %s", test.hasError, err)
		})
	}
}

func TestTransportRotator_Restart(t *testing.T) {
	category.Set(t, category.Unit)

	type fields struct {
		client     *http.Client
		transports []request.MetaTransport
	}
	tests := []struct {
		name     string
		fields   fields
		expected *url.URL
		hasError bool
	}{
		{
			name: "no transports",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{},
			},
			hasError: true,
		},
		{
			name: "nil transports",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: nil,
			},
			hasError: true,
		},
		{
			name: "no url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}, {Transport: transportB{}}},
			},
			hasError: false,
		},
		{
			name: "nil rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}, {Transport: transportB{}}},
			},
			hasError: false,
		},
		{
			name: "transport with url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}},
			},
			hasError: false,
		},
		{
			name: "transport with failing url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}},
			},
			hasError: true,
		},
		{
			name: "transports with url rotator",
			fields: fields{
				client: &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{
					{Transport: transportA{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
				},
			},
			hasError: false,
		},
		{
			name: "transports with failing rotator",
			fields: fields{
				client: &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{
					{Transport: transportA{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
					{Transport: transportA{}},
					{Transport: transportB{}},
				},
			},
			hasError: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := request.NewHTTPClient(test.fields.client, "", &subs.Subject[string]{}, nil)
			r := NewTransportRotator(c, test.fields.transports)
			c.CompleteRotator = r
			if !test.hasError {
				r.Restart()
				assert.Equal(t, test.fields.transports[0], r.transports[0])
			}
		})
	}
}
