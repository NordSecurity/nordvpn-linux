package log

import (
	"io"
	"log"
)

const (
	DebugPrefix = "[Debug]"
	// DeferPrefix is used when logging errors in deferred or cleanup code.
	DeferPrefix = "[Defer]"
	// ErrorPrefix is used when logging errors, which impact control flow.
	ErrorPrefix = "[Error]"
	// WarningPrefix is used when logging errors, which don't impact control flow.
	WarningPrefix = "[Warning]"
	InfoPrefix    = "[Info]"

	LstdFlags     = log.LstdFlags
	Lshortfile    = log.Lshortfile
	Lmicroseconds = log.Lmicroseconds
)

func SetFlags(flag int) {
	log.SetFlags(flag)
}

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func Println(args ...any) {
	log.Println(args...)
}

func Printf(format string, args ...any) {
	log.Printf(format, args[1:]...)
}

func Debugln(args ...any) {
	newArgs := append([]any{DebugPrefix}, args...)
	log.Println(newArgs...)
}

func Debugf(format string, args ...any) {
	log.Printf(DebugPrefix+" "+format, args...)
}

func Deferln(args ...any) {
	newArgs := append([]any{DeferPrefix}, args...)
	log.Println(newArgs...)
}

func Infoln(args ...any) {
	newArgs := append([]any{InfoPrefix}, args...)
	log.Println(newArgs...)
}

func Infof(format string, args ...any) {
	log.Printf(InfoPrefix+" "+format, args...)
}

func Warnln(args ...any) {
	newArgs := append([]any{WarningPrefix}, args...)
	log.Println(newArgs...)
}

func Warnf(format string, args ...any) {
	log.Printf(WarningPrefix+" "+format, args...)
}

func Errorln(args ...any) {
	newArgs := append([]any{ErrorPrefix}, args...)
	log.Println(newArgs...)
}

func Errorf(args ...any) {
	log.Printf(ErrorPrefix+" "+args[0].(string), args[1:]...)
}

func Fatalln(args ...any) {
	log.Fatalln(args...)
}

func Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}
