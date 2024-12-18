//go:build quench

package quench

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/amd64/latest -lquench
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/i386/latest -lquench
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/armel/latest -lquench
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/armhf/latest -lquench
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/aarch64/latest -lquench
// #cgo LDFLAGS: -ldl -lm
import "C"
