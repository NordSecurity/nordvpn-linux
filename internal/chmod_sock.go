package internal

import (
	"log"
	"os"
)

// UpdateFilePermissions sets permissions of a given file if it exists and logs the error to stdout
func UpdateFilePermissions(name string, mode os.FileMode) {
	if !FileExists(name) {
		return
	}
	// #nosec G302 -- need world writable permissions
	if err := os.Chmod(name, mode); err != nil {
		log.Println(ErrorPrefix, err)
	}
}
