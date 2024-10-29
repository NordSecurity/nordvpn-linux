//go:build cgo

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/NordSecurity/nordvpn-linux/internal"
)

/*
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

void redirect_stdout_to_fd(int fd) {
    dup2(fd, fileno(stdout));
    dup2(fd, fileno(stderr));
}
*/
import "C"

// logSetup sets the stdout and stderr outside go runtime to a given file
func logSetup(f *os.File) {
	C.redirect_stdout_to_fd(C.int(f.Fd()))
	// Ignore default printing of go stack trace in case of SIGABRT in prod builds which may be
	// produced by Rust panics
	if internal.IsProdEnv(Environment) {
		signal.Notify(make(chan os.Signal), syscall.SIGABRT)
	}
}
