package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/response"
	"github.com/NordSecurity/nordvpn-linux/request"
	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMeshAPI_Register(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []testCase{
		testNewCase(t, http.StatusOK, urlMeshRegister, "mesh_register", nil),
		testNewCase(t, http.StatusBadRequest, urlMeshRegister, "mesh_register", ErrMaximumDeviceCount),
		testNewCase(t, http.StatusUnauthorized, urlMeshRegister, "mesh_register", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, urlMeshRegister, "mesh_register", ErrForbidden),
		testNewCase(t, http.StatusConflict, urlMeshRegister, "mesh_register", ErrConflict),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.Register("bearer", mesh.Machine{
				ID:        uuid.New(),
				PublicKey: uuid.New().String(),
				OS: mesh.OperatingSystem{
					Name:   "linux",
					Distro: "Arch",
				},
			})
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Update(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	url := fmt.Sprintf(urlMeshMachines, id.String())
	tests := []testCase{
		testNewCase(t, http.StatusOK, url, "mesh_update", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_update", ErrBadRequest),
		testNewCase(t, http.StatusUnauthorized, url, "mesh_update", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "mesh_update", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_update", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Update(
				"bearer",
				id,
				[]netip.AddrPort{
					netip.MustParseAddrPort("123.123.123.123:1234"),
				},
			)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Configure(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	targetID := uuid.New()
	url := fmt.Sprintf(urlMeshMachinesPeers, id.String(), targetID)
	tests := []testCase{
		testNewCase(t, http.StatusOK, url, "peer_update", nil),
		testNewCase(t, http.StatusBadRequest, url, "peer_update", ErrBadRequest),
		testNewCase(t, http.StatusUnauthorized, url, "peer_update", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "peer_update", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "peer_update", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Configure(
				"bearer",
				id,
				targetID,
				false,
				false,
				false,
				false,
				false,
			)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Unregister(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	url := fmt.Sprintf(urlMeshMachines, id.String())
	tests := []testCase{
		testNewCase(t, http.StatusAccepted, url, "", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_unregister", ErrBadRequest),
		testNewCase(t, http.StatusUnauthorized, url, "mesh_unregister", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "mesh_unregister", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_unregister", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Unregister("bearer", id)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_List(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	url := fmt.Sprintf(urlMeshPeers, id.String())
	tests := []testCase{
		testNewCase(t, http.StatusOK, url, "mesh_list", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_list", ErrBadRequest),
		testNewCase(t, http.StatusUnauthorized, url, "mesh_list", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "mesh_list", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_list", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.List("bearer", id)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Unpair(t *testing.T) {
	category.Set(t, category.Unit)

	myID := uuid.New()
	otherID := uuid.New()
	url := fmt.Sprintf(urlMeshMachinesPeers, myID.String(), otherID.String())

	tests := []testCase{
		testNewCase(t, http.StatusNoContent, url, "", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_unpair", ErrBadRequest),
		testNewCase(t, http.StatusUnauthorized, url, "mesh_unpair", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "mesh_unpair", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_unpair", ErrNotFound),
		testNewCase(t, http.StatusConflict, url, "mesh_unpair", ErrConflict),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Unpair("bearer", myID, otherID)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Invite(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	url := fmt.Sprintf(urlInvitationSend, id.String())

	tests := []testCase{
		testNewCase(t, http.StatusCreated, url, "mesh_send_invite", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_send_invite", ErrMaximumDeviceCount),
		testNewCase(t, http.StatusUnauthorized, url, "mesh_send_invite", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "mesh_send_invite", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_send_invite", ErrNotFound),
		testNewCase(t, http.StatusConflict, url, "mesh_send_invite", ErrConflict),
		testNewCase(t, http.StatusTooManyRequests, url, "mesh_send_invite", ErrTooManyRequests),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Invite("bearer", id, "elite@hacker.nord", false, false, false, false)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Received(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	url := fmt.Sprintf(urlReceivedInvitationsList, id.String())

	tests := []testCase{
		testNewCase(t, http.StatusOK, url, "mesh_received_invitations", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_received_invitations", ErrBadRequest),
		testNewCase(t, http.StatusForbidden, url, "mesh_received_invitations", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_received_invitations", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.Received("bearer", id)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Sent(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	url := fmt.Sprintf(urlSentInvitationsList, id.String())

	tests := []testCase{
		testNewCase(t, http.StatusOK, url, "mesh_sent_invitations", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_sent_invitations", ErrBadRequest),
		testNewCase(t, http.StatusForbidden, url, "mesh_sent_invitations", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_sent_invitations", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			_, err := api.Sent("bearer", id)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Accept(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	invitationID := uuid.New()
	url := fmt.Sprintf(urlAcceptInvitation, id.String(), invitationID.String())

	tests := []testCase{
		testNewCase(t, http.StatusOK, url, "", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_accept_invitation", ErrMaximumDeviceCount),
		testNewCase(t, http.StatusUnauthorized, url, "mesh_accept_invitation", ErrUnauthorized),
		testNewCase(t, http.StatusForbidden, url, "mesh_accept_invitation", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_accept_invitation", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Accept(
				"bearer",
				id,
				invitationID,
				false,
				false,
				false,
				false,
			)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Reject(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	invitationID := uuid.New()
	url := fmt.Sprintf(urlRejectInvitation, id.String(), invitationID.String())

	tests := []testCase{
		testNewCase(t, http.StatusNoContent, url, "", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_reject_invitation", ErrBadRequest),
		testNewCase(t, http.StatusForbidden, url, "mesh_reject_invitation", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_reject_invitation", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Reject("bearer", id, invitationID)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshAPI_Revoke(t *testing.T) {
	category.Set(t, category.Unit)

	id := uuid.New()
	invitationID := uuid.New()
	url := fmt.Sprintf(urlRevokeInvitation, id.String(), invitationID.String())

	tests := []testCase{
		testNewCase(t, http.StatusNoContent, url, "", nil),
		testNewCase(t, http.StatusBadRequest, url, "mesh_revoke_invitation", ErrBadRequest),
		testNewCase(t, http.StatusForbidden, url, "mesh_revoke_invitation", ErrForbidden),
		testNewCase(t, http.StatusNotFound, url, "mesh_revoke_invitation", ErrNotFound),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()

			api := NewDefaultAPI(
				"",
				server.URL,
				request.NewHTTPClient(&http.Client{}, nil, nil),
				response.MockValidator{},
			)
			err := api.Revoke("bearer", id, invitationID)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestMeshApiUtils_responseToMachinePeers(t *testing.T) {
	rawPeers := []mesh.MachinePeerResponse{
		{
			ID:                        uuid.MustParse("cb5a8446-e404-11ed-b5ea-0242ac120002"),
			PublicKey:                 "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t=",
			Hostname:                  "test0-everest.nord",
			Addresses:                 []netip.Addr{netip.MustParseAddr("192.17.30.5")},
			OS:                        "linux",
			Distro:                    "ubuntu",
			Endpoints:                 []netip.AddrPort{},
			Email:                     "test@mailto.com",
			IsLocal:                   true,
			DoesPeerAllowRouting:      true,
			DoesPeerAllowInbound:      true,
			DoesPeerAllowLocalNetwork: true,
			DoesPeerAllowFileshare:    true,
			DoesPeerSupportRouting:    true,
			DoIAllowInbound:           true,
			DoIAllowRouting:           true,
			DoIAllowLocalNetwork:      true,
			DoIAllowFileshare:         true,
		},
		{
			ID:                        uuid.MustParse("a7e4e7d6-e404-11ed-b5ea-0242ac120002"),
			PublicKey:                 "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp=",
			Hostname:                  "test0-everest.nord",
			Addresses:                 []netip.Addr{},
			OS:                        "linux",
			Distro:                    "ubuntu",
			Endpoints:                 []netip.AddrPort{},
			Email:                     "test@mailto.com",
			IsLocal:                   false,
			DoesPeerAllowRouting:      true,
			DoesPeerAllowInbound:      true,
			DoesPeerAllowLocalNetwork: true,
			DoesPeerAllowFileshare:    true,
			DoesPeerSupportRouting:    true,
			DoIAllowInbound:           true,
			DoIAllowRouting:           true,
			DoIAllowLocalNetwork:      true,
			DoIAllowFileshare:         true,
		},
	}

	expectedPeers := []mesh.MachinePeer{
		{
			ID:                        uuid.MustParse("cb5a8446-e404-11ed-b5ea-0242ac120002"),
			PublicKey:                 "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t=",
			Hostname:                  "test0-everest.nord",
			Address:                   netip.MustParseAddr("192.17.30.5"),
			OS:                        mesh.OperatingSystem{Name: "linux", Distro: "ubuntu"},
			Endpoints:                 []netip.AddrPort{},
			Email:                     "test@mailto.com",
			IsLocal:                   true,
			DoesPeerAllowRouting:      true,
			DoesPeerAllowInbound:      true,
			DoesPeerAllowLocalNetwork: true,
			DoesPeerAllowFileshare:    true,
			DoesPeerSupportRouting:    true,
			DoIAllowInbound:           true,
			DoIAllowRouting:           true,
			DoIAllowLocalNetwork:      true,
			DoIAllowFileshare:         true,
		},
		{
			ID:                        uuid.MustParse("a7e4e7d6-e404-11ed-b5ea-0242ac120002"),
			PublicKey:                 "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp=",
			Hostname:                  "test0-everest.nord",
			Address:                   netip.Addr{},
			OS:                        mesh.OperatingSystem{Name: "linux", Distro: "ubuntu"},
			Endpoints:                 []netip.AddrPort{},
			Email:                     "test@mailto.com",
			IsLocal:                   false,
			DoesPeerAllowRouting:      true,
			DoesPeerAllowInbound:      true,
			DoesPeerAllowLocalNetwork: true,
			DoesPeerAllowFileshare:    true,
			DoesPeerSupportRouting:    true,
			DoIAllowInbound:           true,
			DoIAllowRouting:           true,
			DoIAllowLocalNetwork:      true,
			DoIAllowFileshare:         true,
		},
	}

	peers := peersResponseToMachinePeers(rawPeers)

	assert.Equal(t, expectedPeers, peers)
}

func TestMeshApiUtils_responseToLocalPeers(t *testing.T) {
	rawPeers := []mesh.MachinePeerResponse{
		{
			ID:        uuid.MustParse("cb5a8446-e404-11ed-b5ea-0242ac120002"),
			PublicKey: "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t=",
			Hostname:  "test0-everest.nord",
			Addresses: []netip.Addr{netip.MustParseAddr("192.17.30.5")},
		},
		{
			ID:        uuid.MustParse("a7e4e7d6-e404-11ed-b5ea-0242ac120002"),
			PublicKey: "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp=",
			Hostname:  "test0-everest.nord",
			Addresses: []netip.Addr{},
		},
	}

	expectedPeers := []mesh.Machine{
		{
			ID:        uuid.MustParse("cb5a8446-e404-11ed-b5ea-0242ac120002"),
			PublicKey: "uXGPBcjbGrM62g5ew9gyPZaJsFNJI1peuFFhv1WYc4t=",
			Hostname:  "test0-everest.nord",
			Address:   netip.MustParseAddr("192.17.30.5"),
		},
		{
			ID:        uuid.MustParse("a7e4e7d6-e404-11ed-b5ea-0242ac120002"),
			PublicKey: "bu5BB8ks1pGgvDpENonCr7w51od5gWUM7RwO4SsvHmp=",
			Hostname:  "test0-everest.nord",
			Address:   netip.Addr{},
		},
	}

	peers := peersResponseToLocalPeers(rawPeers)

	assert.Equal(t, peers, expectedPeers)
}
