//go:build moose

package main

import (
	"log"
	"os"

	"github.com/NordSecurity/nordvpn-linux/events/moose"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

func drainStart(dbPath string) {
	res := moose.DrainStart(dbPath)
	log.Println(internal.InfoPrefix, "moose drain start status:", res)

	if err := os.Chmod(dbPath, internal.PermUserRW); err != nil {
		log.Println(internal.ErrorPrefix, "error on setting moose db permissions:", err)
	}
}
