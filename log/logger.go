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

type logLevel uint32

// showCallerAsSource instructs logger to skip frames and show `log.XYZ` function
// caller file and line of code instead of file and line of `log.XYZ` function.
const showCallerAsSource = 3

const (
	levelUnknown logLevel = iota
	levelDebug
	levelInfo
	levelWarn
	levelError
	levelFatal
	levelOff
)

const (
	debugPrefix   = "[Debug]"
	infoPrefix    = "[Info]"
	warningPrefix = "[Warning]"
	deferPrefix   = "[Defer]"
	errorPrefix   = "[Error]"
	fatalPrefix   = "[Fatal]"
)

func (l logLevel) String() string {
	switch l {
	case levelDebug:
		return "debug"
	case levelInfo:
		return "info"
	case levelWarn:
		return "warn"
	case levelError:
		return "error"
	case levelFatal:
		return "fatal"
	case levelOff:
		return "off"
	case levelUnknown:
		fallthrough
	default:
		return "unknown"
	}
}

// CancelFunc stops a logger resource such as the level file watcher.
type CancelFunc func()

var level atomic.Uint32

// DefaultLevel returns LevelDebug for dev environments and LevelInfo otherwise.
func DefaultLevel(devEnv bool) logLevel {
	return levelDebug
}

// SetupLogger configures the log output and initial level, then starts watching
// levelFilePath for runtime level changes. The returned CancelFunc must be
// called on shutdown to stop the watcher.
func SetupLogger(out io.Writer, levelFilePath string, initialLevel logLevel) CancelFunc {
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

func SetLevel(l logLevel) {
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
	if atLevel(levelDebug) {
		output(fmt.Sprint(append([]any{debugPrefix}, v...)...))
	}
}

func atLevel(l logLevel) bool {
	return level.Load() <= uint32(l)
}

func output(msg string) {
	if err := log.Output(showCallerAsSource, msg); err != nil {
		fmt.Fprintf(os.Stderr, "log.Output: %v\n", err)
	}
}

func logMsg(prefix string, v []any) string {
	msg := fmt.Sprintln(v...)
	return prefix + " " + msg[:len(msg)-1] // trim trailing \n
}

func Debugf(format string, v ...any) {
	if atLevel(levelDebug) {
		output(fmt.Sprintf(debugPrefix+" "+format, v...))
	}
}

func Info(v ...any) {
	if atLevel(levelInfo) {
		output(logMsg(infoPrefix, v))
	}
}

func Infof(format string, v ...any) {
	if atLevel(levelInfo) {
		output(fmt.Sprintf(infoPrefix+" "+format, v...))
	}
}

func Warn(v ...any) {
	if atLevel(levelWarn) {
		output(logMsg(warningPrefix, v))
	}
}

func Warnf(format string, v ...any) {
	if atLevel(levelWarn) {
		output(fmt.Sprintf(warningPrefix+" "+format, v...))
	}
}

func Defer(v ...any) {
	if atLevel(levelInfo) {
		output(logMsg(deferPrefix, v))
	}
}

func Deferf(format string, v ...any) {
	if atLevel(levelInfo) {
		output(fmt.Sprintf(deferPrefix+" "+format, v...))
	}
}

func Error(v ...any) {
	if atLevel(levelError) {
		output(logMsg(errorPrefix, v))
	}
}

func Errorf(format string, v ...any) {
	if atLevel(levelError) {
		output(fmt.Sprintf(errorPrefix+" "+format, v...))
	}
}

func Fatal(v ...any) {
	output(logMsg(fatalPrefix, v))
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	output(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Print(v ...any) {
	log.Print(v...)
}

func atLevel(l LogLevel) bool {
	return level.Load() <= uint32(l)
}

func output(depth int, msg string) {
	if err := log.Output(depth+1, msg); err != nil {
		fmt.Fprintf(os.Stderr, "log.Output: %v\n", err)
	}
}

func Fatalln(v ...any) {
	log.Fatalln(v...)
}
