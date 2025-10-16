package tray

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	nordclient "github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/godbus/dbus/v5"
)

// The pattern for actions is to return 'true' on success and 'false' (along with emitting a notification) on failure

func (ti *Instance) login() {
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println(internal.ErrorPrefix, "Failed to login:", err)
		ti.notify(NoForce, "Login failed")
		return
	}
	if resp.Status == pb.LoginStatus_CONSENT_MISSING {
		// ask user for consent by opening terminal with consent flow,
		openURI(internal.SubcommandURI(internal.ConsentSubcommand))
		return
	}

	if resp.GetIsLoggedIn() {
		ti.notify(NoForce, "You are already logged in")
		return
	}

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
		log.Println(internal.ErrorPrefix, "analytics consent should be already completed at this point")
		ti.notify(NoForce, internal.ErrAnalyticsConsentMissing.Error())
	case pb.LoginStatus_SUCCESS:
		if url := loginResp.Url; url != "" {
			// #nosec G204 -- user input is not passed in
			cmd := exec.Command("xdg-open", url)
			err = cmd.Run()
			if err != nil {
				log.Println(internal.ErrorPrefix, "Failed to open login webpage:", err)
				// we want to force a notification here, otherwise there will be no reaction to user action
				ti.notify(Force, "Continue log in in the browser: %s", url)
			}
		}
	}
}

func openURI(uri string) {
	if err := tryDbus(uri); err != nil {
		log.Printf(internal.ErrorPrefix+" failed to open URI '%s' using D-Bus: %v\n", uri, err)
	}
}

func tryDbus(uri string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	obj := conn.Object("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	call := obj.CallWithContext(ctx,
		"org.freedesktop.portal.OpenURI.OpenURI", 0,
		"", uri, map[string]dbus.Variant{},
	)
	if call.Err != nil {
		return fmt.Errorf("DBus OpenURI failed: %w", call.Err)
	}

	return nil
}

func (ti *Instance) logout(persistToken bool) bool {
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

func (ti *Instance) connect(serverTag string, serverGroup string) bool {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer close(ch)
	go func(ch chan os.Signal) {
		for range ch {
			// #nosec G104 -- LVPN-2090
			ti.client.Disconnect(context.Background(), &pb.Empty{})
		}
	}(ch)

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
			ti.notify(NoForce, "Connect error: %s", nordclient.ConnectCantConnect)
		case internal.CodeExpiredRenewToken:
			ti.notify(NoForce, nordclient.RelogRequest)
			ti.login()
			return ti.connect(serverTag, serverGroup)
		case internal.CodeTokenRenewError:
			ti.notify(NoForce, nordclient.AccountTokenRenewError)
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
		case internal.CodeDoubleGroupError:
			ti.notify(NoForce, internal.DoubleGroupErrorMessage)
		case internal.CodeVPNRunning:
			ti.notify(NoForce, nordclient.ConnectConnected)
		case internal.CodeNothingToDo:
			ti.notify(NoForce, nordclient.ConnectConnecting)
		case internal.CodeUFWDisabled:
			ti.notify(NoForce, nordclient.UFWDisabledMessage)
		case internal.CodeConnecting:
		case internal.CodeConnected:
			return true
		}
	}

	return false
}

func (ti *Instance) disconnect() bool {
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

func (ti *Instance) setNotify(flag bool) bool {
	flagText := "off"
	if flag {
		flagText = "on"
	}
	resp, err := ti.client.SetNotify(context.Background(), &pb.SetNotifyRequest{
		Notify: flag,
	})
	if err != nil {
		log.Printf("%s Setting notifications %s error: %s", internal.ErrorPrefix, flagText, err)
		ti.notify(NoForce, "Setting notifications %s error: %s", flagText, err)
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Printf("%s Setting notifications %s error: %s", internal.ErrorPrefix, flagText, "Config file error")
		ti.notify(NoForce, "Setting notifications %s error: %s", flagText, "Config file error")
		return false
	case internal.CodeNothingToDo:
	case internal.CodeSuccess:
	}

	_, err = ti.fileshareClient.SetNotifications(context.Background(), &filesharepb.SetNotificationsRequest{Enable: flag})
	if err != nil {
		log.Printf("%s Setting fileshare notifications %s error: %s", internal.ErrorPrefix, flagText, err)
	}

	if resp.Type == internal.CodeNothingToDo {
		ti.notify(NoForce, "Notifications already %s", flagText)
	}

	return true
}

func (ti *Instance) setTray(flag bool) bool {
	flagText := "off"
	if flag {
		flagText = "on"
	}

	if !flag {
		log.Printf("%s Tray icon disabled. To enable it again, run the \"nordvpn set tray on\" command.", internal.InfoPrefix)
		ti.notify(Force, "Tray icon disabled. To enable it again, run the \"nordvpn set tray on\" command.")
	}

	resp, err := ti.client.SetTray(context.Background(), &pb.SetTrayRequest{
		Uid:  int64(os.Getuid()),
		Tray: flag,
	})
	if err != nil {
		log.Printf("%s Setting tray %s error: %s", internal.ErrorPrefix, flagText, err)
		ti.notify(NoForce, "Setting tray %s error: %s", flagText, err)
		return false
	}

	switch resp.Type {
	case internal.CodeConfigError:
		log.Printf("%s Setting tray %s error: %s", internal.ErrorPrefix, flagText, "Config file error")
		ti.notify(NoForce, "Setting tray %s error: %s", flagText, "Config file error")
		return false
	case internal.CodeNothingToDo:
		ti.notify(NoForce, "Tray already %s", flagText)
	case internal.CodeSuccess:
	}

	return true
}
