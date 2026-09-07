package tray

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/alert"
	"github.com/NordSecurity/nordvpn-linux/cli"
	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/core"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"
)

const actionKeyOpenHelpGuide = "open-connection-limit-help-guide"

// gatedNotifier suppresses every alert, even Urgent ones,
// until the tray has completed its sync with the daemon.
type gatedNotifier struct {
	alert.Notifier
	isReady func() bool
}

func (g *gatedNotifier) Alert(body string) *alert.AlertBuilder {
	if !g.isReady() {
		log.Systray.Infof("Notification suppressed (initial sync not completed): %s", body)
		return alert.NewAlertBuilder(func(alert.Alert) bool { return false }, body)
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
	case internal.CodeConnectionLimitReached:
		return ti.n.Alert(client.ENSConnectionLimitReached(core.TrayAppID)).
			Summary(client.ENSConnectionLimitReachedSummary).
			Action(actionKeyOpenHelpGuide, "Open help guide", func() {
				if err := ti.openURI(core.ConnectionLimitReachedGuideURL(core.TrayAppID)); err != nil {
					log.Systray.Errorf("failed to open URI: %v", err)
				}

				_, _ = ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
					FormReference: pb.UIEvent_TRAY,
					ItemName:      pb.UIEvent_SESSION_LIMIT,
					ItemType:      pb.UIEvent_CLICK,
					ItemValue:     pb.UIEvent_ITEM_VALUE_LEARN_MORE,
				})
			}).
			OnShown(func() {
				_, _ = ti.client.ReportUIEvent(context.Background(), &pb.UIEvent{
					FormReference: pb.UIEvent_TRAY,
					ItemName:      pb.UIEvent_SESSION_LIMIT,
					ItemType:      pb.UIEvent_SHOW,
					ItemValue:     pb.UIEvent_ITEM_VALUE_UNSPECIFIED,
				})
			}).
			Urgent()
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
