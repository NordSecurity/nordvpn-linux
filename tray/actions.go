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
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/filewatch"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
	"github.com/NordSecurity/nordvpn-linux/snapconf"
)

const dbusCallTimeout = 3 * time.Second

// The pattern for actions is to return 'true' on success and 'false' (along with emitting a notification) on failure

func getDesktopEnvironment() ([]string, error) {
	environment := []string{}
	out, err := exec.Command("systemctl", "--user", "show-environment").Output()
	if err != nil {
		return environment, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "DISPLAY=") ||
			strings.HasPrefix(line, "WAYLAND_DISPLAY=") ||
			strings.HasPrefix(line, "XAUTHORITY=") ||
			strings.HasPrefix(line, "DBUS_SESSION_BUS_ADDRESS=") {
			environment = append(environment, line)
		}
	}
	return environment, nil
}

func (ti *Instance) login() {
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil {
		log.Error("Failed to login:", err)
		ti.n.Alert("Login failed").Show()
		return
	}
	if resp.Status == pb.LoginStatus_CONSENT_MISSING {
		// ask user for consent by opening terminal with consent flow,
		if err := ti.openURI(internal.SubcommandURI(internal.ConsentSubcommand)); err != nil {
			log.Errorf("failed to open consent URI: %v", err)
		}
		return
	}

	if resp.GetIsLoggedIn() {
		ti.n.Alert("You are already logged in").Show()
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
		ti.n.Alert(fmt.Sprintf("Login error: %s", err)).Show()
		return
	}

	switch loginResp.Status {
	case pb.LoginStatus_UNKNOWN_OAUTH2_ERROR:
		ti.n.Alert(fmt.Sprintf("Login error: %s", internal.ErrUnhandled)).Show()
	case pb.LoginStatus_NO_NET:
		ti.n.Alert(internal.ErrNoNetWhenLoggingIn.Error()).Show()
	case pb.LoginStatus_ALREADY_LOGGED_IN:
		ti.n.Alert(internal.ErrAlreadyLoggedIn.Error()).Show()
	case pb.LoginStatus_CONSENT_MISSING:
		// NOTE: This should never happen, because analytics consent is
		// triggered above, so at this point it should already be completed.
		log.Error("analytics consent should be already completed at this point")
		ti.n.Alert(internal.ErrAnalyticsConsentMissing.Error()).Show()
	case pb.LoginStatus_SUCCESS:
		if url := loginResp.Url; url != "" {
			// #nosec G204 -- user input is not passed in
			cmd := exec.Command("xdg-open", url)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Error("Failed to open login webpage:", err)
				// we want to force a notification here, otherwise there will be no reaction to user action
				ti.n.Alert(fmt.Sprintf("Continue log in in the browser: %s", url)).Urgent().Show()
				return
			}

			if !strings.Contains(string(out), "Error: no DISPLAY environment variable specified") {
				return
			}

			log.Warn("Desktop related environment variables not set, attempting to load them manually")
			environment, err := getDesktopEnvironment()
			if err != nil {
				log.Error("Failed to read desktop environment manually:", err)
				ti.n.Alert(fmt.Sprintf("Continue log in in the browser: %s", url)).Urgent().Show()
				return
			}

			// #nosec G204 -- user input is not passed in
			cmd = exec.Command("xdg-open", url)
			cmd.Env = environment
			err = cmd.Run()
			if err != nil {
				log.Error("Failed to open login webpage with manually loaded environment:", err)
				ti.n.Alert(fmt.Sprintf("Continue log in in the browser: %s", url)).Urgent().Show()
			}
		}
	}
}

const (
	guiBinaryName      = "nordvpn-gui"
	guiLaunchURI       = "nordvpn-gui://open"
	guiDownloadPageURL = "https://nordvpn.com/download/linux/?utm_medium=app&utm_source=nordvpn-linux-tray"
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

	if err := ti.openURI(guiLaunchURI); err != nil {
		log.Error("Failed to open GUI:", err)
		ti.n.Alert("Failed to open the NordVPN app").Urgent().Show()
	}
}

// openGUIDownloadPage tries to open download page for GUI application
func (ti *Instance) openGUIDownloadPage() {
	log.Infof("opening NordVPN GUI download page via %q", guiDownloadPageURL)
	// #nosec G104 -- fire-and-forget analytics
	ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_TRAY,
		ItemName:      pb.UIEvent_DOWNLOAD_APP,
		ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
		ItemType:      pb.UIEvent_CLICK,
	})

	if err := ti.openURI(guiDownloadPageURL); err != nil {
		log.Error("Failed to open GUI download page:", err)
		ti.n.Alert("Failed to open the NordVPN download page").Urgent().Show()
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
		ti.n.Alert(fmt.Sprintf("Logout error: %s", err)).Show()
		return false
	}

	switch resp.Type {
	case internal.CodeSuccess:
		return true
	case internal.CodeTokenInvalidated:
		return true
	default:
		ti.n.Alert(cli.CheckYourInternetConnMessage).Show()
		return false
	}
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
		ti.n.Alert(fmt.Sprintf("Connect error: %s", err)).Show()
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.n.Alert(fmt.Sprintf("Connect error: %s", err)).Show()
			return false
		}

		if b := ti.connectionResultAlert(out); b != nil {
			b.Show()
		}

		if out.Type == internal.CodeExpiredRenewToken {
			ti.login()
			return ti.connectWithUIEvent(serverTag, serverGroup, itemName, itemValue)
		}

		if out.Type == internal.CodeConnected {
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
		ti.n.Alert(fmt.Sprintf("Disconnect error: %s", err)).Show()
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.n.Alert(fmt.Sprintf("Disconnect error: %s", err)).Show()
			return false
		}

		switch out.Type {
		case internal.CodeVPNNotRunning:
			ti.n.Alert(cli.DisconnectNotConnected).Show()
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
		ti.n.Alert("Pause failed. Please try again.").Show()
		return false
	}

	switch resp.Type {
	case internal.CodePauseAttemptWhenConnectedToMeshPeer:
		log.Error("Pause attempt when connected to meshnet peer")
		ti.n.Alert("Pause is not available while connected to a Meshnet device.").Show()
		return false
	case internal.CodeFailure:
		log.Error("Pause attempt failed")
		ti.n.Alert("Pause failed. Please try again.").Show()
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
		ti.n.Alert(fmt.Sprintf("Setting notifications %s error: %s", flagText, err)).Show()
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Errorf("Setting notifications %s error: %s", flagText, "Config file error")
		ti.n.Alert(fmt.Sprintf("Setting notifications %s error: %s", flagText, "Config file error")).Show()
		return false
	case internal.CodeNothingToDo:
	case internal.CodeSuccess:
	}

	ti.fileshare.SetNotifications(flag)

	if resp.Type == internal.CodeNothingToDo {
		ti.n.Alert(fmt.Sprintf("Notifications already %s", flagText)).Show()
	}

	return true
}

func (ti *Instance) setTray(flag bool) bool {
	flagText := getFlagText(flag)

	if !flag {
		log.Info("Tray icon disabled. To enable it again, run the \"nordvpn set tray on\" command.")
		ti.n.Alert("Tray icon disabled. To enable it again, run the \"nordvpn set tray on\" command.").Urgent().Show()
	}

	resp, err := ti.client.SetTray(context.Background(), &pb.SetTrayRequest{
		Tray: flag,
	})
	if err != nil {
		log.Errorf("Setting tray %s error: %s", flagText, err)
		ti.n.Alert(fmt.Sprintf("Setting tray %s error: %s", flagText, err)).Show()
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Errorf("Setting tray %s error: %s", flagText, "Config file error")
		ti.n.Alert(fmt.Sprintf("Setting tray %s error: %s", flagText, "Config file error")).Show()
		return false
	case internal.CodeNothingToDo:
		ti.n.Alert(fmt.Sprintf("Tray already %s", flagText)).Show()
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
