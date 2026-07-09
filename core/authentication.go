package core

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/daemon/response"
)

// Authentication is responsible for verifying user's identity.
type Authentication interface {
	Login(bool) (string, error)
	Token(string) (*LoginResponse, error)
}

type OAuth2 struct {
	baseURL   string
	client    *http.Client
	validator response.Validator
	// challenge is used to login
	challenge string
	// verifier is used to retrieve the token
	verifier string
	// attempt is used to retrieve the token
	attempt string
	sync.Mutex
}

func NewOAuth2(client *http.Client, baseURL string, validator response.Validator) *OAuth2 {
	return &OAuth2{
		baseURL:   baseURL,
		client:    client,
		validator: validator,
	}
}

func (o *OAuth2) Login(regularLogin bool) (string, error) {
	o.Lock()
	defer o.Unlock()

	path, err := url.Parse(o.baseURL + urlOAuth2Login)
	if err != nil {
		return "", err
	}

	o.verifier, o.challenge, err = newProofKeyPair(24)
	if err != nil {
		return "", err
	}

	preferredFlow := "registration"
	if regularLogin {
		preferredFlow = "login"
	}

	jsonBody, err := json.Marshal(loginBody{
		Challenge:     o.challenge,
		PreferredFlow: preferredFlow,
		RedirectFlow:  "default",
	})
	if err != nil {
		return "", err
	}

	reqBody := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, path.String(), reqBody)
	if err != nil {
		return "", err
	}

	body, err := o.doValidatedRequest(req, "validating login response")
	if err != nil {
		return "", err
	}

	var loginResp oAuth2LoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		return "", err
	}

	o.attempt = loginResp.Attempt
	return loginResp.URI, nil
}

// Token to be used for further API requests.
func (o *OAuth2) Token(exchangeToken string) (*LoginResponse, error) {
	o.Lock()
	defer o.Unlock()

	path, err := url.Parse(o.baseURL + urlOAuth2Token)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Add("attempt", o.attempt)
	query.Add("verifier", o.verifier)
	query.Add("exchange_token", exchangeToken)
	path.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		return nil, err
	}

	body, err := o.doValidatedRequest(req, "validating token response")
	if err != nil {
		return nil, err
	}

	var tokenResp LoginResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (o *OAuth2) doValidatedRequest(req *http.Request, errContext string) ([]byte, error) {
	if req.Method != http.MethodGet {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	body, err := MaxBytesReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := o.validator.Validate(resp.StatusCode, resp.Header, body); err != nil {
		return nil, fmt.Errorf("%s: %w", errContext, err)
	}

	return body, nil
}

// newProofKeyPair implements PKCE code pair generation from RFC7636.
func newProofKeyPair(length int) (string, string, error) {
	bs := make([]byte, length)
	_, err := rand.Read(bs)
	if err != nil {
		return "", "", err
	}

	verifier := hex.EncodeToString(bs)
	hasher := sha256.New()
	_, err = hasher.Write([]byte(verifier))
	if err != nil {
		return "", "", err
	}

	challenge := hex.EncodeToString(hasher.Sum(nil))
	return verifier, challenge, nil
}

type loginBody struct {
	Challenge     string `json:"challenge"`
	PreferredFlow string `json:"preferred_flow"`
	RedirectFlow  string `json:"redirect_flow"`
}
