package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
	"golang.org/x/sys/unix"
	"gopkg.in/natefinch/lumberjack.v2"
)

// redirectStdOutputToLumberjack - redirects the stdout and stderr to lumberjack
func redirectStdOutputToLumberjack(lj *lumberjack.Logger) (cleanup func() error, retErr error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	defer func() {
		if retErr != nil {
			_ = w.Close()
			_ = r.Close()
		}
	}()

	// Point fd 1 and fd 2 at the pipe's write end.
	// This is what makes C printf/fprintf(stderr,...) land in the pipe.
	if err := unix.Dup2(int(w.Fd()), int(os.Stdout.Fd())); err != nil {
		return nil, err
	}
	if err := unix.Dup2(int(w.Fd()), int(os.Stderr.Fd())); err != nil {
		return nil, err
	}
	// Keep Go's *os.File handles consistent with the underlying fds.
	os.Stdout = w
	os.Stderr = w

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(lj, r) // streams chunks straight into lumberjack
	}()

	cleanup = func() error {
		_ = w.Close() // EOF for the goroutine
		<-done        // wait for drain
		return lj.Close()
	}
	if internal.IsProdEnv(Environment) {
		signal.Notify(make(chan os.Signal, 1), syscall.SIGABRT)
	}
	return cleanup, nil
}
