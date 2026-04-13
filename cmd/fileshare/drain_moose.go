//go:build moose

package main

import (
	"os"

	"github.com/NordSecurity/nordvpn-linux/events/moose"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

func drainStart(dbPath string) {
	res := moose.DrainStart(dbPath)
	log.Info("moose drain start status:", res)

	if err := os.Chmod(dbPath, internal.PermUserRW); err != nil {
		log.Error("error on setting moose db permissions:", err)
	}
}
