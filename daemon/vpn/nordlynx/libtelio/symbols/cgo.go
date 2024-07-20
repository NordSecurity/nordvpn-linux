package libtelio

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../../bin/deps/lib/amd64/latest -ltelio -lsqlite3
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../../bin/deps/lib/i386/latest -ltelio -lsqlite3
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../../bin/deps/lib/armel/latest -ltelio -lsqlite3
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../../bin/deps/lib/armhf/latest -ltelio -lsqlite3
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../../bin/deps/lib/arm64/latest -ltelio -lsqlite3
// #cgo LDFLAGS: -ldl -lm
import "C"
