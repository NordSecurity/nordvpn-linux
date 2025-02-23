package registry

import (
	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"
	"github.com/google/uuid"
)

type NotifyingRegistry struct {
	inner mesh.Registry
	sub   events.PublishSubcriber[[]string]
}

func NewNotifyingRegistry(inner mesh.Registry, sub events.PublishSubcriber[[]string]) *NotifyingRegistry {
	return &NotifyingRegistry{
		inner: inner,
		sub:   sub,
	}
}

func (r *NotifyingRegistry) Register(token string, self mesh.Machine) (*mesh.Machine, error) {
	return r.inner.Register(token, self)
}
func (r *NotifyingRegistry) Update(token string, id uuid.UUID, info mesh.MachineUpdateRequest) error {
	if err := r.inner.Update(token, id, info); err != nil {
		return err
	}
	r.sub.Publish([]string{id.String()})
	return nil
}
func (r *NotifyingRegistry) Configure(
	token string,
	id uuid.UUID,
	peerID uuid.UUID,
	peerUpdateInfo mesh.PeerUpdateRequest,
) error {
	if err := r.inner.Configure(token, id, peerID, peerUpdateInfo); err != nil {
		return err
	}
	r.sub.Publish([]string{id.String()})
	return nil
}
func (r *NotifyingRegistry) Unregister(token string, id uuid.UUID) error {
	if err := r.inner.Unregister(token, id); err != nil {
		return err
	}
	r.sub.Publish([]string{id.String()})
	return nil
}
func (r *NotifyingRegistry) Unpair(token string, self uuid.UUID, peer uuid.UUID) error {
	if err := r.inner.Unpair(token, self, peer); err != nil {
		return err
	}
	r.sub.Publish([]string{peer.String()})
	return nil
}
func (r *NotifyingRegistry) NotifyNewTransfer(
	token string,
	self uuid.UUID,
	peer uuid.UUID,
	fileName string,
	fileCount int,
	transferID string,
) error {
	return r.inner.NotifyNewTransfer(token, self, peer, fileName, fileCount, transferID)
}
