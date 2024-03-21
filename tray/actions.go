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
	meshpb "github.com/NordSecurity/nordvpn-linux/meshnet/pb"
)

// The pattern for actions is to return 'true' on success and 'false' (along with emitting a notification) on failure

func login(client pb.DaemonClient) {
	resp, err := client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil || resp.GetValue() {
		notification("warning", "You are already logged in")
		return
	}

	cl, err := client.LoginOAuth2(
		context.Background(),
		&pb.Empty{},
	)
	if err != nil {
		notification("error", "Login error: %s", err)
		return
	}

	for {
		resp, err := cl.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			notification("error", "Login error: %s", err)
			return
		}

		if url := resp.GetData(); url != "" {
			// #nosec G204 -- user input is not passed in
			cmd := exec.Command("xdg-open", url)
			err = cmd.Start()
			if err != nil {
				notification("warning", "Failed to start xdg-open: %v", err)
			}
			err = cmd.Wait()

			if err != nil {
				notification("warning", "Failed to open the web browser: %v", err)
				notification("info", "Continue log in in the browser: %s", url)
			}
		}
	}
}

func logout(client pb.DaemonClient, persistToken bool) bool {
	payload, err := client.Logout(context.Background(), &pb.LogoutRequest{
		PersistToken: persistToken,
	})
	if err != nil {
		notification("error", "Logout error: %s", err)
		return false
	}

	switch payload.Type {
	case internal.CodeSuccess:
		if !NotifyEnabled {
			notification("info", cli.LogoutSuccess)
		}
		return true
	case internal.CodeTokenInvalidated:
		if !NotifyEnabled {
			notification("info", cli.LogoutTokenSuccess)
		}
		return true
	default:
		notification("error", cli.CheckYourInternetConnMessage)
		return false
	}
}

func connect(client pb.DaemonClient, serverTag string, serverGroup string) bool {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer close(ch)
	go func(ch chan os.Signal) {
		for range ch {
			// #nosec G104 -- LVPN-2090
			client.Disconnect(context.Background(), &pb.Empty{})
		}
	}(ch)

	resp, err := client.Connect(context.Background(), &pb.ConnectRequest{
		ServerTag:   serverTag,
		ServerGroup: serverGroup,
	})
	if err != nil {
		notification("error", "Connect error: %s", err)
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			notification("error", "Connect error: %s", err)
			return false
		}

		switch out.Type {
		case internal.CodeFailure:
			notification("error", "Connect error: %s", nordclient.ConnectCantConnect)
		case internal.CodeExpiredRenewToken:
			notification("warning", nordclient.RelogRequest)
			login(client)
			return connect(client, serverTag, serverGroup)
		case internal.CodeTokenRenewError:
			notification("error", nordclient.AccountTokenRenewError)
		case internal.CodeAccountExpired:
			notification("error", cli.ErrAccountExpired.Error())
		case internal.CodeDisconnected:
			notification("info", internal.DisconnectSuccess)
		case internal.CodeTagNonexisting:
			notification("error", internal.TagNonexistentErrorMessage)
		case internal.CodeGroupNonexisting:
			notification("error", internal.GroupNonexistentErrorMessage)
		case internal.CodeServerUnavailable:
			notification("error", internal.ServerUnavailableErrorMessage)
		case internal.CodeDoubleGroupError:
			notification("error", internal.DoubleGroupErrorMessage)
		case internal.CodeVPNRunning:
			notification("warning", nordclient.ConnectConnected)
		case internal.CodeUFWDisabled:
			notification("warning", nordclient.UFWDisabledMessage)
		case internal.CodeConnecting:
		case internal.CodeConnected:
			return true
		}
	}

	return false
}

func disconnect(client pb.DaemonClient) bool {
	resp, err := client.Disconnect(context.Background(), &pb.Empty{})
	if err != nil {
		notification("error", "Disconnect error: %s", err)
		return false
	}

	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			notification("error", "Disconnect error: %s", err)
			return false
		}

		switch out.Type {
		case internal.CodeVPNNotRunning:
			notification("warning", cli.DisconnectNotConnected)
		case internal.CodeDisconnected:
			if !NotifyEnabled {
				notification("info", internal.DisconnectSuccess)
			}
		}
	}
	return true
}

func enableMeshnet(meshClient meshpb.MeshnetClient) bool {
	resp, err := meshClient.EnableMeshnet(context.Background(), &meshpb.Empty{})
	if err != nil {
		notification("error", "Enable meshnet error: %s", err)
		return false
	}
	if err := cli.MeshnetResponseToError(resp); err != nil {
		notification("error", "Enable meshnet error: %s", err)
		return false
	}

	if !NotifyEnabled {
		notification("info", cli.MsgSetMeshnetSuccess, "enabled")
	}

	// TODO: c.fileshareProcessManager.StartProcess() is called here in the CLI
	return true
}

func disableMeshnet(meshClient meshpb.MeshnetClient) bool {
	resp, err := meshClient.DisableMeshnet(context.Background(), &meshpb.Empty{})
	if err != nil {
		notification("error", "Disable meshnet error: %s", err)
		return false
	}
	if err := cli.MeshnetResponseToError(resp); err != nil {
		notification("error", "Disable meshnet error: %s", err)
		return false
	}

	if !NotifyEnabled {
		notification("info", cli.MsgSetMeshnetSuccess, "disabled")
	}

	return true
}
