//go:build !internal

package libtelio

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/amd64/latest -ltelio
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/i386/latest -ltelio
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/armel/latest -ltelio
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/armhf/latest -ltelio
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/aarch64/latest -ltelio
// #cgo LDFLAGS: -ldl -lm
import "C"
