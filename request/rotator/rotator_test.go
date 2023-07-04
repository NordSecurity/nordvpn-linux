package rotator

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRotator_Rotate(t *testing.T) {
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
			expected: nil,
			hasError: true,
		},
		{
			name: "nil transports",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: nil,
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "no url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}, {Transport: transportB{}}},
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "nil rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}, {Transport: transportB{}}},
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "transport with url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}},
			},
			expected: nil,
			hasError: true,
		},
		{
			name: "transport with failing url rotator",
			fields: fields{
				client:     &http.Client{Transport: transportA{}},
				transports: []request.MetaTransport{{Transport: transportA{}}},
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
			},
			expected: nil,
			hasError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := NewRotator(test.fields.transports)
			if err != nil {
				assert.Equal(t, test.hasError, err != nil, "expected: %t, actual: %s", test.hasError, err)
				return
			}
			for i := 1; i < len(r.elements); i++ {
				transport, err := r.Rotate()
				assert.Equal(t, r.elements[i], transport)
				assert.NoError(t, err)
			}
			_, err = r.Rotate()
			assert.Equal(t, test.hasError, err != nil, "expected: %t, actual: %s", test.hasError, err)
		})
	}
}

func TestRotator_Restart(t *testing.T) {
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
			r, err := NewRotator(test.fields.transports)
			if err != nil {
				assert.True(t, test.hasError)
				return
			}
			require.NoError(t, err)
			if !test.hasError {
				r.Restart()
				assert.Equal(t, test.fields.transports[0], r.elements[0])
			}
		})
	}
}
