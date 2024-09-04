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
	"net/http"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/request"
)

const (
	// linuxPlatformID defines the linux platform ID on the Notification Centre
	linuxPlatformID = 500
)

type CredentialsAPI interface {
	NotificationCredentials(token, appUserID string) (NotificationCredentialsResponse, error)
	NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error)
	ServiceCredentials(string) (*CredentialsResponse, error)
	TokenRenew(string) (*TokenRenewResponse, error)
	Services(string) (ServicesResponse, error)
	CurrentUser(string) (*CurrentUserResponse, error)
	DeleteToken(string) error
	TrustedPassToken(string) (*TrustedPassTokenResponse, error)
	MultifactorAuthStatus(string) (*MultifactorAuthStatusResponse, error)
	Logout(token string) error
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

type CombinedAPI interface {
	InsightsAPI
	Base() string
	Plans() (*Plans, error)
	CreateUser(email, password string) (*UserCreateResponse, error)
}

type DefaultAPI struct {
	agent     string
	baseURL   string
	client    *http.Client
	validator response.Validator
	mu        sync.Mutex
}

func NewDefaultAPI(
	agent string,
	baseURL string,
	client *http.Client,
	validator response.Validator,
) *DefaultAPI {
	return &DefaultAPI{
		agent:     agent,
		baseURL:   baseURL,
		client:    client,
		validator: validator,
	}
}

func (api *DefaultAPI) Base() string {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.baseURL
}

func (api *DefaultAPI) request(path, method string, data []byte, token string) (*http.Response, error) {
	req, err := request.NewRequestWithBearerToken(method, api.agent, api.baseURL, path, "application/json", "", "gzip, deflate", bytes.NewBuffer(data), token)
	if err != nil {
		return nil, err
	}

	return api.do(req)
}

// do request regardless of the authentication.
func (api *DefaultAPI) do(req *http.Request) (*http.Response, error) {
	resp, err := api.client.Do(req)

	// Transport of the request is already up to date

	if err != nil {
		return nil, err
	}

	switch req.Method {
	case http.MethodHead:
		resp.Body = io.NopCloser(bytes.NewReader(nil))
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

	defer resp.Body.Close()

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	if err := api.validator.Validate(resp.StatusCode, resp.Header, body); err != nil {
		return nil, fmt.Errorf("validating headers: %w", err)
	}

	err = ExtractError(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (api *DefaultAPI) Plans() (*Plans, error) {
	var ret *Plans
	req, err := request.NewRequest(http.MethodGet, api.agent, api.baseURL, PlanURL, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req)
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
	resp, err := api.request(CredentialsURL, http.MethodGet, nil, token)
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
	resp, err := api.request(ServicesURL, http.MethodGet, nil, token)
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

	req, err := request.NewRequest(http.MethodPost, api.agent, api.baseURL, UsersURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req)
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
	resp, err := api.request(CurrentUserURL, http.MethodGet, nil, token)
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
	resp, err := api.request(TokensURL, http.MethodDelete, nil, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// TokenRenew queries the renew token and returns new token data
func (api *DefaultAPI) TrustedPassToken(token string) (*TrustedPassTokenResponse, error) {
	resp, err := api.request(TrustedPassTokenURL, http.MethodPost, nil, token)
	if err != nil {
		return nil, fmt.Errorf("making api request: %w", err)
	}
	defer resp.Body.Close()

	var ret *TrustedPassTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	return ret, nil
}

// MultifactorAuthStatus queries and returns the status of MFA
func (api *DefaultAPI) MultifactorAuthStatus(token string) (*MultifactorAuthStatusResponse, error) {
	resp, err := api.request(MFAStatusURL, http.MethodGet, nil, token)
	if err != nil {
		return nil, fmt.Errorf("making api request: %w", err)
	}
	defer resp.Body.Close()

	var ret *MultifactorAuthStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&ret); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}

	return ret, nil
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

	req, err := request.NewRequest(http.MethodPost, api.agent, api.baseURL, TokenRenewURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req)
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
	req, err := request.NewRequest(http.MethodGet, api.agent, api.baseURL, ServersURL+ServersURLConnectQuery, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := api.do(req)
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
	req, err := request.NewRequest(http.MethodGet, api.agent, api.baseURL, ServersCountriesURL, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := api.do(req)
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
		if filter.Group == config.ServerGroup_UNDEFINED {
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
	if filter.Group != config.ServerGroup_UNDEFINED {
		filterQuery += fmt.Sprintf(RecommendedServersGroupsFilter, filter.Group)
	}

	url := RecommendedServersURL + fmt.Sprintf(RecommendedServersURLConnectQuery, filter.Limit, filter.Tech, longitude, latitude) + filterQuery
	req, err := request.NewRequest(http.MethodGet, api.agent, api.baseURL, url, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err := api.do(req)
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
	req, err := request.NewRequest(http.MethodGet, api.agent, api.baseURL, ServersURL+fmt.Sprintf(ServersURLSpecificQuery, id), "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req)
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
	req, err := request.NewRequest(http.MethodGet, api.agent, api.baseURL, InsightsURL, "application/json", "", "gzip, deflate", nil)
	if err != nil {
		return nil, err
	}

	resp, err := api.do(req)
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

type NotificationCredentialsRevokeRequest struct {
	AppUserID    string `json:"app_user_uid"`
	PurgeSession bool   `json:"purge_session"`
}

type NotificationCredentialsRevokeResponse struct {
	Status string `json:"status"`
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
	req, err := request.NewRequestWithBearerToken(http.MethodPost, api.agent, api.baseURL, notificationTokenURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data), token)
	if err != nil {
		return NotificationCredentialsResponse{}, fmt.Errorf("creating nc credentials request: %w", err)
	}
	rawResp, err := api.do(req)
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
		return NotificationCredentialsResponse{}, fmt.Errorf("unmarshalling HTTP response: %w", err)
	}
	return resp, nil
}

// NotificationCredentialsRevoke revokes the credentials for notification center appUserID
func (api *DefaultAPI) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error) {
	// Calling tokens/revoke endpoint with just bearer token will revoke user credentials for every user device.
	// For example, if user has VPN app for android/iOS/mac, whatever, all of his/her devices will be disconnected.
	// If you provide additionally app_user_id, then only credential for specific app/device will be revoked.
	// Connection on other devices will stay unaffected.
	// The purge_session param make sense only in cases when you definitely know, that app_user_id was generated
	// just for one time usage and after this usage it is not needed anymore. The good example is exactly tests,
	// where on each run you generate different app_user_id. In usual scenarios in ideal case, app_user_id on the
	// same device for the same user and same app should stay constant, even if app were reinstalled. So there is no
	// need to use purge_session at all, and it even can have a bad consequences.

	if appUserID == "" {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("refusing to send a request with empty appUserID")
	}

	data, err := json.Marshal(NotificationCredentialsRevokeRequest{
		AppUserID:    appUserID,
		PurgeSession: purgeSession,
	})
	if err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("marshaling the request data: %w", err)
	}
	req, err := request.NewRequestWithBearerToken(http.MethodPost, api.agent, api.baseURL, notificationTokenRevokeURL, "application/json", "", "gzip, deflate", bytes.NewBuffer(data), token)
	if err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("creating nc credentials revoke request: %w", err)
	}
	rawResp, err := api.do(req)
	if err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("executing HTTP POST request: %w", err)
	}
	defer rawResp.Body.Close()
	out, err := io.ReadAll(rawResp.Body)
	if err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("reading HTTP response body: %w", err)
	}

	var resp NotificationCredentialsRevokeResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("unmarshalling HTTP response: %w", err)
	}
	return resp, nil
}

func (api *DefaultAPI) Logout(token string) error {
	resp, err := api.request(urlOAuth2Logout, http.MethodPost, nil, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
