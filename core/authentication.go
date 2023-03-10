package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/request"
)

// Authentication is responsible for verifying user's identity.
type Authentication interface {
	Login() (string, error)
	Token(string) (*LoginResponse, error)
}

type OAuth2 struct {
	client *request.HTTPClient
	// challenge is used to login
	challenge string
	// verifier is used to retrieve the token
	verifier string
	// attempt is used to retrieve the token
	attempt string
	sync.Mutex
}

func NewOAuth2(client *request.HTTPClient) *OAuth2 {
	return &OAuth2{client: client}
}

func (o *OAuth2) Login() (string, error) {
	o.Lock()
	defer o.Unlock()

	path, err := url.Parse(o.client.BaseURL + urlOAuth2Login)
	if err != nil {
		return "", err
	}

	o.verifier, o.challenge, err = newProofKeyPair(24)
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Add("challenge", o.challenge)
	query.Add("preferred_flow", "login")
	query.Add("redirect_flow", "default")
	path.RawQuery = query.Encode()
	log.Println("oauth2 login url", path.String())

	req, err := http.NewRequest(http.MethodPost, path.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = ExtractError(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response oAuth2LoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	o.attempt = response.Attempt
	return response.URI, nil
}

// Token to be used for further API requests.
func (o *OAuth2) Token(exchangeToken string) (*LoginResponse, error) {
	o.Lock()
	defer o.Unlock()

	path, err := url.Parse(o.client.BaseURL + urlOAuth2Token)
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
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = ExtractError(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response LoginResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
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
