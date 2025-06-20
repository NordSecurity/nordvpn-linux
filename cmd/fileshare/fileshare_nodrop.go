//go:build !drop

package main

import "github.com/NordSecurity/nordvpn-linux/fileshare"

func newFileshare(
	eventCallback fileshare.EventCallback,
	eventsDBPath string,
	isProd bool,
	pubkeyFunc func(string) []byte,
	privateKey string,
	storagePath string,
) (fileshare.Fileshare, fileshare.Storage, error) {
	return nil, nil, nil
}
