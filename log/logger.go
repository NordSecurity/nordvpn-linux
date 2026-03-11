package log

import (
	"io"
	"log"
)

const (
	Lmicroseconds = log.Lmicroseconds
	Lshortfile    = log.Lshortfile
	LstdFlags     = log.Ldate | log.Ltime
)

func SetOutput(w io.Writer) {
	log.SetOutput(w)
}

func SetFlags(flag int) {
	log.SetFlags(flag)
}

func Print(v ...any) {
	log.Print(v...)
}

func Println(v ...any) {
	log.Println(v...)
}

func Printf(format string, v ...any) {
	log.Printf(format, v...)
}

func Fatal(v ...any) {
	log.Fatal(v...)
}

func Fatalln(v ...any) {
	log.Fatalln(v...)
}

func Fatalf(format string, v ...any) {
	log.Fatalf(format, v...)
}
