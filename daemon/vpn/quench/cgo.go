//go:build quench

package quench

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../bin/deps/lib/libquench/current/amd64
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../bin/deps/lib/libquench/current/i386
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../bin/deps/lib/libquench/current/armel
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../bin/deps/lib/libquench/current/armhf
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../bin/deps/lib/libquench/current/aarch64
// #cgo LDFLAGS: -lquench -ldl -lm
import "C"
