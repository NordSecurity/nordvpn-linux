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
	"github.com/google/uuid"
)

type SimpleCredentialsAPI interface {
	NotificationCredentials(token, appUserID string) (NotificationCredentialsResponse, error)
	NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error)
	ServiceCredentials(string) (*CredentialsResponse, error)
	TokenRenew(token string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error)
	Services(string) (ServicesResponse, error)
	CurrentUser(string) (*CurrentUserResponse, error)
	DeleteToken(string) error
	TrustedPassToken(string) (*TrustedPassTokenResponse, error)
	MultifactorAuthStatus(string) (*MultifactorAuthStatusResponse, error)
	Logout(token string) error
}

type SimpleInsightsAPI interface {
	Insights() (*Insights, error)
}

type SimpleServersAPI interface {
	Servers() (Servers, http.Header, error)
	RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error)
	Server(id int64) (*Server, error)
	ServersCountries() (Countries, http.Header, error)
}

type SimpleCombinedAPI interface {
	SimpleInsightsAPI
	Base() string
	Plans() (*Plans, error)
	CreateUser(email, password string) (*UserCreateResponse, error)
}

// SubscriptionAPI is responsible for fetching the subscription data of the user
type SimpleSubscriptionAPI interface {
	// Orders returns a list of orders done by the user
	Orders(token string) ([]Order, error)
	// Payments returns a list of payments done by the user
	Payments(token string) ([]PaymentResponse, error)
}

type SimpleClientAPI struct {
	agent     string
	baseURL   string
	client    *http.Client
	validator response.Validator
	mu        sync.Mutex
}

func NewSimpleAPI(
	agent string,
	baseURL string,
	client *http.Client,
	validator response.Validator,
) *SimpleClientAPI {
	return &SimpleClientAPI{
		agent:     agent,
		baseURL:   baseURL,
		client:    client,
		validator: validator,
	}
}

func (api *SimpleClientAPI) Base() string {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.baseURL
}

func (api *SimpleClientAPI) request(path, method string, data []byte, token string) (*http.Response, error) {
	req, err := request.NewRequestWithBearerToken(method, api.agent, api.baseURL, path, "application/json", "", "gzip, deflate", bytes.NewBuffer(data), token)
	if err != nil {
		return nil, err
	}

	return api.do(req)
}

// do request regardless of the authentication.
func (api *SimpleClientAPI) do(req *http.Request) (*http.Response, error) {
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
	body, err = MaxBytesReadAll(resp.Body)
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

func (api *SimpleClientAPI) Plans() (*Plans, error) {
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
func (api *SimpleClientAPI) ServiceCredentials(token string) (*CredentialsResponse, error) {
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
func (api *SimpleClientAPI) Services(token string) (ServicesResponse, error) {
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
func (api *SimpleClientAPI) CreateUser(email, password string) (*UserCreateResponse, error) {
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
func (api *SimpleClientAPI) CurrentUser(token string) (*CurrentUserResponse, error) {
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

func (api *SimpleClientAPI) DeleteToken(token string) error {
	resp, err := api.request(TokensURL, http.MethodDelete, nil, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// TokenRenew queries the renew token and returns new token data
func (api *SimpleClientAPI) TrustedPassToken(token string) (*TrustedPassTokenResponse, error) {
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
func (api *SimpleClientAPI) MultifactorAuthStatus(token string) (*MultifactorAuthStatusResponse, error) {
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
func (api *SimpleClientAPI) TokenRenew(token string, idempotencyKey uuid.UUID) (*TokenRenewResponse, error) {
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
	req.Header.Add("Idempotency-Key", idempotencyKey.String())

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
func (api *SimpleClientAPI) Servers() (Servers, http.Header, error) {
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

	// if at least one record is not valid - reject whole list, assuming something wrong is with whole list
	if err = ret.Validate(); err != nil {
		return nil, nil, err
	}

	return ret, resp.Header, nil
}

// ServersCountries returns server countries list
func (api *SimpleClientAPI) ServersCountries() (Countries, http.Header, error) {
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
func (api *SimpleClientAPI) RecommendedServers(filter ServersFilter, longitude, latitude float64) (Servers, http.Header, error) {
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

	// if at least one record is not valid - reject whole list, assuming something wrong is with whole list
	if err = ret.Validate(); err != nil {
		return nil, nil, err
	}

	return ret, resp.Header, nil
}

// Server returns specific server
func (api *SimpleClientAPI) Server(id int64) (*Server, error) {
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
func (api *SimpleClientAPI) Insights() (*Insights, error) {
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
func (api *SimpleClientAPI) NotificationCredentials(token, appUserID string) (NotificationCredentialsResponse, error) {
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
	out, err := MaxBytesReadAll(rawResp.Body)
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
func (api *SimpleClientAPI) NotificationCredentialsRevoke(token, appUserID string, purgeSession bool) (NotificationCredentialsRevokeResponse, error) {
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
	out, err := MaxBytesReadAll(rawResp.Body)
	if err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("reading HTTP response body: %w", err)
	}

	var resp NotificationCredentialsRevokeResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return NotificationCredentialsRevokeResponse{}, fmt.Errorf("unmarshalling HTTP response: %w", err)
	}
	return resp, nil
}

func (api *SimpleClientAPI) Logout(token string) error {
	resp, err := api.request(urlOAuth2Logout, http.MethodPost, nil, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (api *SimpleClientAPI) Orders(token string) ([]Order, error) {
	return getData[[]Order](api, token, urlOrders)
}

func (api *SimpleClientAPI) Payments(token string) ([]PaymentResponse, error) {
	return getData[[]PaymentResponse](api, token, urlPayments)
}

// getData calls a HTTP get request for the endpoints requiring authentication and returns the
// requested data.
func getData[T any](api *SimpleClientAPI, token string, url string) (T, error) {
	var data T
	resp, err := api.request(url, http.MethodGet, nil, token)
	if err != nil {
		return data, fmt.Errorf("executing HTTP GET request: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("decoding data from JSON: %w", err)
	}

	return data, nil
}

// // Register peer to the mesh network.
// func (api *SimpleClientAPI) Register(token string, peer mesh.Machine) (*mesh.Machine, error) {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	if peer.PublicKey == "" {
// 		return nil, ErrPublicKeyNotProvided
// 	}

// 	if peer.OS.Name == "" || peer.OS.Distro == "" {
// 		return nil, ErrPeerOSNotProvided
// 	}

// 	data, err := json.Marshal(mesh.MachineCreateRequest{
// 		PublicKey:       peer.PublicKey,
// 		HardwareID:      peer.HardwareID,
// 		OS:              peer.OS.Name,
// 		Distro:          peer.OS.Distro,
// 		Endpoints:       peer.Endpoints,
// 		SupportsRouting: peer.SupportsRouting,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, err := api.request(
// 		urlMeshRegister,
// 		http.MethodPost,
// 		data,
// 		token,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	if err := ExtractError(resp); err != nil {
// 		return nil, err
// 	}

// 	var raw mesh.MachineCreateResponse
// 	body, err := MaxBytesReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = json.Unmarshal(body, &raw)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if len(raw.Addresses) < 1 {
// 		return nil, errors.New("invalid response")
// 	}

// 	var addr netip.Addr
// 	if len(raw.Addresses) > 0 {
// 		addr = raw.Addresses[0]
// 	}

// 	return &mesh.Machine{
// 		ID:              raw.Identifier,
// 		Hostname:        raw.Hostname,
// 		OS:              peer.OS,
// 		PublicKey:       peer.PublicKey,
// 		Endpoints:       raw.Endpoints,
// 		Address:         addr,
// 		Nickname:        raw.Nickname,
// 		SupportsRouting: raw.SupportsRouting,
// 	}, nil
// }

// // Update publishes new endpoints.
// func (api *SimpleClientAPI) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	data, err := json.Marshal(info)
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := api.request(
// 		fmt.Sprintf(urlMeshMachines, id.String()),
// 		http.MethodPatch,
// 		data,
// 		token,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Configure interaction with a specific peer.
// func (api *SimpleClientAPI) Configure(
// 	token string,
// 	id uuid.UUID,
// 	peerID uuid.UUID,
// 	peerUpdateInfo mesh.PeerUpdateRequest,
// ) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	data, err := json.Marshal(peerUpdateInfo)
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := api.request(
// 		fmt.Sprintf(urlMeshMachinesPeers, id.String(), peerID.String()),
// 		http.MethodPatch,
// 		data,
// 		token,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Unregister peer from the mesh network.
// func (api *SimpleClientAPI) Unregister(token string, self uuid.UUID) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlMeshMachines, self.String()),
// 		http.MethodDelete,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()
// 	return ExtractError(resp)
// }

// func (api *SimpleClientAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlMeshMap, self.String()),
// 		http.MethodGet,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if err := ExtractError(resp); err != nil {
// 		return nil, err
// 	}

// 	body, err := MaxBytesReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var raw mesh.MachineMapResponse
// 	err = json.Unmarshal(body, &raw)
// 	if err != nil {
// 		return nil, err
// 	}

// 	peers := peersResponseToMachinePeers(raw.Peers)

// 	var addr netip.Addr
// 	if len(raw.Addresses) > 0 {
// 		addr = raw.Addresses[0]
// 	}

// 	return &mesh.MachineMap{
// 		Machine: mesh.Machine{
// 			ID:              raw.ID,
// 			Hostname:        raw.Hostname,
// 			PublicKey:       raw.PublicKey,
// 			Endpoints:       raw.Endpoints,
// 			Address:         addr,
// 			Nickname:        raw.Nickname,
// 			SupportsRouting: raw.SupportsRouting,
// 		},
// 		Hosts: raw.DNS.Hosts,
// 		Peers: peers,
// 		Raw:   body,
// 	}, nil
// }

// // List peers in the mesh network for a given peer.
// func (api *SimpleClientAPI) List(token string, self uuid.UUID) (mesh.MachinePeers, error) {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlMeshPeers, self.String()),
// 		http.MethodGet,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if err := ExtractError(resp); err != nil {
// 		return nil, err
// 	}

// 	body, err := MaxBytesReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var rawPeers []mesh.MachinePeerResponse
// 	err = json.Unmarshal(body, &rawPeers)
// 	if err != nil {
// 		return nil, err
// 	}

// 	peers := peersResponseToMachinePeers(rawPeers)

// 	return peers, nil
// }

// // Unpair a given peer.
// func (api *SimpleClientAPI) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlMeshUnpair, self.String(), peer.String()),
// 		http.MethodDelete,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Invite to mesh.
// func (api *SimpleClientAPI) Invite(
// 	token string,
// 	self uuid.UUID,
// 	email string,
// 	doIAllowInbound bool,
// 	doIAllowRouting bool,
// 	doIAllowLocalNetwork bool,
// 	doIAllowFileshare bool,
// ) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	data, err := json.Marshal(&mesh.SendInvitationRequest{
// 		Email:             email,
// 		AllowInbound:      doIAllowInbound,
// 		AllowRouting:      doIAllowRouting,
// 		AllowLocalNetwork: doIAllowLocalNetwork,
// 		AllowFileshare:    doIAllowFileshare,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := api.request(
// 		fmt.Sprintf(urlInvitationSend, self.String()),
// 		http.MethodPost,
// 		data,
// 		token,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Received invitations from other users.
// func (api *SimpleClientAPI) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlReceivedInvitationsList, self.String()),
// 		http.MethodGet,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if err := ExtractError(resp); err != nil {
// 		return nil, err
// 	}

// 	body, err := MaxBytesReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var invitations mesh.Invitations
// 	err = json.Unmarshal(body, &invitations)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return invitations, nil
// }

// // Sent invitations to other users.
// func (api *SimpleClientAPI) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlSentInvitationsList, self.String()),
// 		http.MethodGet,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if err := ExtractError(resp); err != nil {
// 		return nil, err
// 	}

// 	body, err := MaxBytesReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var invitations mesh.Invitations
// 	err = json.Unmarshal(body, &invitations)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return invitations, nil
// }

// // Accept invitation.
// func (api *SimpleClientAPI) Accept(
// 	token string,
// 	self uuid.UUID,
// 	invitation uuid.UUID,
// 	doIAllowInbound bool,
// 	doIAllowRouting bool,
// 	doIAllowLocalNetwork bool,
// 	doIAllowFileshare bool,
// ) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	data, err := json.Marshal(&mesh.AcceptInvitationRequest{
// 		AllowInbound:      doIAllowInbound,
// 		AllowRouting:      doIAllowRouting,
// 		AllowLocalNetwork: doIAllowLocalNetwork,
// 		AllowFileshare:    doIAllowFileshare,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	resp, err := api.request(
// 		fmt.Sprintf(urlAcceptInvitation, self.String(), invitation.String()),
// 		http.MethodPost,
// 		data,
// 		token,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Reject invitation.
// func (api *SimpleClientAPI) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlRejectInvitation, self.String(), invitation.String()),
// 		http.MethodPost,
// 		nil,
// 		token,
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Revoke invitation.
// func (api *SimpleClientAPI) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	resp, err := api.request(
// 		fmt.Sprintf(urlRevokeInvitation, self.String(), invitation.String()),
// 		http.MethodDelete,
// 		nil,
// 		token,
// 	)

// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }

// // Notify peer about a new incoming transfer
// func (api *SimpleClientAPI) NotifyNewTransfer(
// 	token string,
// 	self uuid.UUID,
// 	peer uuid.UUID,
// 	fileName string,
// 	fileCount int,
// 	transferID string,
// ) error {
// 	api.mu.Lock()
// 	defer api.mu.Unlock()

// 	dataUnmarshaled := mesh.NotificationNewTransactionRequest{
// 		ReceiverMachineIdentifier: peer.String(),
// 		FileCount:                 fileCount,
// 		TransferID:                transferID,
// 	}
// 	dataUnmarshaled.FileName = fileName // We must not log filenames, so setting it after log
// 	data, err := json.Marshal(dataUnmarshaled)
// 	if err != nil {
// 		return fmt.Errorf("marshaling request: %w", err)
// 	}

// 	resp, err := api.request(
// 		fmt.Sprintf(urlNotifyFileTransfer, self.String()),
// 		http.MethodPost,
// 		data,
// 		token,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	// 500 Internal Server Error is returned when peer machine have not registered its app_user_uid
// 	// Not all platforms implemented it yet, so suppress that error to not clutter logs
// 	if errors.Is(err, ErrServerInternal) {
// 		return nil
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return ExtractError(resp)
// }
