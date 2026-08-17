package tray

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/alert"
	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

// gatedNotifier suppresses every alert, even Urgent ones,
// until the tray has completed its sync with the daemon.
type gatedNotifier struct {
	alert.Notifier
	state *trayState
}

func (g *gatedNotifier) Alert(body string) *alert.AlertBuilder {
	g.state.mu.RLock()
	ready := g.state.initialSyncCompleted
	g.state.mu.RUnlock()

	if !ready {
		log.Systray.Infof("Notification suppressed (initial sync not completed): %s", body)
		return alert.NewAlertBuilder(func(alert.Alert) {}, body)
	}

	return g.Notifier.Alert(body)
}

func (ti *Instance) connectionResultAlert(out *pb.Payload) *alert.AlertBuilder {
	switch out.Type {
	case internal.CodeFailure:
		return ti.n.Alert(fmt.Sprintf("Connect error: %s", client.ConnectCantConnect))
	case internal.CodeExpiredRenewToken:
		return ti.n.Alert(client.RelogRequest)
	case internal.CodeTokenRenewError:
		return ti.n.Alert(client.AccountTokenRenewError)
	case internal.CodeAccountExpired:
		link := ti.trustedPassURLOrDefault(client.SubscriptionURL, client.SubscriptionURLLogin)
		return ti.n.Alert(fmt.Sprintf(cli.ExpiredAccountMessage, link)).Urgent()
	case internal.CodeDedicatedIPRenewError:
		link := ti.trustedPassURLOrDefault(client.SubscriptionDedicatedIPURL, client.SubscriptionDedicatedIPURLLogin)
		return ti.n.Alert(fmt.Sprintf(cli.NoDedicatedIPMessage, link)).Urgent()
	case internal.CodeDisconnected:
		return ti.n.Alert(fmt.Sprintf(client.ConnectCanceled, internal.StringsToInterfaces(out.Data)...))
	case internal.CodeTagNonexisting:
		return ti.n.Alert(internal.TagNonexistentErrorMessage)
	case internal.CodeGroupNonexisting:
		return ti.n.Alert(internal.GroupNonexistentErrorMessage)
	case internal.CodeServerUnavailable:
		return ti.n.Alert(internal.ServerUnavailableErrorMessage)
	case internal.CodeVirtualLocationDisabled:
		return ti.n.Alert(internal.SpecifiedServerIsVirtualLocation)
	case internal.CodeDoubleGroupError:
		return ti.n.Alert(internal.DoubleGroupErrorMessage)
	case internal.CodeVPNRunning:
		return ti.n.Alert(client.ConnectConnected)
	case internal.CodeNothingToDo:
		return ti.n.Alert(client.ConnectConnecting)
	case internal.CodeUFWDisabled:
		return ti.n.Alert(client.UFWDisabledMessage)
	case internal.CodeDedicatedServersRenewError:
		link := ti.trustedPassURLOrDefault(client.DedicatedServersUpselURL, client.DedicatedServersUpselURLLogin)
		return ti.n.Alert(fmt.Sprintf(cli.DedicatedServersNoServiceMessage, link)).Urgent()
	case internal.CodeDedicatedServersServiceButNoServers:
		link := ti.trustedPassURLOrDefault(client.DedicatedServersSetupURL, client.DedicatedServersSetupURLLogin)
		return ti.n.Alert(fmt.Sprintf(cli.DedicatedServersNoServersAvailable, link)).Urgent()
	case internal.CodeDedicatedServersServerNotSetUp:
		link := ti.trustedPassURLOrDefault(client.DedicatedServersSetupURL, client.DedicatedServersSetupURLLogin)
		return ti.n.Alert(fmt.Sprintf(cli.DedicatedServersNoServersAvailable, link)).Urgent()
	case internal.CodeDedicatedServersNotReady:
		return ti.n.Alert(cli.DedicatedServersServerNotReadyMessage).Urgent()
	case internal.CodeDedicatedServersNoNordlynx:
		return ti.n.Alert(cli.DedicatedServersNoNordlynxMessage).Urgent()
	case internal.CodeDedicatedServersCanNotConnect:
		return ti.n.Alert(cli.DedicatedServersCanNotConnectMessage).Urgent()
	case internal.CodeDedicatedServersSessionMaxLimitReached:
		return ti.n.Alert(cli.DedicatedServersConnectionLimitReached).Urgent()
	case internal.CodeDedicatedServersPq:
		return ti.n.Alert(internal.ServerUnavailableErrorMessage).Urgent()
	case internal.CodeConnecting: // no notification
	case internal.CodeConnected:
		// NOTE: connection success is not handled here on purpose
	}
	return nil
}

func (ti *Instance) trustedPassURLOrDefault(defaultURL string, trustedPassURL string) string {
	resp, err := ti.client.TokenInfo(context.Background(), &pb.Empty{})

	link := defaultURL
	if err == nil && (resp.TrustedPassToken != "" && resp.TrustedPassOwnerId != "") {
		link = fmt.Sprintf(trustedPassURL, resp.TrustedPassToken, resp.TrustedPassOwnerId)
	}

	return link
}
