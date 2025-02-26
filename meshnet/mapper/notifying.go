package mapper

import (
	"bytes"
	"strings"
	"sync"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/events"

	"github.com/google/uuid"
)

// NotifyingMapper is a wrapper around the inner Mapper which checks whether this device was
// removed and notifies the subscribers about it.
type NotifyingMapper struct {
	inner   mesh.CachingMapper
	selfPub events.Publisher[any]
	peerPub events.Publisher[[]string]
	lastMap []byte
	mu      sync.RWMutex
}

// NewNotifyingMapper creates a new Mapper instance.
func NewNotifyingMapper(
	inner mesh.CachingMapper,
	selfPub events.Publisher[any],
	peerPub events.Publisher[[]string],
) *NotifyingMapper {
	return &NotifyingMapper{inner: inner, selfPub: selfPub, peerPub: peerPub}
}

func (r *NotifyingMapper) notifySelfRemoved() { r.selfPub.Publish(nil) }

func (r *NotifyingMapper) Map(
	token string,
	self uuid.UUID,
	forceUpdate bool,
) (resp *mesh.MachineMap, err error) {
	anotherMachine := self
	resp, err = r.inner.Map(token, anotherMachine, forceUpdate)
	if err != nil {
		if strings.Contains(err.Error(), "Machine not found") {
			r.notifySelfRemoved()
		}
		return nil, err
	}
	r.mu.RLock()
	lastMap := r.lastMap
	peerPub := r.peerPub
	r.mu.RUnlock()
	// Peer update should only be published when update is not forced due to the following
	// reasons:
	// 1. Subscriber may be issuing forced map fetch on peer update notification, therefore this
	// would result in an infinite loop;
	// 2. Forced update is already called after operations that modify meshnet map, thus,
	// already notifying about mesh map update.
	if !forceUpdate && !bytes.Equal(lastMap, resp.Raw) {
		peerPub.Publish(nil)
	}
	r.mu.Lock()
	r.lastMap = resp.Raw
	r.mu.Unlock()
	return resp, nil
}
