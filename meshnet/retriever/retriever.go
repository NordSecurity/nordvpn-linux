// Package registry implements extra error handling over MeshAPI request
package retriever

import (
	"strings"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"

	"github.com/google/uuid"
)

// Retriever is a wrapper around the inner Retriever which checks whether this device was removed
// and notifies the subscribers about it.
type Retriever struct {
	inner     mesh.Retriever
	publisher events.Publisher[any]
}

// NewRetriever create new Retriever instance
func NewRetriever(
	ret mesh.Retriever,
	publisher events.Publisher[any],
) *Retriever {
	return &Retriever{inner: ret, publisher: publisher}
}

func (r *Retriever) SetNotificationSubject(pub events.Publisher[any]) { r.publisher = pub }

func (r *Retriever) notifySelfRemoved() { r.publisher.Publish(nil) }

// List given peer's neighbours in the mesh network.
func (r *Retriever) List(token string, self uuid.UUID) (resp mesh.MachinePeers, err error) {
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

func (r *Retriever) Map(token string, self uuid.UUID) (resp *mesh.MachineMap, err error) {
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
