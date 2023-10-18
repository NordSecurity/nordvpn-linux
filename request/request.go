// Package request provides convenient way for sending HTTP requests.
package request

import (
	"errors"
	"fmt"
	"io"
	"net/http"
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
