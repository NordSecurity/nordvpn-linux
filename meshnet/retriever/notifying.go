// Package registry implements extra error handling over MeshAPI request
package retriever

import (
	"bytes"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"

	"github.com/google/uuid"
)

// NotifyingRetriever is a wrapper around the inner Retriever which checks whether this device was removed
// and notifies the subscribers about it.
type NotifyingRetriever struct {
	inner   mesh.CachingRetriever
	selfPub events.Publisher[any]
	peerPub events.Publisher[[]string]
	lastMap []byte
	mu      sync.Mutex
}

// NewNotifyingRetriever create new Retriever instance
func NewNotifyingRetriever(
	ret mesh.CachingRetriever,
	selfPub events.Publisher[any],
	peerPub events.Publisher[[]string],
) *NotifyingRetriever {
	return &NotifyingRetriever{inner: ret, selfPub: selfPub, peerPub: peerPub}
}

func (r *NotifyingRetriever) notifySelfRemoved() { r.selfPub.Publish(nil) }

func (r *NotifyingRetriever) Map(token string, self uuid.UUID, forceUpdate bool) (resp *mesh.MachineMap, err error) {
	anotherMachine := self
	resp, err = r.inner.Map(token, anotherMachine, forceUpdate)
	if err != nil {
		if strings.Contains(err.Error(), "Machine not found") {
			r.notifySelfRemoved()
		}
		return nil, err
	}
	// This is only important for non forced updates as forced ones usually refresh the meshnet
	if !forceUpdate && !bytes.Equal(r.lastMap, resp.Raw) {
		r.peerPub.Publish(nil)
	}
	r.mu.Lock()
	r.lastMap = resp.Raw
	r.mu.Unlock()
	return resp, nil
}
