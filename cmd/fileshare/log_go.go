//go:build !cgo

package main

import "os"

// redirectNativeOutput is a placeholder function for non CGo builds
func redirectNativeOutput(f *os.File) {}
