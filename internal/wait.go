package internal

import (
	"os"
	"os/signal"

	linux "golang.org/x/sys/unix"
)

// WaitSignal for app to shutdown
func WaitSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, linux.SIGTERM)
	<-signals
}
