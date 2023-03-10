// Utility for validating glibc version of an elf binary.
package main

import (
	"debug/elf"
	"log"
	"os"
	"strings"

	"golang.org/x/mod/semver"
)

const (
	libcBinaryName   = "libc.so.6"
	libcSymbolPrefix = "GLIBC"
)

func isGreater(version string, than string) bool {
	return semver.Compare(version, than) == 1
}

func requiresLibc(symbol elf.Symbol) bool {
	return symbol.Library == libcBinaryName ||
		strings.HasPrefix(symbol.Version, libcSymbolPrefix) // for example, libpthread.so
}

func main() {
	path, glibcVersion := os.Args[1], os.Args[2]
	file, err := elf.Open(path)
	if err != nil {
		log.Fatalln(err)
	}

	symbols, err := file.DynamicSymbols()
	if err != nil {
		log.Fatalln(err)
	}

	versionSet := map[string]bool{}
	for _, symbol := range symbols {
		if requiresLibc(symbol) {
			versionSet[symbol.Version] = true
		}
	}

	var semvers []string
	for symbol := range versionSet {
		_, version, ok := strings.Cut(symbol, "_")
		if ok {
			// v prefix is required by the semver library
			// without it neither comparisons nor sorting works and just panics
			semvers = append(semvers, "v"+version)
		}
	}

	semver.Sort(semvers)
	lastSymbol := semvers[len(semvers)-1]
	if isGreater(lastSymbol, "v"+glibcVersion) {
		log.Fatalf(
			"%s requires %s version of the glibc, which exceeds the expected %s\n",
			path,
			strings.TrimPrefix(lastSymbol, "v"),
			glibcVersion,
		)
	}
}
