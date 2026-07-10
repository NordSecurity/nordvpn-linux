// Package log wraps the standard library logger with level-filtered functions
// (Debug, Info, Warn, Error). The active level is stored atomically and can
// be changed at runtime by writing to the file watched by SetupLogger.
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
)

type logLevel uint32

// showCallerAsSource instructs logger to skip frames and show `log.XYZ` function
// caller file and line of code instead of file and line of `log.XYZ` function.
const showCallerAsSource = 4

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

func DefaultLevel() logLevel {
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

// Logger is a logger with a fixed prefix prepended to every message.
type Logger struct {
	prefix string
}

// NewLogger returns a Logger that prepends prefix to every log message.
func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}

func (l *Logger) Prefix() string { return l.prefix }

func (l *Logger) Debug(v ...any) { logAt(levelDebug, debugPrefix, prepend(l.prefix, v)) }
func (l *Logger) Debugf(format string, v ...any) {
	logAtf(levelDebug, debugPrefix, l.prefix+" "+format, v)
}
func (l *Logger) Info(v ...any) { logAt(levelInfo, infoPrefix, prepend(l.prefix, v)) }
func (l *Logger) Infof(format string, v ...any) {
	logAtf(levelInfo, infoPrefix, l.prefix+" "+format, v)
}
func (l *Logger) Warn(v ...any) { logAt(levelWarn, warningPrefix, prepend(l.prefix, v)) }
func (l *Logger) Warnf(format string, v ...any) {
	logAtf(levelWarn, warningPrefix, l.prefix+" "+format, v)
}
func (l *Logger) Error(v ...any) { logAt(levelError, errorPrefix, prepend(l.prefix, v)) }
func (l *Logger) Errorf(format string, v ...any) {
	logAtf(levelError, errorPrefix, l.prefix+" "+format, v)
}

func (l *Logger) Fatal(v ...any) {
	logAt(levelFatal, fatalPrefix, prepend(l.prefix, v))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	logAtf(levelFatal, fatalPrefix, l.prefix+" "+format, v)
	os.Exit(1)
}

func prepend(prefix string, v []any) []any {
	return append([]any{prefix}, v...)
}

func Debug(v ...any)                 { logAt(levelDebug, debugPrefix, v) }
func Debugf(format string, v ...any) { logAtf(levelDebug, debugPrefix, format, v) }
func Info(v ...any)                  { logAt(levelInfo, infoPrefix, v) }
func Infof(format string, v ...any)  { logAtf(levelInfo, infoPrefix, format, v) }
func Warn(v ...any)                  { logAt(levelWarn, warningPrefix, v) }
func Warnf(format string, v ...any)  { logAtf(levelWarn, warningPrefix, format, v) }
func Error(v ...any)                 { logAt(levelError, errorPrefix, v) }
func Errorf(format string, v ...any) { logAtf(levelError, errorPrefix, format, v) }

func Fatal(v ...any) {
	logAt(levelFatal, fatalPrefix, v)
	os.Exit(1)
}

func Fatalf(format string, v ...any) {
	logAtf(levelFatal, fatalPrefix, format, v)
	os.Exit(1)
}

func logAt(l logLevel, prefix string, v []any) {
	if level.Load() <= uint32(l) {
		msg := strings.TrimRight(fmt.Sprintln(v...), "\n")
		output(showCallerAsSource, prefix+" "+msg)
	}
}

func output(calldepth int, msg string) {
	if err := log.Output(calldepth, msg); err != nil {
		fmt.Fprintf(os.Stderr, "log.Output: %v\n", err)
	}
}

func logAtf(l logLevel, prefix, format string, v []any) {
	if level.Load() <= uint32(l) {
		output(showCallerAsSource, fmt.Sprintf(prefix+" "+format, v...))
	}
}
