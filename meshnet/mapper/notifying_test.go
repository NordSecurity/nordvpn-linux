package mapper

import (
	"errors"
	"io"
	"testing"

	"github.com/NordSecurity/nordvpn-linux/core/mesh"
	"github.com/NordSecurity/nordvpn-linux/daemon/events"
	"github.com/NordSecurity/nordvpn-linux/test/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotifyingMapper_Map(t *testing.T) {
	errNotFound := errors.New("Machine not found")
	initialMap := &mesh.MachineMap{Raw: []byte{0x02}}
	updatedMap := &mesh.MachineMap{Raw: []byte{0x02, 0x04}}
	for _, tt := range []struct {
		name           string
		err            error
		mmap           *mesh.MachineMap
		forceUpdate    bool
		selfPublished  bool
		peersPublished bool
	}{
		{
			name: "success no notifications",
			mmap: initialMap,
		},
		{
			name: "error no notifications",
			err:  io.EOF,
		},
		{
			name:          "self removed when not found",
			err:           errNotFound,
			selfPublished: true,
		},
		{
			name:           "map updated",
			mmap:           updatedMap,
			peersPublished: true,
		},
		{
			name:        "no notification when force update",
			mmap:        updatedMap,
			forceUpdate: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			selfPub := &events.MockPublisherSubscriber[any]{}
			peerPub := &events.MockPublisherSubscriber[[]string]{}
			inner := &mock.CachingMapperMock{Error: tt.err, Value: tt.mmap}
			mapper := NewNotifyingMapper(inner, selfPub, peerPub)
			mapper.lastMap = initialMap.Raw
			mmap, err := mapper.Map("", uuid.UUID{}, tt.forceUpdate)
			assert.ErrorIs(t, tt.err, err)
			assert.EqualValues(t, tt.mmap, mmap)
			assert.Equal(t, tt.selfPublished, selfPub.EventPublished)
			assert.Equal(t, tt.peersPublished, peerPub.EventPublished)
		})
	}
}
