//go:build !internal

package libdrop

// TODO: Fix paths
// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/amd64/latest -L${SRCDIR}/../../../libdrop/target/release -lfoss -lnorddrop
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/i386/latest -L${SRCDIR}/../../../libdrop/target/release -lfoss -lnorddrop
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/armel/latest -L${SRCDIR}/../../../libdrop/target/release -lfoss -lnorddrop
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/armhf/latest -L${SRCDIR}/../../../libdrop/target/release -lfoss -lnorddrop
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/foss/aarch64/latest -L${SRCDIR}/../../../libdrop/target/release -lfoss -lnorddrop
// #cgo LDFLAGS: -L/usr/local/lib -L${SRCDIR}/../../../libdrop/target/release -lfoss -lnorddrop
import "C"
