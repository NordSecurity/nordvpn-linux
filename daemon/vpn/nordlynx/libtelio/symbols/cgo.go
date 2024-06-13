//go:build !internal

package libtelio

// #cgo amd64 LDFLAGS: -L${SRCDIR}/../../../../../../libtelio/target/release -ltelio
// #cgo 386 LDFLAGS: -L${SRCDIR}/../../../../../../libtelio/target/release -ltelio
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../../../libtelio/target/release -ltelio
// #cgo arm LDFLAGS: -L${SRCDIR}/../../../../../../libtelio/target/release -ltelio
// #cgo arm64 LDFLAGS: -L${SRCDIR}/../../../../../../libtelio/target/release -ltelio
// #cgo LDFLAGS: -ldl -lm
import "C"
