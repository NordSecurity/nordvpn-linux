/*
Package core provides Go HTTP client for interacting with Core API a.k.a. NordVPN API
*/
package core

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/request"
)

const (
	// linuxPlatformID defines the linux platform ID on the Notification Centre
	linuxPlatformID = 500
)

type CredentialsAPI interface {
	NotificationCredentials(token, appUserID string) (NotificationCredentialsResponse, error)
	ServiceCredentials(string) (*CredentialsResponse, error)
	TokenRenew(string) (*TokenRenewResponse, error)
	Services(string) (ServicesResponse, error)
	CurrentUser(string) (*CurrentUserResponse, error)
	DeleteToken(string) error
}

type InsightsAPI interface {
	Insights() (*Insights, error)
}

type ServersAPI interface {
	Servers() (Servers, http.Header, error)
	RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error)
	Server(id int64) (*Server, error)
	ServersCountries() (Countries, http.Header, error)
}

type DefaultAPI struct {
	version       string
	agent         string
	environment   internal.Environment
	pkVault       response.PKVault
	Client        *request.HTTPClient
	validatorFunc response.ValidatorFunc
	publisher     events.Publisher[events.DataRequestAPI]
	sync.Mutex
}

func NewDefaultAPI(
	version string,
	agent string,
	environment internal.Environment,
	pkVault response.PKVault,
	client *request.HTTPClient,
	validatorFunc response.ValidatorFunc,
	publisher events.Publisher[events.DataRequestAPI],
) *DefaultAPI {
	return &DefaultAPI{
		version:       version,
		agent:         agent,
		environment:   environment,
		pkVault:       pkVault,
		Client:        client,
		validatorFunc: validatorFunc,
		publisher:     publisher,
	}
}

func (api *DefaultAPI) Base() string {
	api.Lock()
	defer api.Unlock()
	return api.Client.BaseURL
}

func (api *DefaultAPI) request(path, method string, data []byte, auth *request.BasicAuth) (*http.Response, error) {
	req, err := request.NewRequestWithBasicAuth(method, api.agent, api.Client.BaseURL, path, "application/json", "", "gzip, deflate", bytes.NewBuffer(data), auth)
	if err != nil {
		return nil, err
	}

	return api.do(req, path)
}

// do request regardless of the authentication.
func (api *DefaultAPI) do(req *http.Request, endpoint string) (*http.Response, error) {
	resp, err := api.Client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var body []byte
	switch req.Method {
	case http.MethodHead:
		resp.Body = ioutil.NopCloser(bytes.NewReader(nil))
	default:
		// Decode response body if it is encoded
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = reader
		}
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err := api.validatorFunc(resp.Header, body, api.pkVault); err != nil {
		return nil, fmt.Errorf("validating headers: %w", err)
	}

	statusCode := resp.StatusCode
	err = ExtractError(resp)
	api.publisher.Publish(events.DataRequestAPI{
		Endpoint:      endpoint,
		Hostname:      api.Client.BaseURL,
		DNSResolution: 0,
		IsSuccessful:  err == nil,
		ResponseCode:  statusCode,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (api *DefaultAPI) Plans() (*Plans, error) {
	var ret *Plans
	req, err := request.NewRequest(http.MethodGet, api.agent, api.Client.BaseURL, PlanURL, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req, PlanURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// ServiceCredentials returns service credentials
func (api *DefaultAPI) ServiceCredentials(token string) (*CredentialsResponse, error) {
	resp, err := api.request(CredentialsURL, http.MethodGet, nil, &request.BasicAuth{Username: "token", Password: token})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret *CredentialsResponse
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// Services returns all previously and currently used services by the user
func (api *DefaultAPI) Services(token string) (ServicesResponse, error) {
	resp, err := api.request(ServicesURL, http.MethodGet, nil, &request.BasicAuth{Username: "token", Password: token})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret ServicesResponse
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// CreateUser accepts email and password as arguments
// and creates the user
func (api *DefaultAPI) CreateUser(email, password string) (*UserCreateResponse, error) {
	data, err := json.Marshal(UserCreateRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := request.NewRequest(http.MethodPost, api.agent, api.Client.BaseURL, UsersURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req, UsersURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret *UserCreateResponse
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// CurrentUser returns metadata of current user
func (api *DefaultAPI) CurrentUser(token string) (*CurrentUserResponse, error) {
	resp, err := api.request(CurrentUserURL, http.MethodGet, nil, &request.BasicAuth{Username: "token", Password: token})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret *CurrentUserResponse
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (api *DefaultAPI) DeleteToken(token string) error {
	resp, err := api.request(TokensURL, http.MethodDelete, nil, &request.BasicAuth{Username: "token", Password: token})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// TokenRenew queries the renew token and returns new token data
func (api *DefaultAPI) TokenRenew(token string) (*TokenRenewResponse, error) {
	if token == "" {
		return nil, ErrBadRequest
	}
	data, err := json.Marshal(struct {
		RenewToken string `json:"renewToken"`
	}{
		RenewToken: token,
	})
	if err != nil {
		return nil, err
	}

	req, err := request.NewRequest(http.MethodPost, api.agent, api.Client.BaseURL, TokenRenewURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req, TokenRenewURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret *TokenRenewResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// Servers returns servers list
func (api *DefaultAPI) Servers() (Servers, http.Header, error) {
	req, err := request.NewRequest(http.MethodGet, api.agent, api.Client.BaseURL, ServersURL+ServersURLConnectQuery, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := api.do(req, ServersURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var ret Servers
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, nil, err
	}
	return ret, resp.Header, nil
}

// ServersCountries returns server countries list
func (api *DefaultAPI) ServersCountries() (Countries, http.Header, error) {
	req, err := request.NewRequest(http.MethodGet, api.agent, api.Client.BaseURL, ServersCountriesURL, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := api.do(req, ServersCountriesURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var ret Countries
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, nil, err
	}
	return ret, resp.Header, nil
}

// RecommendedServers returns recommended servers list
func (api *DefaultAPI) RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error) {
	var filterQuery string
	switch filter.Tag.Action { //nolint:exhaustive // libmoose deprecates this
	case ServerBySpeed:
		// Set group filter from tag only if group flag is not defined
		if filter.Group == config.UndefinedGroup {
			filterQuery = fmt.Sprintf(RecommendedServersGroupsFilter, filter.Tag.ID)
		}
	case ServerByCountry:
		filterQuery = fmt.Sprintf(RecommendedServersCountryFilter, filter.Tag.ID)
	case ServerByCity:
		filterQuery = fmt.Sprintf(RecommendedServersCityFilter, filter.Tag.ID)
	default:
		filterQuery = ""
	}
	// When flag is defined append it to filter query
	if filter.Group != config.UndefinedGroup {
		filterQuery += fmt.Sprintf(RecommendedServersGroupsFilter, filter.Group)
	}

	url := RecommendedServersURL + fmt.Sprintf(RecommendedServersURLConnectQuery, filter.Limit, filter.Tech, longitude, latitude) + filterQuery
	req, err := request.NewRequest(http.MethodGet, api.agent, api.Client.BaseURL, url, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := api.do(req, RecommendedServersURL)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var ret Servers
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, nil, err
	}

	return ret, resp.Header, nil
}

// Server returns specific server
func (api *DefaultAPI) Server(id int64) (*Server, error) {
	req, err := request.NewRequest(http.MethodGet, api.agent, api.Client.BaseURL, ServersURL+fmt.Sprintf(ServersURLSpecificQuery, id), "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req, ServersURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret Servers
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}

	if len(ret) != 1 {
		return nil, fmt.Errorf("invalid response")
	}
	return &ret[0], nil
}

// Insights returns insights about user
func (api *DefaultAPI) Insights() (*Insights, error) {
	req, err := request.NewRequest(http.MethodGet, api.agent, api.Client.BaseURL, InsightsURL, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req, InsightsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ret Insights
	if err = json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

type NotificationCredentialsRequest struct {
	AppUserID  string `json:"app_user_uid"`
	PlatformID int    `json:"platform_id"`
}

type NotificationCredentialsResponse struct {
	Endpoint string `json:"endpoint"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// NotificationCredentials retrieves the credentials for notification center appUserID
func (api *DefaultAPI) NotificationCredentials(token, appUserID string) (NotificationCredentialsResponse, error) {
	data, err := json.Marshal(NotificationCredentialsRequest{
		AppUserID:  appUserID,
		PlatformID: linuxPlatformID,
	})
	if err != nil {
		return NotificationCredentialsResponse{}, fmt.Errorf("marshaling the request data: %w", err)
	}
	req, err := request.NewRequestWithBearerToken(http.MethodPost, api.agent, api.Client.BaseURL, notificationTokenURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data), token)
	if err != nil {
		return NotificationCredentialsResponse{}, fmt.Errorf("creating nc credentials request: %w", err)
	}
	rawResp, err := api.do(req, notificationTokenURL)
	if err != nil {
		return NotificationCredentialsResponse{}, fmt.Errorf("executing HTTP POST request: %w", err)
	}
	defer rawResp.Body.Close()
	out, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return NotificationCredentialsResponse{}, fmt.Errorf("reading HTTP response body: %w", err)
	}

	var resp NotificationCredentialsResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return NotificationCredentialsResponse{}, fmt.Errorf("unmarshaling HTTP response: %w", err)
	}
	return resp, nil
}

func (api *DefaultAPI) Logout(token string) error {
	resp, err := api.request(urlOAuth2Logout, http.MethodPost, nil, &request.BasicAuth{Username: "token", Password: token})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
