//go:build !internal

package drop

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/amd64/latest -lfoss
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/i386/latest -lfoss
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/armel/latest -lfoss
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/armhf/latest -lfoss
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/aarch64/latest -lfoss
// #cgo LDFLAGS: -ldl -lm
import "C"
