//go:build drop

package main

import (
	"github.com/NordSecurity/nordvpn-linux/fileshare"
	"github.com/NordSecurity/nordvpn-linux/fileshare/libdrop"
)

func newFileshare(
	eventCallback fileshare.EventCallback,
	eventsDBPath string,
	isProd bool,
	pubkeyFunc func(string) []byte,
	privateKey string,
	storagePath string,
) (fileshare.Fileshare, fileshare.Storage, error) {
	fs, err := libdrop.New(
		eventCallback,
		eventsDBPath,
		isProd,
		pubkeyFunc,
		privateKey,
		storagePath,
	)
	return fs, fs, err
}
