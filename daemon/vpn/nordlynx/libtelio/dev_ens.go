package libtelio

import (
	"errors"
	"sync/atomic"

	telio "github.com/NordSecurity/libtelio-go/v6"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

type ensDev struct {
	ch                      chan vpnConnError
	nextConnectLimitReached atomic.Bool
}

func newEnsDev(prod bool) *ensDev {
	var ch chan vpnConnError
	if !prod {
		ch = make(chan vpnConnError)
	}
	return &ensDev{
		ch: ch,
	}
}

func (ed *ensDev) startNewConnect() {
	if ed.ch == nil {
		return
	}
	if ed.nextConnectLimitReached.Load() {
		ed.ch <- vpnConnError{code: telio.VpnConnectionErrorConnectionLimitReached}
		ed.nextConnectLimitReached.Store(false)
	}
}

func (ed *ensDev) add(event vpnConnError) error {
	if ed.ch == nil {
		return errors.New("ENS injection is unavailable outside the dev environment")
	}
	if event.code == telio.VpnConnectionErrorConnectionLimitReached {
		ed.nextConnectLimitReached.Store(true)
	} else {
		if !internal.TrySend(ed.ch, event) {
			return errors.New("dropping event")
		}
	}

	return nil
}
