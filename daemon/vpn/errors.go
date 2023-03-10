package vpn

import "errors"

var (
	ErrVPNAIsAlreadyStarted = errors.New("vpn is already started")
	ErrTunnelAlreadyExists  = errors.New("tunnel already exists")
)
