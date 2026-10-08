package desktop

import (
	"time"
)

const (
	StatusConnected    = "Connected"
	StatusConnecting   = "Connecting"
	StatusDisconnected = "Disconnected"
)

func Eventually(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(500 * time.Millisecond)
	}
}
