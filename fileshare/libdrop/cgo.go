package libdrop

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/amd64/latest -lnorddrop -lsqlite3
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/i386/latest -lnorddrop -lsqlite3
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/armel/latest -lnorddrop -lsqlite3
// #cgo arm LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/armhf/latest -lnorddrop -lsqlite3
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../bin/deps/lib/aarch64/latest -lnorddrop -lsqlite3
// #cgo LDFLAGS: -ldl -lm
import "C"
