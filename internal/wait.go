package internal

import (
	"os"
	"os/signal"

	linux "golang.org/x/sys/unix"
)

func GetSignalChan() <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, linux.SIGTERM, linux.SIGHUP)
	return signals
}

// WaitSignal for app to shutdown
func WaitSignal() {
	signals := GetSignalChan()
	<-signals
}
