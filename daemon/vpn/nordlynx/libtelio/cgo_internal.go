//go:build internal

package libtelio

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libmoose-nordvpnapp/current/amd64
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libmoose-nordvpnapp/current/i386
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libmoose-nordvpnapp/current/armel
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libmoose-nordvpnapp/current/armhf
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../bin/deps/lib/libmoose-nordvpnapp/current/aarch64
// #cgo LDFLAGS: -lsqlite3
import "C"
