package state

import (
	"testing"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"github.com/stretchr/testify/assert"
)

func TestStatePublisher_DoneSubscribersRemoval(t *testing.T) {
	category.Set(t, category.Unit)

	state := NewState()
	state.AddSubscriber()
	eventsChan, doneChan := state.AddSubscriber()
	state.AddSubscriber()

	close(doneChan)

	state.AddSubscriber()

	_, isOpen := <-eventsChan
	assert.False(t, isOpen, "Event channel remains open for unsubscribed subscriber.")
	assert.Len(t, state.subscribers, 3, "Unexpected number of subscribers.")
}
