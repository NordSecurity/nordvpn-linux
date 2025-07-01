//go:build !drop

package main

import "github.com/NordSecurity/nordvpn-linux/fileshare"

func newFileshare(
	fileshare.EventCallback,
	string,
	bool,
	func(string) []byte,
	string,
	string,
) (fileshare.Fileshare, fileshare.Storage, error) {
	return nil, nil, nil
}
