//go:build vinis

package vinis

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/amd64/latest -lvinis
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/i386/latest -lvinis
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/armel/latest -lvinis
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/armhf/latest -lvinis
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/aarch64/latest -lvinis
// #cgo LDFLAGS: -ldl -lm
import "C"
