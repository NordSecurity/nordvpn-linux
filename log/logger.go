// Package log wraps the standard library logger with level-filtered functions
// (Debug, Info, Warn, Error, Defer). The active level is stored atomically and
// can be changed at runtime by writing to the file watched by SetupLogger.
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
)

type LogLevel uint32

const (
	LevelUnknown LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelOff
)

const (
	DebugPrefix = "[Debug]"
	InfoPrefix  = "[Info]"
	// WarningPrefix is used when logging errors, which don't impact control flow.
	WarningPrefix = "[Warning]"
	// DeferPrefix is used when logging errors in deferred or cleanup code.
	DeferPrefix = "[Defer]"
	// ErrorPrefix is used when logging errors, which impact control flow.
	ErrorPrefix = "[Error]"
	FatalPrefix = "[Fatal]"
)

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelOff:
		return "off"
	case LevelUnknown:
		fallthrough
	default:
		return "unknown"
	}
}

// CancelFunc stops a logger resource such as the level file watcher.
type CancelFunc func()

var level atomic.Uint32

// DefaultLevel returns LevelDebug for dev environments and LevelInfo otherwise.
func DefaultLevel(devEnv bool) LogLevel {
	if devEnv {
		return LevelDebug
	}
	return LevelInfo
}

// SetupLogger configures the log output and initial level, then starts watching
// levelFilePath for runtime level changes. The returned CancelFunc must be
// called on shutdown to stop the watcher.
func SetupLogger(out io.Writer, levelFilePath string, initialLevel LogLevel) CancelFunc {
	SetOutput(out)
	SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	SetLevel(initialLevel)
	cancel, err := WatchLevelFile(levelFilePath)
	if err != nil {
		Error("failed to start log level watcher:", err)
		return func() {}
	}
	return cancel
}

func SetLevel(l LogLevel) {
	Info("setting log level to", l)
	level.Store(uint32(l))
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func SetFlags(flag int) {
	log.SetFlags(flag)
}

func Debug(v ...any) {
	if atLevel(LevelDebug) {
		output(2, fmt.Sprint(append([]any{DebugPrefix}, v...)...))
	}
}

func atLevel(l LogLevel) bool {
	return level.Load() <= uint32(l)
}

func output(depth int, msg string) {
	if err := log.Output(depth+1, msg); err != nil {
		fmt.Fprintf(os.Stderr, "log.Output: %v\n", err)
	}
}

func logMsg(prefix string, v []any) string {
	msg := fmt.Sprintln(v...)
	return prefix + " " + msg[:len(msg)-1] // trim trailing \n
}

func Debugf(format string, v ...any) {
	if atLevel(LevelDebug) {
		output(2, fmt.Sprintf(DebugPrefix+" "+format, v...))
	}
}

func Info(v ...any) {
	if atLevel(LevelInfo) {
		output(2, logMsg(InfoPrefix, v))
	}
}

func Infof(format string, v ...any) {
	if atLevel(LevelInfo) {
		output(2, fmt.Sprintf(InfoPrefix+" "+format, v...))
	}
}

func Warn(v ...any) {
	if atLevel(LevelWarn) {
		output(2, logMsg(WarningPrefix, v))
	}
}

func Warnf(format string, v ...any) {
	if atLevel(LevelWarn) {
		output(2, fmt.Sprintf(WarningPrefix+" "+format, v...))
	}
}

func Defer(v ...any) {
	if atLevel(LevelInfo) {
		output(2, logMsg(DeferPrefix, v))
	}
}

func Deferf(format string, v ...any) {
	if atLevel(LevelInfo) {
		output(2, fmt.Sprintf(DeferPrefix+" "+format, v...))
	}
}

func Error(v ...any) {
	if atLevel(LevelError) {
		output(2, logMsg(ErrorPrefix, v))
	}
}

func Errorf(format string, v ...any) {
	if atLevel(LevelError) {
		output(2, fmt.Sprintf(ErrorPrefix+" "+format, v...))
	}
}

func Fatal(v ...any) {
	output(2, logMsg(FatalPrefix, v))
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}
