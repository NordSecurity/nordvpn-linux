package libtelio

import (
	"errors"
	"sync/atomic"

	telio "github.com/NordSecurity/libtelio-go/v6"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// The was extracted from libtelio to handle the ENS events injection in dev builds.
// It also handles the cases where the application is "put" in error mode,
// e.g. ENS VpnConnectionErrorConnectionLimitReached is injected, for the next VPN connect

type ensDev struct {
	ch chan vpnConnError
	// Setting this to true will result in an ENS event generated next time startNewConnect is executed.
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

// This needs to be called when libtelio starts the events monitoring, for the new VPN connect
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
