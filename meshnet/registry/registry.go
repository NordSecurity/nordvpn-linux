// Package registry implements extra error handling over MeshAPI request
package registry

import (
	"net/netip"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"

	"github.com/google/uuid"
)

// Registry structure holds neccessary data to execute API calls
type Registry struct {
	inner     mesh.Registry
	publisher events.Publisher[any]
}

// NewRegistry create new Registry instance
func NewRegistry(
	reg mesh.Registry,
	publisher events.Publisher[any],
) *Registry {
	return &Registry{inner: reg, publisher: publisher}
}

func (r *Registry) SetNotificationSubject(pub events.Publisher[any]) { r.publisher = pub }

func (r *Registry) notifySelfRemoved() { r.publisher.Publish(nil) }

// Register Self to mesh network.
func (r *Registry) Register(token string, self mesh.Machine) (*mesh.Machine, error) {
	return r.inner.Register(token, self)
}

// Update already registered peer.
func (r *Registry) Update(token string, id uuid.UUID, endpoints []netip.AddrPort) error {
	return r.inner.Update(token, id, endpoints)
}

// Configure interaction with specific peer.
func (r *Registry) Configure(
	token string,
	id uuid.UUID,
	peerID uuid.UUID,
	doIAllowInbound bool,
	doIAllowRouting bool,
	doIAllowLocalNetwork bool,
	doIAllowSend bool,
) error {
	return r.inner.Configure(token, id, peerID, doIAllowInbound, doIAllowRouting, doIAllowLocalNetwork, doIAllowSend)
}

// Unregister Peer from the mesh network.
func (r *Registry) Unregister(token string, self uuid.UUID) error {
	return r.inner.Unregister(token, self)
}

// Local peers owned the same user.
func (r *Registry) Local(token string) (mesh.Machines, error) {
	return r.inner.Local(token)
}

// List given peer's neighbours in the mesh network.
func (r *Registry) List(token string, self uuid.UUID) (resp mesh.MachinePeers, err error) {
	anotherMachine := self
	resp, err = r.inner.List(token, anotherMachine)
	if err != nil {
		if strings.Contains(err.Error(), "Machine not found") {
			r.notifySelfRemoved()
		}
		return nil, err
	}
	return resp, nil
}

func (r *Registry) Map(token string, self uuid.UUID) (resp *mesh.MachineMap, err error) {
	anotherMachine := self
	resp, err = r.inner.Map(token, anotherMachine)
	if err != nil {
		if strings.Contains(err.Error(), "Machine not found") {
			r.notifySelfRemoved()
		}
		return nil, err
	}
	return resp, nil
}

// Unpair invited peer.
func (r *Registry) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	return r.inner.Unpair(token, self, peer)
}

func (r *Registry) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
) error {
	return r.inner.NotifyNewTransfer(token, self, peer, fileName, fileCount)
}
