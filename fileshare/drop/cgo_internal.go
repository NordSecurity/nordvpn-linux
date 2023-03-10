//go:build internal

package drop

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/amd64/latest -lnord
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/i386/latest -lnord
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/armel/latest -lnord
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/armhf/latest -lnord
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/nord/aarch64/latest -lnord
// #cgo LDFLAGS: -ldl -lm
import "C"
