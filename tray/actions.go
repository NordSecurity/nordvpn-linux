package tray

import (
	"context"
	"io"
	"os"
	"os/exec"
	"os/signal"

	"github.com/NordSecurity/nordvpn-linux/cli"
	nordclient "github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
)

// The pattern for actions is to return 'true' on success and 'false' (along with emitting a notification) on failure

func (ti *Instance) login() {
	resp, err := ti.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil || resp.GetValue() {
		ti.notify("You are already logged in")
		return
	}

	cl, err := ti.client.LoginOAuth2(
		context.Background(),
		&pb.Empty{},
	)
	if err != nil {
		ti.notify("Login error: %s", err)
		return
	}

	for {
		resp, err := cl.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.notify("Login error: %s", err)
			return
		}

		if url := resp.GetData(); url != "" {
			// #nosec G204 -- user input is not passed in
			cmd := exec.Command("xdg-open", url)
			err = cmd.Start()
			if err != nil {
				ti.notify("Failed to start xdg-open: %v", err)
			}
			err = cmd.Wait()

			if err != nil {
				ti.notify("Failed to open the web browser: %v", err)
				ti.notify("Continue log in in the browser: %s", url)
			}
		}
	}
}

func (ti *Instance) logout(persistToken bool) bool {
	payload, err := ti.client.Logout(context.Background(), &pb.LogoutRequest{
		PersistToken: persistToken,
	})
	if err != nil {
		ti.notify("Logout error: %s", err)
		return false
	}

	switch payload.Type {
	case internal.CodeSuccess:
		return true
	case internal.CodeTokenInvalidated:
		return true
	default:
		ti.notify(cli.CheckYourInternetConnMessage)
		return false
	}
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
		ServerTag:   serverTag,
		ServerGroup: serverGroup,
	})
	if err != nil {
		ti.notify("Connect error: %s", err)
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.notify("Connect error: %s", err)
			return false
		}

		switch out.Type {
		case internal.CodeFailure:
			ti.notify("Connect error: %s", nordclient.ConnectCantConnect)
		case internal.CodeExpiredRenewToken:
			ti.notify(nordclient.RelogRequest)
			ti.login()
			return ti.connect(serverTag, serverGroup)
		case internal.CodeTokenRenewError:
			ti.notify(nordclient.AccountTokenRenewError)
		case internal.CodeAccountExpired:
			ti.notify(cli.ErrAccountExpired.Error())
		case internal.CodeDisconnected:
			ti.notify(internal.DisconnectSuccess)
		case internal.CodeTagNonexisting:
			ti.notify(internal.TagNonexistentErrorMessage)
		case internal.CodeGroupNonexisting:
			ti.notify(internal.GroupNonexistentErrorMessage)
		case internal.CodeServerUnavailable:
			ti.notify(internal.ServerUnavailableErrorMessage)
		case internal.CodeDoubleGroupError:
			ti.notify(internal.DoubleGroupErrorMessage)
		case internal.CodeVPNRunning:
			ti.notify(nordclient.ConnectConnected)
		case internal.CodeUFWDisabled:
			ti.notify(nordclient.UFWDisabledMessage)
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
		ti.notify("Disconnect error: %s", err)
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			ti.notify("Disconnect error: %s", err)
			return false
		}

		switch out.Type {
		case internal.CodeVPNNotRunning:
			ti.notify(cli.DisconnectNotConnected)
		case internal.CodeDisconnected:
		}
	}
	return true
}
