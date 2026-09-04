package log

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var (
	benchErr = errors.New("connection refused")
	benchStr = "meshnet peer"
	benchInt = 51820
)

func benchEnv(b *testing.B, out io.Writer, flags int, lvl logLevel) {
	b.Helper()
	oldFlags := log.Flags()
	oldLevel := level.Load()
	log.SetOutput(out)
	log.SetFlags(flags)
	level.Store(uint32(lvl))
	b.Cleanup(func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(oldFlags)
		level.Store(oldLevel)
	})
	b.ReportAllocs()
}

const prodFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

// BenchmarkWriter shows how much of a real log call is the write itself.
func BenchmarkWriter(b *testing.B) {
	b.Run("Discard", func(b *testing.B) {
		benchEnv(b, io.Discard, prodFlags, levelDebug)
		for b.Loop() {
			Info(benchStr, benchInt, benchErr)
		}
	})
	b.Run("File", func(b *testing.B) {
		f, err := os.Create(filepath.Join(b.TempDir(), "bench.log"))
		if err != nil {
			b.Fatal(err)
		}
		b.Cleanup(func() { f.Close() })
		benchEnv(b, f, prodFlags, levelDebug)
		for b.Loop() {
			Info(benchStr, benchInt, benchErr)
		}
	})
	b.Run("BufferedFile", func(b *testing.B) {
		f, err := os.Create(filepath.Join(b.TempDir(), "bench.log"))
		if err != nil {
			b.Fatal(err)
		}
		w := bufio.NewWriter(f)
		b.Cleanup(func() { w.Flush(); f.Close() })
		benchEnv(b, w, prodFlags, levelDebug)
		for b.Loop() {
			Info(benchStr, benchInt, benchErr)
		}
	})
}
