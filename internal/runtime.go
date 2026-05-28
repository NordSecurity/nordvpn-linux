package internal

import (
	"fmt"
	"runtime"
)

func GetCaller() string {
	// we need to skip two frames, one for getCaller, and one for caller of getCaller
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}

// Recommended to use only in debug builds
func GetStack() []string {
	output := []string{}

	// how many frames to bring
	pcs := make([]uintptr, 10)
	// we need to skip two frames, one for getCaller, and one for caller of getCaller
	n := runtime.Callers(2, pcs)

	// take only the valid pcs
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		output = append(output, fmt.Sprintf("%s:%d", frame.File, frame.Line))

		if !more {
			break
		}
	}
	return output
}
