package internal

import (
	"log"
	"os"
)

// UpdateSocketFilePermissions set socket file permissions
func UpdateSocketFilePermissions(sockFileName string) {
	if !FileExists(sockFileName) {
		return
	}
	// #nosec G302 -- need world writable permissions
	if err := os.Chmod(sockFileName, PermUserRWGroupRWOthersRW); err != nil {
		log.Println(ErrorPrefix, err)
	}
}
