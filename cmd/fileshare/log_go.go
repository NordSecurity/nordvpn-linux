//go:build !cgo

package main

import "os"

// logSetup is a placeholder function for non CGo builds
func logSetup(f *os.File) {}
