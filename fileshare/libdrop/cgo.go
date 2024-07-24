package libdrop

// NOTE: I'm not configuring the path to *.so files here. Instead, I'm setting
// `LIBRARY_PATH` to point to the directory, containing *.so files during build
// and `LD_LIBRARY_PATH` for runtime when running tests.

// #cgo LDFLAGS: -lm -lnorddrop
import "C"
