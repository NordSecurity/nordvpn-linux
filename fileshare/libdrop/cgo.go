package libdrop

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libdrop/current/amd64
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libdrop/current/i386
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libdrop/current/armel
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libdrop/current/armhf
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/libdrop/current/aarch64
// #cgo LDFLAGS: -ldl -lm -lnorddrop -lsqlite3
import "C"
