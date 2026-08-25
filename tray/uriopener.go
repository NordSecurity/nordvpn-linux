package tray

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/godbus/dbus/v5"
)

// OpenURI opens uri via the desktop portal, falling back to xdg-open if the portal call fails
func OpenURI(uri string) error {
	portalErr := openURIViaPortal(uri)
	if portalErr == nil {
		return nil
	}

	log.Warnf("portal open failed for %q (%v), trying xdg-open", uri, portalErr)
	// #nosec G204 -- callers pass fixed URIs, no user input
	if xdgErr := exec.Command("xdg-open", uri).Run(); xdgErr != nil {
		return fmt.Errorf("opening URI %q failed via portal (%v) and xdg-open (%w)", uri, portalErr, xdgErr)
	}
	log.Infof("opened via xdg-open as a fallback: %q", uri)
	return nil
}

// openURIViaPortal requests the freedesktop.desktop.portal (over the D-Bus) to open uri.
func openURIViaPortal(uri string) error {
	// using a private connection for a specific short-lived task
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("connecting to session bus: %w", err)
	}
	defer conn.Close()

	// subscribe before the actual call so a response is not missed
	matchRules := []dbus.MatchOption{
		dbus.WithMatchInterface("org.freedesktop.portal.Request"),
		dbus.WithMatchMember("Response"),
	}
	if err := conn.AddMatchSignal(matchRules...); err != nil {
		return fmt.Errorf("subscribing to portal's 'Response': %w", err)
	}
	rxChan := make(chan *dbus.Signal, 8)
	conn.Signal(rxChan)

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")
	ctx, cancel := context.WithTimeout(context.Background(), dbusCallTimeout)
	defer cancel()

	log.Debugf("portal: OpenURI for %q", uri)
	var requestPath dbus.ObjectPath
	call := obj.CallWithContext(ctx, "org.freedesktop.portal.OpenURI.OpenURI", 0,
		"", uri, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("portal OpenURI call failed: %w", call.Err)
	}
	if err := call.Store(&requestPath); err != nil {
		return fmt.Errorf("storing portal request handle: %w", err)
	}
	log.Debugf("portal: OpenURI accepted, waiting for response on %q", requestPath)

	timeout := time.After(dbusCallTimeout)
	for {
		select {
		case <-timeout:
			// the portal already accepted the request, so a late "Response" most
			// likely means that the launch was dispatched
			log.Warnf("portal 'Response' for %q timed out, assuming launch was dispatched", uri)
			return nil
		case sig := <-rxChan:
			if sig == nil || sig.Path != requestPath || len(sig.Body) == 0 {
				continue
			}
			code, ok := sig.Body[0].(uint32)
			if !ok {
				return fmt.Errorf("malformed portal response for %q: %v", uri, sig.Body)
			}
			log.Infof("portal: OpenURI response for %q: code=%d body=%v", uri, code, sig.Body)
			if code == 0 {
				return nil
			}
			return fmt.Errorf("portal OpenURI failed for %q (response code %d)", uri, code)
		}
	}
}
