package internal

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/NordSecurity/nordvpn-linux/log"
	linux "golang.org/x/sys/unix"
)

func GetSignalChan() <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, linux.SIGTERM, linux.SIGHUP, linux.SIGUSR1)
	return signals
}

// RunAsync executes the function on a different goroutine and waits until goroutine starts
func RunAsync(fn func()) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		wg.Done()
		fn()
	}()

	wg.Wait()
}

// TrySendTimeout - tries to send value on channel.
// Blocks for the specified duration. If channel is blocked for the entire duration, value is discarded.
func TrySendTimeout[Type any](ch chan<- Type, v Type, d time.Duration) bool {
	select {
	case ch <- v:
	case <-time.After(d):
		log.Warn("no listener dropping after delay", v)
		return false
	}

	return true
}

// TrySend - tries to send the value.
// If the channel is busy then the value is discarded immediately.
func TrySend[Type any](ch chan<- Type, v Type) bool {
	select {
	case ch <- v:
	default:
		log.Warn("no listener dropping", v)
		return false
	}

	return true
}
