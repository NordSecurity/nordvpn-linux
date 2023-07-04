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
	Restart() MetaTransport
	Rotate() (MetaTransport, error)
}

type HTTPClient struct {
	client           *http.Client
	publisher        events.Publisher[events.DataRequestAPI]
	h3reviveTime     time.Duration
	lastH3okTime     time.Time
	lastOKTransport  MetaTransport
	currentTransport MetaTransport
	rotator          CompleteRotator
	mu               sync.Mutex
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
	IsH3      bool
}

func (t MetaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.Transport.RoundTrip(req)
}

func NewHTTPClient(
	client *http.Client,
	rotator CompleteRotator,
	publisher events.Publisher[events.DataRequestAPI],
) *HTTPClient {
	currentTransport := MetaTransport{}
	if rotator != nil {
		currentTransport = rotator.Restart()
	}
	return &HTTPClient{
		client:           client,
		rotator:          rotator,
		lastH3okTime:     time.Now(),
		h3reviveTime:     1 * time.Minute, // Provisional time for testing
		publisher:        publisher,
		currentTransport: currentTransport,
	}
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
	c.mu.Lock()
	defer c.mu.Unlock()

	var resp *http.Response
	var cliErr error

	// If it is not the 1st time after App has started, it checks if the HTTP/3 transport should be revived
	if c.lastOKTransport.Transport != nil {
		transport := c.lastOKTransport
		if !transport.IsH3 && (time.Now()).After(c.lastH3okTime.Add(c.h3reviveTime)) && c.rotator != nil {
			c.currentTransport = c.rotator.Restart()
		}
	}
	for resp, cliErr = c.do(req); cliErr != nil && c.rotator != nil; resp, cliErr = c.do(req) {
		log.Println(cliErr)
		if c.rotator == nil {
			break
		}
		nextTransport, err := c.rotator.Rotate()
		if errors.Is(err, ErrNothingMoreToRotate) {
			break
		}
		c.currentTransport = nextTransport
		if err != nil {
			return nil, fmt.Errorf("rotating: %w", err)
		}
	}
	if cliErr != nil {
		return nil, fmt.Errorf("doing http request: %w", cliErr)
	}
	c.lastOKTransport = c.currentTransport
	if c.currentTransport.IsH3 {
		c.lastH3okTime = time.Now()
	}
	return resp, nil
}

func (c *HTTPClient) do(req *http.Request) (*http.Response, error) {
	c.setRequestTransferProtocol(req)
	tmpCli := *c.client
	if c.currentTransport.Transport != nil && c.currentTransport.Transport != c.client.Transport {
		tmpCli.Transport = c.currentTransport
	}
	startTime := time.Now()
	resp, err := tmpCli.Do(req)
	if c.publisher != nil {
		c.publisher.Publish(events.DataRequestAPI{
			Request:  req,
			Response: resp,
			Error:    err,
			Duration: time.Since(startTime),
		})
	}
	return resp, err
}

type transferProtocol struct {
	proto string
	major int
	minor int
}

func (c *HTTPClient) setRequestTransferProtocol(req *http.Request) {
	transferProto := transportToProtocol(c.currentTransport)
	req.Proto = transferProto.proto
	req.ProtoMajor = transferProto.major
	req.ProtoMinor = transferProto.minor
}

func transportToProtocol(transport MetaTransport) transferProtocol {
	var transferProto transferProtocol

	if transport.IsH3 {
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
