// Package request provides convenient way for sending HTTP requests.
package request

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/events"
)

var (
	ErrBearerTokenNotProvided         = errors.New("bearer token not provided")
	ErrNothingMoreToRotate            = errors.New("nothing more to rotate")
	ErrUsernameAndPasswordNotProvided = errors.New("username and password not provided")
)

type BasicAuth struct {
	Username string
	Password string
}

type CompleteRotator interface {
	Restart()
	Rotate() error
}

type HTTPClient struct {
	client            *http.Client
	BaseURL           string
	LastResponseTime  time.Duration
	h3reviveTime      time.Duration
	lastH3okTime      time.Time
	publisher         events.Publisher[string]
	lastOKTransport   MetaTransport
	SelectedTransport MetaTransport
	CompleteRotator
	sync.Mutex
}

type RoundTripperEx interface {
	http.RoundTripper
	NotifyConnect(events.DataConnect) error
}

// MetaTransport contains a transport and information whether the
// transport is a HTTP3 implementation or not. This was done in order
// to avoid type checks
type MetaTransport struct {
	Transport RoundTripperEx
	Name      string
	isH3      bool
	create    func() RoundTripperEx
}

func NewH1Transport(fn func() RoundTripperEx) MetaTransport {
	return MetaTransport{
		Transport: fn(),
		create:    fn,
	}
}

func NewH3Transport(fn func() RoundTripperEx) MetaTransport {
	return MetaTransport{
		Transport: fn(),
		isH3:      true,
		create:    fn,
	}
}

func (t MetaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.Transport.RoundTrip(req)
}

func NewHTTPClient(
	client *http.Client,
	baseURL string,
	publisher events.Publisher[string],
	rotator CompleteRotator,
) *HTTPClient {
	return &HTTPClient{
		client:          client,
		BaseURL:         baseURL,
		publisher:       publisher,
		CompleteRotator: rotator,
		lastH3okTime:    time.Now(),
		h3reviveTime:    1 * time.Minute, // Provisional time for testing
	}
}

func (c *HTTPClient) SetTransport(transport MetaTransport) {
	c.SelectedTransport = transport
}

// NewRequest builds an unauthenticated request.
func NewRequest(
	method, agent, baseURL, pathURL, contentType, contentLength, encoding string,
	body io.Reader,
) (*http.Request, error) {
	tmpBody := body
	if method == http.MethodGet {
		tmpBody = nil
	}

	req, err := http.NewRequest(method, baseURL+pathURL, tmpBody)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	req.Header.Set("User-Agent", agent)
	req.Header.Set("Accept-Encoding", encoding)
	req.Header.Set("Content-Type", contentType)
	req.Header.Add("Content-Length", contentLength)

	return req, nil
}

// NewRequestWithBasicAuth builds a Basic Auth authenticated request.
func NewRequestWithBasicAuth(
	method, agent, baseURL, pathURL, contentType, contentLength, encoding string,
	body io.Reader,
	auth *BasicAuth,
) (*http.Request, error) {
	req, err := NewRequest(method, agent, baseURL, pathURL, contentType, contentLength, encoding, body)
	if err != nil {
		return nil, err
	}

	if auth == nil || auth.Username == "" || auth.Password == "" {
		return nil, ErrUsernameAndPasswordNotProvided
	}
	req.SetBasicAuth(auth.Username, auth.Password)

	return req, nil
}

// NewRequestWithBearerToken builds a Bearer Token authenticated request.
func NewRequestWithBearerToken(
	method, agent, baseURL, pathURL, contentType, contentLength, encoding string,
	body io.Reader,
	token string,
) (*http.Request, error) {
	req, err := NewRequest(method, agent, baseURL, pathURL, contentType, contentLength, encoding, body)
	if err != nil {
		return nil, err
	}

	if token == "" {
		return nil, ErrBearerTokenNotProvided
	}
	req.Header.Set("Authorization", "Bearer token:"+token)

	return req, nil
}

// DoRequest performs a HTTP request. The behavior is as follows:
// 1. The client starts with the h3 transport.
// 2. If the request fails, it rotates the transport.
// 3. If it keeps failing, it rotates to the last available transport, the default one.
// Throughout this process, if the h3 timer runs out, the transport is reset to h3 and the process starts again.
func (c *HTTPClient) DoRequest(req *http.Request) (*http.Response, error) {
	c.Lock()
	defer c.Unlock()

	var resp *http.Response
	var cliErr error

	// If it is not the 1st time after App has started, it checks if the HTTP/3 transport should be revived
	if c.lastOKTransport.Transport != nil {
		transport := c.lastOKTransport
		if !transport.isH3 && (time.Now()).After(c.lastH3okTime.Add(c.h3reviveTime)) {
			c.CompleteRotator.Restart()
		}
	}
	for resp, cliErr = c.do(req); cliErr != nil && c.CompleteRotator != nil; resp, cliErr = c.do(req) {
		log.Println(cliErr)
		err := c.CompleteRotator.Rotate()
		if errors.Is(err, ErrNothingMoreToRotate) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("rotating: %w", err)
		}
	}
	if cliErr != nil {
		return nil, fmt.Errorf("doing http request: %w", cliErr)
	}
	c.lastOKTransport = c.SelectedTransport
	if c.SelectedTransport.isH3 {
		c.lastH3okTime = time.Now()
	}
	return resp, nil
}

func (c *HTTPClient) do(req *http.Request) (*http.Response, error) {
	if c.publisher != nil {
		c.publisher.Publish(fmt.Sprintf("URL: %s\n", c.BaseURL))
	}
	c.setRequestTransferProtocol(req)
	tmpCli := *c.client
	if c.SelectedTransport.Transport != nil && c.SelectedTransport.Transport != c.client.Transport {
		tmpCli.Transport = c.SelectedTransport
	}
	startTime := time.Now()
	resp, err := tmpCli.Do(req)
	c.LastResponseTime = time.Since(startTime)
	return resp, err
}

type transferProtocol struct {
	proto string
	major int
	minor int
}

func (c *HTTPClient) setRequestTransferProtocol(req *http.Request) {
	transferProto := transportToProtocol(c.SelectedTransport)
	req.Proto = transferProto.proto
	req.ProtoMajor = transferProto.major
	req.ProtoMinor = transferProto.minor
}

func transportToProtocol(transport MetaTransport) transferProtocol {
	var transferProto transferProtocol

	if transport.isH3 {
		transferProto.proto = "HTTP/3"
		transferProto.major = 3
		transferProto.minor = 0
	} else {
		transferProto.proto = "HTTP/1.1"
		transferProto.major = 1
		transferProto.minor = 1
	}

	return transferProto
}
