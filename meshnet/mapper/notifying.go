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

// NotifyingMapper is a wrapper around the inner Mapper which checks whether this device was
// removed or there were any other updates in the mesh map and notifies the subscribers about it.
type NotifyingMapper struct {
	inner    mesh.CachingMapper
	selfPub  events.Publisher[any]
	peersPub events.Publisher[[]string]
	lastResp []byte
	mu       sync.Mutex
}

// NewNotifyingMapper create new Mapper instance
func NewNotifyingMapper(
	ret mesh.CachingMapper,
	selfPub events.Publisher[any],
	peersPub events.Publisher[[]string],
) *NotifyingMapper {
	return &NotifyingMapper{inner: ret, selfPub: selfPub, peersPub: peersPub}
}

func (r *NotifyingMapper) SetNotificationSubject(pub events.Publisher[any]) { r.selfPub = pub }

func (r *NotifyingMapper) notifySelfRemoved() { r.selfPub.Publish(nil) }

func (r *NotifyingMapper) Map(token string, self uuid.UUID, forceUpdate bool) (resp *mesh.MachineMap, err error) {
	resp, err = r.inner.Map(token, self, forceUpdate)
	if err != nil {
		if strings.Contains(err.Error(), "Machine not found") {
			r.notifySelfRemoved()
		}
		return nil, err
	}
	// forceUpdate indicates that caller will likely initiate mesh map update on its own
	if !forceUpdate && !bytes.Equal(r.lastResp, resp.Raw) {
		r.peersPub.Publish(nil)
	}
	r.mu.Lock()
	r.lastResp = resp.Raw
	r.mu.Unlock()
	return resp, nil
}
