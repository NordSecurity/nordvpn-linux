package events

import (
	"fmt"
	"testing"
	"time"

	"github.com/NordSecurity/nordvpn-linux/test/category"
	"gotest.tools/v3/assert"
)

func TestPublishDisconnect(t *testing.T) {
	category.Set(t, category.Unit)

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "disconnect success",
			err:  nil,
		},
		{
			name: "disconnect failure",
			err:  fmt.Errorf("disconnect failure"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var e DataDisconnect
			var eventPublished bool
			publishFunc := func(event DataDisconnect) {
				eventPublished = true
				e = event
			}

			sender := NewDisconnectSender(DataDisconnect{}, publishFunc)
			sender.PublishDisconnect(time.Now(), test.err)

			assert.Equal(t, eventPublished, true, "Event not published after PublishDisconnect was called.")
			if test.err != nil {
				assert.Equal(t, e.EventStatus, StatusFailure,
					"Event published with invalid status published, expected StatusFailure.")
			} else {
				assert.Equal(t, e.EventStatus, StatusSuccess,
					"Event published with invalid status published, expected StatusSucccess.")
			}

			assert.Equal(t, e.IsRefresh, true, "IsRefresh should be always true.")
		})
	}
}
