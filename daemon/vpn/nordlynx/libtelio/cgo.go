package libtelio

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libtelio/current/amd64
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libtelio/current/i386
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libtelio/current/armel
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libtelio/current/armhf
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libtelio/current/aarch64
// #cgo LDFLAGS: -ldl -lm -ltelio
import "C"
