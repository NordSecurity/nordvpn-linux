//go:build !internal

package symbols

// TODO: Fix paths

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../libdrop/target/release -lnorddrop
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../libdrop/target/release -lnorddrop
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../libdrop/target/release -lnorddrop
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../libdrop/target/release -lnorddrop
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../libdrop/target/release -lnorddrop
// #cgo LDFLAGS: -L/usr/local/lib -L${SRCDIR}/../../../libdrop/target/release -lnorddrop
import "C"
