package tray

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/filewatch"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
	"github.com/godbus/dbus/v5"
)

const dbusCallTimeout = 3 * time.Second

// The pattern for actions is to return 'true' on success and 'false' (along with emitting a notification) on failure

func (ti *Instance) login() {
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil {
		log.Error("Failed to login:", err)
		ti.notify(NoForce, "Login failed")
		return
	}
	if resp.Status == pb.LoginStatus_CONSENT_MISSING {
		// ask user for consent by opening terminal with consent flow,
		if err := openURI(internal.SubcommandURI(internal.ConsentSubcommand)); err != nil {
			log.Errorf("failed to open consent URI: %v", err)
		}
		return
	}

	if resp.GetIsLoggedIn() {
		ti.notify(NoForce, "You are already logged in")
		return
	}

	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_LOGIN,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		ItemType:      pb.UIEvent_CLICK,
	})
	loginResp, err := ti.client.LoginOAuth2(
		context.Background(),
		&pb.LoginOAuth2Request{
			Type: pb.LoginType_LoginType_LOGIN,
		},
	)
	if err != nil {
		ti.notify(NoForce, "Login error: %s", err)
		return
	}

	switch loginResp.Status {
	case pb.LoginStatus_UNKNOWN_OAUTH2_ERROR:
		ti.notify(NoForce, "Login error: %s", internal.ErrUnhandled)
	case pb.LoginStatus_NO_NET:
		ti.notify(NoForce, internal.ErrNoNetWhenLoggingIn.Error())
	case pb.LoginStatus_ALREADY_LOGGED_IN:
		ti.notify(NoForce, internal.ErrAlreadyLoggedIn.Error())
	case pb.LoginStatus_CONSENT_MISSING:
		// NOTE: This should never happen, because analytics consent is
		// triggered above, so at this point it should already be completed.
		log.Error("analytics consent should be already completed at this point")
		ti.notify(NoForce, internal.ErrAnalyticsConsentMissing.Error())
	case pb.LoginStatus_SUCCESS:
		if url := loginResp.Url; url != "" {
			// #nosec G204 -- user input is not passed in
			cmd := exec.Command("xdg-open", url)
			err = cmd.Run()
			if err != nil {
				log.Error("Failed to open login webpage:", err)
				// we want to force a notification here, otherwise there will be no reaction to user action
				ti.notify(Force, "Continue log in in the browser: %s", url)
			}
		}
	}
}

// openURI opens uri via the desktop portal, falling back to xdg-open if the portal call fails
func openURI(uri string) error {
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

const (
	guiBinaryName  = "nordvpn-gui"
	guiLaunchURI   = "nordvpn-gui://open"
	guiDownloadURL = "https://nordvpn.com/download/linux/"
)

// isGUIAvailable check whether the system already has the application
func isGUIAvailable() bool {
	if snapconf.IsUnderSnap() {
		return true
	}
	return internal.IsCommandAvailable(guiBinaryName)
}

// openGUI tries to open GUI application
func (ti *Instance) openGUI() {
	log.Infof("opening NordVPN GUI via %q", guiLaunchURI)
	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_OPEN_APP,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		ItemType:      pb.UIEvent_CLICK,
	})

	if err := openURI(guiLaunchURI); err != nil {
		log.Error("Failed to open GUI:", err)
		ti.notify(Force, "Failed to open the NordVPN app")
	}
}

// openGUIDownloadPage tries to open download page for GUI application
func (ti *Instance) openGUIDownloadPage() {
	log.Infof("opening NordVPN GUI download page via %q", guiDownloadURL)
	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_DOWNLOAD_APP,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		ItemType:      pb.UIEvent_CLICK,
	})

	if err := openURI(guiDownloadURL); err != nil {
		log.Error("Failed to open GUI download page:", err)
		ti.notify(Force, "Failed to open the NordVPN download page")
	}
}

// watchGUIInstallation redraws (async) the tray when the native GUI binary appears in the system
// or disappears.
func (ti *Instance) watchGUIInstallation(ctx context.Context) {
	if snapconf.IsUnderSnap() {
		// GUI is always bundled with the package
		return
	}

	const guiBinDir = "/usr/bin"
	watcher, err := filewatch.GetFileWatcher(guiBinDir)
	if err != nil {
		log.Error("Failed to get watcher for GUI installation:", err)
		return
	}
	defer watcher.Close()

	available := isGUIAvailable()
	for {
		select {
		case <-ctx.Done():
			// stop watching
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) != guiBinaryName {
				continue
			}
			if now := isGUIAvailable(); now != available {
				available = now
				ti.redraw(true)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("GUI installation watcher error:", err)
		}
	}
}

func (ti *Instance) logout(persistToken bool) bool {
	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_LOGOUT,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		ItemType:      pb.UIEvent_CLICK,
	})
	resp, err := ti.client.Logout(context.Background(), &pb.LogoutRequest{
		PersistToken: persistToken,
	})
	if err != nil {
		ti.notify(NoForce, "Logout error: %s", err)
		return false
	}

	switch resp.Type {
	case internal.CodeSuccess:
		return true
	case internal.CodeTokenInvalidated:
		return true
	default:
		ti.notify(NoForce, cli.CheckYourInternetConnMessage)
		return false
	}
}

func (ti *Instance) notifyServiceExpired(url string, trustedPassURL string, message string) {
	resp, err := ti.client.TokenInfo(context.Background(), &pb.Empty{})

	link := url
	if err == nil && (resp.TrustedPassToken != "" && resp.TrustedPassOwnerId != "") {
		link = fmt.Sprintf(trustedPassURL, resp.TrustedPassToken, resp.TrustedPassOwnerId)
	}

	ti.notify(Force, message, link)
}

func (ti *Instance) connect(serverTag string, serverGroup string) {
	ti.connectWithUIEvent(serverTag, serverGroup, pb.UIEvent_CONNECT, pb.UIEvent_ITEM_VALUE_UNSPECIFIED)
}

func (ti *Instance) connectWithUIEvent(
	serverTag, serverGroup string,
	itemName pb.UIEvent_ItemName,
	itemValue pb.UIEvent_ItemValue,
) bool {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer close(ch)
	go func(ch chan os.Signal) {
		for range ch {
			// #nosec G104 -- LVPN-2090
			ti.client.Disconnect(context.Background(), &pb.Empty{})
		}
	}(ch)

	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      itemName,
		ItemType:      pb.UIEvent_CLICK,
		ItemValue:     itemValue,
	})
	resp, err := ti.client.Connect(context.Background(), &pb.ConnectRequest{
		ServerTag:   strings.ToLower(serverTag),
		ServerGroup: strings.ToLower(serverGroup),
	})
	if err != nil {
		ti.notify(NoForce, "Connect error: %s", err)
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.notify(NoForce, "Connect error: %s", err)
			return false
		}

		switch out.Type {
		case internal.CodeFailure:
			ti.notify(NoForce, "Connect error: %s", client.ConnectCantConnect)
		case internal.CodeExpiredRenewToken:
			ti.notify(NoForce, client.RelogRequest)
			ti.login()
			return ti.connectWithUIEvent(serverTag, serverGroup, itemName, itemValue)
		case internal.CodeTokenRenewError:
			ti.notify(NoForce, client.AccountTokenRenewError)
		case internal.CodeAccountExpired:
			ti.notifyServiceExpired(client.SubscriptionURL, client.SubscriptionURLLogin, cli.ExpiredAccountMessage)
		case internal.CodeDedicatedIPRenewError:
			ti.notifyServiceExpired(client.SubscriptionDedicatedIPURL, client.SubscriptionDedicatedIPURLLogin, cli.NoDedicatedIPMessage)
		case internal.CodeDisconnected:
			ti.notify(NoForce, client.ConnectCanceled, internal.StringsToInterfaces(out.Data)...)
		case internal.CodeTagNonexisting:
			ti.notify(NoForce, internal.TagNonexistentErrorMessage)
		case internal.CodeGroupNonexisting:
			ti.notify(NoForce, internal.GroupNonexistentErrorMessage)
		case internal.CodeServerUnavailable:
			ti.notify(NoForce, internal.ServerUnavailableErrorMessage)
		case internal.CodeVirtualLocationDisabled:
			ti.notify(NoForce, internal.ServerUnavailableErrorMessage)
		case internal.CodeDoubleGroupError:
			ti.notify(NoForce, internal.DoubleGroupErrorMessage)
		case internal.CodeVPNRunning:
			ti.notify(NoForce, client.ConnectConnected)
		case internal.CodeNothingToDo:
			ti.notify(NoForce, client.ConnectConnecting)
		case internal.CodeUFWDisabled:
			ti.notify(NoForce, client.UFWDisabledMessage)
		case internal.CodeDedicatedServersRenewError:
			ti.notifyServiceExpired(client.DedicatedServersUpselURL, client.DedicatedServersUpselURLLogin, cli.DedicatedServersNoServiceMessage)
		case internal.CodeDedicatedServersServiceButNoServers:
			ti.notifyServiceExpired(client.DedicatedServersSetupURL, client.DedicatedServersSetupURLLogin, cli.DedicatedServersNoServersAvailable)
		case internal.CodeDedicatedServersServerNotSetUp:
			ti.notifyServiceExpired(client.DedicatedServersSetupURL, client.DedicatedServersSetupURLLogin, cli.DedicatedServersNoServersAvailable)
		case internal.CodeDedicatedServersNotReady:
			ti.notify(Force, cli.DedicatedServersServerNotReadyMessage)
		case internal.CodeDedicatedServersNoNordlynx:
			ti.notify(Force, cli.DedicatedServersNoNordlynxMessage)
		case internal.CodeDedicatedServersCanNotConnect:
			ti.notify(Force, cli.DedicatedServersCanNotConnectMessage)
		case internal.CodeDedicatedServersSessionMaxLimitReached:
			ti.notify(Force, cli.DedicatedServersConnectionLimitReached)
		case internal.CodeDedicatedServersPq:
			ti.notify(Force, internal.ServerUnavailableErrorMessage)
		case internal.CodeConnecting:
		case internal.CodeConnected:
			return true
		}
	}

	return false
}

func (ti *Instance) disconnect(itemName pb.UIEvent_ItemName, itemValue pb.UIEvent_ItemValue) bool {
	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      itemName,
		ItemValue:     itemValue,
		ItemType:      pb.UIEvent_CLICK,
	})
	resp, err := ti.client.Disconnect(context.Background(), &pb.Empty{})
	if err != nil {
		ti.notify(NoForce, "Disconnect error: %s", err)
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.notify(NoForce, "Disconnect error: %s", err)
			return false
		}

		switch out.Type {
		case internal.CodeVPNNotRunning:
			ti.notify(NoForce, cli.DisconnectNotConnected)
		case internal.CodeDisconnected:
		}
	}
	return true
}

func (ti *Instance) pause(pauseLength pauseLength) bool {
	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_PAUSE,
		ItemValue:     pauseLength.EventValue,
		ItemType:      pb.UIEvent_CLICK,
	})
	resp, err := ti.client.PauseConnection(context.Background(), &pb.PauseRequest{Seconds: pauseLength.DurationSeconds})
	if err != nil {
		ti.notify(NoForce, "Pause failed. Please try again.")
		return false
	}

	switch resp.Type {
	case internal.CodePauseAttemptWhenConnectedToMeshPeer:
		log.Error("Pause attempt when connected to meshnet peer")
		ti.notify(NoForce, "Pause is not available while connected to a Meshnet device.")
		return false
	case internal.CodeFailure:
		log.Error("Pause attempt failed")
		ti.notify(NoForce, "Pause failed. Please try again.")
		return false
	}
	return true
}

func (ti *Instance) setNotify(flag bool) bool {
	flagText := getFlagText(flag)
	resp, err := ti.client.SetNotify(context.Background(), &pb.SetNotifyRequest{
		Notify: flag,
	})
	if err != nil {
		log.Errorf("Setting notifications %s error: %s", flagText, err)
		ti.notify(NoForce, "Setting notifications %s error: %s", flagText, err)
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Errorf("Setting notifications %s error: %s", flagText, "Config file error")
		ti.notify(NoForce, "Setting notifications %s error: %s", flagText, "Config file error")
		return false
	case internal.CodeNothingToDo:
	case internal.CodeSuccess:
	}

	ti.fileshare.SetNotifications(flag)

	if resp.Type == internal.CodeNothingToDo {
		ti.notify(NoForce, "Notifications already %s", flagText)
	}

	return true
}

func (ti *Instance) setTray(flag bool) bool {
	flagText := getFlagText(flag)

	if !flag {
		log.Info("Tray icon disabled. To enable it again, run the \"nordvpn set tray on\" command.")
		ti.notify(Force, "Tray icon disabled. To enable it again, run the \"nordvpn set tray on\" command.")
	}

	resp, err := ti.client.SetTray(context.Background(), &pb.SetTrayRequest{
		Tray: flag,
	})
	if err != nil {
		log.Errorf("Setting tray %s error: %s", flagText, err)
		ti.notify(NoForce, "Setting tray %s error: %s", flagText, err)
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Errorf("Setting tray %s error: %s", flagText, "Config file error")
		ti.notify(NoForce, "Setting tray %s error: %s", flagText, "Config file error")
		return false
	case internal.CodeNothingToDo:
		ti.notify(NoForce, "Tray already %s", flagText)
	case internal.CodeSuccess:
	}

	return true
}

func getFlagText(flag bool) string {
	if flag {
		return "on"
	}
	return "off"
}
