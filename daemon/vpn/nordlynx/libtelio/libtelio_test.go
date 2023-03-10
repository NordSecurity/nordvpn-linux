package libtelio

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"

	"github.com/stretchr/testify/assert"
)

func TestIsConnected(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name      string
		state     state
		publicKey string
		expected  bool
	}{
		{
			name: "connecting",
			state: state{
				State:     "connecting",
				PublicKey: "123",
			},
			publicKey: "123",
		},
		{
			name: "connected",
			state: state{
				State:     "connected",
				PublicKey: "123",
			},
			publicKey: "123",
			expected:  true,
		},
		{
			name: "misbehaving",
			state: state{
				State:     "misbehaving",
				PublicKey: "123",
			},
			publicKey: "123",
		},
		{
			name: "different pubkey",
			state: state{
				State:     "connected",
				PublicKey: "321",
			},
			publicKey: "123",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch := make(chan state)
			go func() { ch <- test.state }()

			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
			defer cancel()
			isConnectedC := isConnected(ctx, ch, test.publicKey)

			assert.Equal(t, test.expected, <-isConnectedC)
		})
	}
}

func TestEventCallback_DoesntBlock(t *testing.T) {
	stateC := make(chan state)
	cb := eventCallback(stateC)
	event, err := json.Marshal(state{})
	assert.NoError(t, err)

	returnedC := make(chan any)
	go func() {
		cb(string(event))
		returnedC <- nil
	}()

	condition := func() bool {
		select {
		case <-returnedC:
			return true
		default:
			return false
		}
	}
	assert.Eventually(t, condition, time.Millisecond*100, time.Millisecond*10)
}
