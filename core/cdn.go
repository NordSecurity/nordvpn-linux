package core

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/request"
)

// CDN provides methods to interact with Nord's Content Delivery Network
type CDN interface {
	ThreatProtectionLite() (*NameServers, error)
	ConfigTemplate(isObfuscated bool, method string) (http.Header, []byte, error)
	GetRemoteFile(name string) ([]byte, error)
}

type CDNAPI struct {
	agent     string
	baseURL   string
	client    *http.Client
	validator response.Validator
	sync.Mutex
}

type CDNAPIResponse struct {
	Headers http.Header
	Body    io.ReadCloser
}

func NewCDNAPI(
	agent string,
	baseURL string,
	client *http.Client,
	validator response.Validator,
) *CDNAPI {
	return &CDNAPI{
		baseURL:   baseURL,
		agent:     agent,
		client:    client,
		validator: validator,
	}
}

func (api *CDNAPI) request(path, method string) (*CDNAPIResponse, error) {
	req, err := request.NewRequest(method, api.agent, api.baseURL, path, "", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}
	api.Lock()
	resp, err := api.client.Do(req)
	api.Unlock()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body []byte
	var reader io.ReadCloser
	if err == nil {
		switch method {
		case http.MethodHead:
			reader = io.NopCloser(bytes.NewReader(nil))
		default:
			switch resp.Header.Get("Content-Encoding") {
			case "gzip":
				reader, err = gzip.NewReader(resp.Body)
				if err != nil {
					return nil, err
				}
			default:
				reader = resp.Body
			}
		}
		body, err = MaxBytesReadAll(reader)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		reader = io.NopCloser(bytes.NewBuffer(body))

		if method == http.MethodHead {
			if !mandatoryHeadersExist(resp.Header) {
				return nil, fmt.Errorf("some of mandatory response headers do not exist")
			}
		} else {
			if err = api.validator.Validate(resp.StatusCode, resp.Header, body); err != nil {
				return nil, fmt.Errorf("cdn api: %w", err)
			}
		}
	}

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	return &CDNAPIResponse{
		Headers: resp.Header,
		Body:    reader,
	}, nil
}

func mandatoryHeadersExist(headers http.Header) bool {
	_, okAuth := headers["X-Authorization"]
	_, okDigest := headers["X-Digest"]
	_, okAccept := headers["X-Accept-Before"]
	_, okSign := headers["X-Signature"]
	return okAuth && okDigest && okAccept && okSign
}

func (api *CDNAPI) ConfigTemplate(isObfuscated bool, method string) (http.Header, []byte, error) {
	var path string
	if isObfuscated {
		path = ovpnObfsTemplateURL
	} else {
		path = ovpnTemplateURL
	}
	resp, err := api.request(path, method)
	if err != nil {
		return nil, nil, err
	}
	body, err := MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	return resp.Headers, body, nil
}

func (api *CDNAPI) ThreatProtectionLite() (*NameServers, error) {
	resp, err := api.request(threatProtectionLiteURL, http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var servers NameServers
	err = json.Unmarshal(body, &servers)
	if err != nil {
		return nil, err
	}

	return &servers, nil
}

func (api *CDNAPI) GetRemoteFile(name string) ([]byte, error) {
	resp, err := api.request(name, http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
