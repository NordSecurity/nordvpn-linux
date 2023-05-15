package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/netip"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/NordSecurity/nordvpn-linux/request"

	"github.com/google/uuid"
)

const (
	// urlMeshRegister is used to register a single mesh machine/device.
	urlMeshRegister = "/v1/meshnet/machines"
	// urlMeshMachines is used to interact with a single mesh machine/device.
	urlMeshMachines = "/v1/meshnet/machines/%s"
	// urlMeshMachinesPeers is used to update peer e.g. if other peer can route through this peer/machine
	urlMeshMachinesPeers = "/v1/meshnet/machines/%s/peers/%s"
	// urlMeshMap is used to refresh libtelio.
	urlMeshMap = urlMeshMachines + "/map"
	// urlMeshPeers is used to interact with one's peers in the mesh network.
	urlMeshPeers = urlMeshMachines + "/peers"
	// urlMeshUnpair is used to unpair the invited peers.
	urlMeshUnpair = urlMeshMachines + "/peers/%s"
	// urlInvitationSend is used to invite other users to mesh network.
	urlInvitationSend = urlMeshMachines + "/invitations"
	// urlSentInvitationsList is used to view sent invitations.
	urlSentInvitationsList = urlInvitationSend + "/sent"
	// urlReceivedInvitationsList is used to view received invitations.
	urlReceivedInvitationsList = urlInvitationSend + "/received"
	// urlAcceptInvitation is used to accept an invitation.
	urlAcceptInvitation = urlInvitationSend + "/%s/accept"
	// urlRejectInvitation is used to reject an invitation.
	urlRejectInvitation = urlInvitationSend + "/%s/reject"
	// urlRevokeInvitation is used to revoke an invitation.
	urlRevokeInvitation = urlInvitationSend + "/%s"
	// urlNotifyFileTransfer is used to notify another peer about an incoming notification
	urlNotifyFileTransfer = urlMeshMachines + "/notifications/file-transfer"
	// logPatchMethodCall is used when formatting log messages for PATCH calls to API.
	logPatchMethodCall = "calling PATCH %s with %s"
	// logPostMethodCall is used when formatting log messages for POST calls to API.
	logPostMethodCall = "calling POST %s with %s"
)

var (
	// ErrPublicKeyNotProvided is returned when peer does not have a public key set.
	ErrPublicKeyNotProvided = errors.New("public key not provided")
	// ErrPeerOSNotProvided is returned when peer does not have os name or os version set.
	ErrPeerOSNotProvided = errors.New("os not provided")
	// ErrPeerEndpointsNotProvided is returned when peer has on endpoints.
	ErrPeerEndpointsNotProvided = errors.New("endpoints not provided")
)

// MeshAPI implements communication with Mesh part of Core team's backend.
type MeshAPI struct {
	base      string
	agent     string
	client    *request.HTTPClient
	vault     response.PKVault
	publisher events.Publisher[string]
	mu        sync.Mutex
}

// NewMeshAPI constructs a MeshAPI and returns a pointer to it.
func NewMeshAPI(
	base string,
	agent string,
	client *request.HTTPClient,
	vault response.PKVault,
	publisher events.Publisher[string],
) *MeshAPI {
	return &MeshAPI{
		base:      base,
		agent:     agent,
		client:    client,
		vault:     vault,
		publisher: publisher,
	}
}

func peersResponseToMachinePeers(rawPeers []mesh.MachinePeerResponse) []mesh.MachinePeer {
	peers := make([]mesh.MachinePeer, 0, len(rawPeers))
	for _, p := range rawPeers {
		var addr netip.Addr
		if len(p.Addresses) > 0 {
			addr = p.Addresses[0]
		}

		peers = append(peers, mesh.MachinePeer{

			ID:       p.ID,
			Hostname: p.Hostname,
			OS: mesh.OperatingSystem{
				Name:   p.OS,
				Distro: p.Distro,
			},
			PublicKey:                 p.PublicKey,
			Endpoints:                 p.Endpoints,
			Address:                   addr,
			Email:                     p.Email,
			IsLocal:                   p.IsLocal,
			DoesPeerAllowRouting:      p.DoesPeerAllowRouting,
			DoesPeerAllowInbound:      p.DoesPeerAllowInbound,
			DoesPeerAllowLocalNetwork: p.DoesPeerAllowLocalNetwork,
			DoesPeerAllowFileshare:    p.DoesPeerAllowFileshare,
			DoesPeerSupportRouting:    p.DoesPeerSupportRouting,
			DoIAllowRouting:           p.DoIAllowRouting,
			DoIAllowInbound:           p.DoIAllowInbound,
			DoIAllowLocalNetwork:      p.DoIAllowLocalNetwork,
			DoIAllowFileshare:         p.DoIAllowFileshare,
		})
	}

	return peers
}

// Register peer to the mesh network.
func (m *MeshAPI) Register(token string, peer mesh.Machine) (*mesh.Machine, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if peer.PublicKey == "" {
		return nil, ErrPublicKeyNotProvided
	}

	if peer.OS.Name == "" || peer.OS.Distro == "" {
		return nil, ErrPeerOSNotProvided
	}

	data, err := json.Marshal(mesh.MachineCreateRequest{
		PublicKey:       peer.PublicKey,
		HardwareID:      peer.HardwareID,
		OS:              peer.OS.Name,
		Distro:          peer.OS.Distro,
		Endpoints:       peer.Endpoints,
		SupportsRouting: peer.SupportsRouting,
	})
	if err != nil {
		return nil, err
	}
	m.publisher.Publish(fmt.Sprintf(logPostMethodCall, urlMeshRegister, string(data)))
	req, err := request.NewRequestWithBearerToken(
		http.MethodPost,
		m.agent,
		m.base,
		urlMeshRegister,
		"application/json",
		"",
		"",
		bytes.NewBuffer(data),
		token,
	)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	var raw mesh.MachineCreateResponse
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	m.publisher.Publish(fmt.Sprintf("received response %s", string(body)))

	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	if len(raw.Addresses) < 1 {
		return nil, errors.New("invalid response")
	}

	return &mesh.Machine{
		ID:        raw.Identifier,
		Hostname:  raw.Hostname,
		OS:        peer.OS,
		PublicKey: peer.PublicKey,
		Endpoints: raw.Endpoints,
		Address:   raw.Addresses[0],
	}, nil
}

// Update publishes new endpoints.
func (m *MeshAPI) Update(token string, id uuid.UUID, endpoints []netip.AddrPort) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(endpoints) == 0 {
		return ErrPeerEndpointsNotProvided
	}

	data, err := json.Marshal(mesh.MachineUpdateRequest{
		Endpoints:       endpoints,
		SupportsRouting: true,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf(urlMeshMachines, id.String())
	m.publisher.Publish(fmt.Sprintf(logPatchMethodCall, url, string(data)))
	req, err := request.NewRequestWithBearerToken(
		http.MethodPatch,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		bytes.NewBuffer(data),
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Configure interaction with a specific peer.
func (m *MeshAPI) Configure(
	token string,
	id uuid.UUID,
	peerID uuid.UUID,
	doIAllowInbound bool,
	doIAllowRouting bool,
	doIAllowLocalNetwork bool,
	doIAllowFileshare bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(mesh.PeerUpdateRequest{
		DoIAllowInbound:      doIAllowInbound,
		DoIAllowRouting:      doIAllowRouting,
		DoIAllowLocalNetwork: doIAllowLocalNetwork,
		DoIAllowFileshare:    doIAllowFileshare,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf(urlMeshMachinesPeers, id.String(), peerID.String())
	m.publisher.Publish(fmt.Sprintf(logPatchMethodCall, url, string(data)))
	req, err := request.NewRequestWithBearerToken(
		http.MethodPatch,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		bytes.NewBuffer(data),
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Unregister peer from the mesh network.
func (m *MeshAPI) Unregister(token string, self uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlMeshMachines, self.String())
	m.publisher.Publish("calling DELETE " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodDelete,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return ExtractError(resp)
}

func peersResponseToLocalPeers(rawPeers []mesh.MachinePeerResponse) []mesh.Machine {
	peers := make([]mesh.Machine, 0, len(rawPeers))

	for _, p := range rawPeers {
		var addr netip.Addr
		if len(p.Addresses) > 0 {
			addr = p.Addresses[0]
		}

		peers = append(peers, mesh.Machine{
			ID:       p.ID,
			Hostname: p.Hostname,
			OS: mesh.OperatingSystem{
				Name: p.OS, Distro: p.Distro,
			},
			PublicKey: p.PublicKey,
			Endpoints: p.Endpoints,
			Address:   addr,
		})
	}

	return peers
}

// Local peer list.
func (m *MeshAPI) Local(token string) (mesh.Machines, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.publisher.Publish("calling GET " + m.base + urlMeshMachines)
	req, err := request.NewRequestWithBearerToken(
		http.MethodGet,
		m.agent,
		m.base,
		urlMeshMachines,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m.publisher.Publish(fmt.Sprintf("received response %s", string(body)))

	var rawPeers []mesh.MachinePeerResponse
	err = json.Unmarshal(body, &rawPeers)
	if err != nil {
		return nil, err
	}

	peers := peersResponseToLocalPeers(rawPeers)

	return peers, nil
}

func (m *MeshAPI) Map(token string, self uuid.UUID) (*mesh.MachineMap, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlMeshMap, self.String())
	m.publisher.Publish("calling GET " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodGet,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m.publisher.Publish(fmt.Sprintf("received response %s", string(body)))

	var raw mesh.MachineMapResponse
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, err
	}

	peers := peersResponseToMachinePeers(raw.Peers)

	return &mesh.MachineMap{
		Machine: mesh.Machine{
			ID:        raw.ID,
			Hostname:  raw.Hostname,
			PublicKey: raw.PublicKey,
			Endpoints: raw.Endpoints,
			Address:   raw.Addresses[0],
		},
		Hosts: raw.DNS.Hosts,
		Peers: peers,
		Raw:   body,
	}, nil
}

// List peers in the mesh network for a given peer.
func (m *MeshAPI) List(token string, self uuid.UUID) (mesh.MachinePeers, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlMeshPeers, self.String())
	m.publisher.Publish("calling GET " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodGet,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m.publisher.Publish(fmt.Sprintf("received response %s", string(body)))

	var rawPeers []mesh.MachinePeerResponse
	err = json.Unmarshal(body, &rawPeers)
	if err != nil {
		return nil, err
	}

	peers := peersResponseToMachinePeers(rawPeers)

	return peers, nil
}

// Unpair a given peer.
func (m *MeshAPI) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlMeshUnpair, self.String(), peer.String())
	m.publisher.Publish("calling DELETE " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodDelete,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Invite to mesh.
func (m *MeshAPI) Invite(
	token string,
	self uuid.UUID,
	email string,
	doIAllowInbound bool,
	doIAllowRouting bool,
	doIAllowLocalNetwork bool,
	doIAllowFileshare bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(&mesh.SendInvitationRequest{
		Email:             email,
		AllowInbound:      doIAllowInbound,
		AllowRouting:      doIAllowRouting,
		AllowLocalNetwork: doIAllowLocalNetwork,
		AllowFileshare:    doIAllowFileshare,
	})
	if err != nil {
		return err
	}
	url := fmt.Sprintf(urlInvitationSend, self.String())

	m.publisher.Publish(fmt.Sprintf(logPostMethodCall, url, string(data)))
	req, err := request.NewRequestWithBearerToken(
		http.MethodPost,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		bytes.NewBuffer(data),
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Received invitations from other users.
func (m *MeshAPI) Received(token string, self uuid.UUID) (mesh.Invitations, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlReceivedInvitationsList, self.String())
	m.publisher.Publish("calling GET " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodGet,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m.publisher.Publish(fmt.Sprintf("received response %s", string(body)))

	var invitations mesh.Invitations
	err = json.Unmarshal(body, &invitations)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// Sent invitations to other users.
func (m *MeshAPI) Sent(token string, self uuid.UUID) (mesh.Invitations, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlSentInvitationsList, self.String())
	m.publisher.Publish("calling GET " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodGet,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return nil, err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := ExtractError(resp); err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m.publisher.Publish(fmt.Sprintf("received response %s", string(body)))

	var invitations mesh.Invitations
	err = json.Unmarshal(body, &invitations)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// Accept invitation.
func (m *MeshAPI) Accept(
	token string,
	self uuid.UUID,
	invitation uuid.UUID,
	doIAllowInbound bool,
	doIAllowRouting bool,
	doIAllowLocalNetwork bool,
	doIAllowFileshare bool,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.Marshal(&mesh.AcceptInvitationRequest{
		AllowInbound:      doIAllowInbound,
		AllowRouting:      doIAllowRouting,
		AllowLocalNetwork: doIAllowLocalNetwork,
		AllowFileshare:    doIAllowFileshare,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf(urlAcceptInvitation, self.String(), invitation.String())
	m.publisher.Publish(fmt.Sprintf(logPostMethodCall, url, string(data)))
	req, err := request.NewRequestWithBearerToken(
		http.MethodPost,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		bytes.NewReader(data),
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Reject invitation.
func (m *MeshAPI) Reject(token string, self uuid.UUID, invitation uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlRejectInvitation, self.String(), invitation.String())
	m.publisher.Publish("calling POST " + m.base + url)
	req, err := request.NewRequestWithBearerToken(http.MethodPost, m.agent, m.base, url, "application/json", "", "", nil, token)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Revoke invitation.
func (m *MeshAPI) Revoke(token string, self uuid.UUID, invitation uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlRevokeInvitation, self.String(), invitation.String())
	m.publisher.Publish("calling DELETE " + m.base + url)
	req, err := request.NewRequestWithBearerToken(
		http.MethodDelete,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		nil,
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}

// Notify peer about a new incoming transfer
func (m *MeshAPI) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	url := fmt.Sprintf(urlNotifyFileTransfer, self.String())

	dataUnmarshaled := mesh.NotificationNewTransactionRequest{
		ReceiverMachineIdentifier: peer.String(),
		FileCount:                 fileCount,
	}
	m.publisher.Publish(fmt.Sprintf(logPostMethodCall, url, fmt.Sprintf("%+v", dataUnmarshaled)))
	dataUnmarshaled.FileName = fileName // We must not log filenames, so setting it after log
	data, err := json.Marshal(dataUnmarshaled)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := request.NewRequestWithBearerToken(
		http.MethodPost,
		m.agent,
		m.base,
		url,
		"application/json",
		"",
		"",
		bytes.NewReader(data),
		token,
	)
	if err != nil {
		return err
	}

	resp, err := m.client.DoRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return ExtractError(resp)
}
