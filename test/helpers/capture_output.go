package helpers

import (
	"bytes"
	"io"
	"os"
	"strings"
)

// CaptureOutput will capture the stdout + stderr of a function execution and return it as a string
func CaptureOutput(f func()) (string, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()

	os.Stdout = writer
	os.Stderr = writer

	f()

	writer.Close() // close to unblock io.Copy(&buf, reader)
	var buf bytes.Buffer
	io.Copy(&buf, reader)
	return strings.TrimSuffix(buf.String(), "\n"), nil
}
