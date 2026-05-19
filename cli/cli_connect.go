package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/uievent"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Connect help text
const (
	ConnectUsageText          = "Connects you to VPN"
	ConnectFlagGroupUsageText = "Specify a server group to connect to"
	ConnectArgsUsageText      = "[<country>|<server>|<country_code>|<city>|<group>|<country> <city>]"
	ConnectDescription        = `Use this command to connect to NordVPN. Adding no arguments to the command will connect you to the recommended server.
Provide a <country> argument to connect to a specific country. For example: 'nordvpn connect Australia'
Provide a <server> argument to connect to a specific server. For example: 'nordvpn connect jp35'
Provide a <country_code> argument to connect to a specific country. For example: 'nordvpn connect us'
Provide a <city> argument to connect to a specific city. For example: 'nordvpn connect Hungary Budapest'
Provide a <group> argument to connect to a specific servers group. For example: 'nordvpn connect Onion_Over_VPN'

Press the Tab key to see auto-suggestions for countries and cities.`
)

type trustedPassTokenData struct {
	token    string
	owner_id string
}

func (c *cmd) getTrustedPassTokenData() (trustedPassTokenData, error) {
	resp, err := c.client.TokenInfo(context.Background(), &pb.Empty{})
	if err != nil {
		return trustedPassTokenData{}, err
	}

	if resp.TrustedPassOwnerId == "" || resp.TrustedPassToken == "" {
		return trustedPassTokenData{}, fmt.Errorf("invalid trusted pass token")
	}

	return trustedPassTokenData{token: resp.TrustedPassToken, owner_id: resp.TrustedPassOwnerId}, nil
}

func (c *cmd) injectLinkIntoMessage(url string, trustedPassURL string, message string) string {
	link := url
	tokenData, err := c.getTrustedPassTokenData()
	if err == nil {
		link = fmt.Sprintf(trustedPassURL, tokenData.token, tokenData.owner_id)
	}
	return fmt.Sprintf(message, link)
}

func (c *cmd) Connect(ctx *cli.Context) error {
	args := ctx.Args()

	// handling the case where options are provided in incorrect order or
	// atleast not the order the library expects them to be in
	if args.Len() > 1 && strings.HasPrefix(args.First(), "-") {
		if err := cli.ShowAppHelp(ctx); err != nil {
			return err
		}
		// the exact error message returned by the lib, when incorrect flag
		// is used, but in correct order
		return fmt.Errorf("Unrecognized flag: '%s'. Run '%s --help' to view available options.", args.Get(1), CommandFullName(ctx, os.Args))
	}

	serverTag, serverGroup, err := parseConnectArgs(ctx)
	if err != nil {
		return formatError(err)
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer close(ch)

	canceled := false
	go func(ch chan os.Signal) {
		for range ch {
			canceled = true
			_, _ = c.client.ConnectCancel(context.Background(), &pb.Empty{})
		}
	}(ch)

	connectCtx := context.Background()
	if iv := uievent.ItemValueFromServerGroupString(serverGroup); iv != pb.UIEvent_ITEM_VALUE_UNSPECIFIED {
		uiCtx := uievent.NewClickContext(pb.UIEvent_CLI, pb.UIEvent_CONNECT)
		uiCtx.ItemValue = iv
		connectCtx = uievent.AttachToOutgoingContext(connectCtx, uiCtx)
	}

	resp, err := c.client.Connect(connectCtx, &pb.ConnectRequest{
		ServerTag:   serverTag,
		ServerGroup: serverGroup,
	})
	if err != nil {
		return formatError(err)
	}

	var rpcErr error
	for {
		out, err := resp.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			// No race condition here as `canceled` is always set before `cancel()`
			if !canceled {
				return formatError(err)
			}
		}
		switch out.Type {
		case internal.CodeFailure:
			rpcErr = errors.New(client.ConnectCantConnect)
		case internal.CodeExpiredRenewToken:
			color.Yellow(client.RelogRequest)
			if rpcErr = c.Login(ctx); rpcErr != nil {
				break
			}
			rpcErr = c.Connect(ctx)
		case internal.CodeTokenRenewError:
			rpcErr = errors.New(client.AccountTokenRenewError)
		case internal.CodeExpiredAccessToken:
			fallthrough
		case internal.CodeRevokedAccessToken:
			return formatError(errors.New(client.AccessTokenExpired))
		case internal.CodeAccountExpired:
			rpcErr = errors.New(c.injectLinkIntoMessage(client.SubscriptionURL, client.SubscriptionURLLogin, ExpiredAccountMessage))
		case internal.CodeDedicatedIPRenewError:
			rpcErr = errors.New(c.injectLinkIntoMessage(client.SubscriptionDedicatedIPURL, client.SubscriptionDedicatedIPURLLogin, NoDedicatedIPMessage))
		case internal.CodeDedicatedIPNoServer:
			rpcErr = errors.New(NoDedidcatedIPServerMessage)
		case internal.CodeDedicatedIPServiceButNoServers:
			rpcErr = errors.New(NoPreferredDedicatedIPLocationSelected)
		case internal.CodeDisconnected:
			message := client.ConnectCanceled
			if len(out.Data) == 1 {
				message = client.ConnectCanceledNoHostname
			}
			color.Yellow(fmt.Sprintf(message, internal.StringsToInterfaces(out.Data)...))
		case internal.CodeTagNonexisting:
			rpcErr = errors.New(internal.TagNonexistentErrorMessage)
		case internal.CodeGroupNonexisting:
			rpcErr = errors.New(internal.GroupNonexistentErrorMessage)
		case internal.CodeServerUnavailable:
			rpcErr = errors.New(internal.ServerUnavailableErrorMessage)
		case internal.CodeVirtualLocationDisabled:
			rpcErr = errors.New(internal.SpecifiedServerIsVirtualLocation)
		case internal.CodeDoubleGroupError:
			rpcErr = errors.New(internal.DoubleGroupErrorMessage)
		case internal.CodeTechnologyDisabled:
			rpcErr = errors.New(TechnologyDisabledMessage)
		case internal.CodeDedicatedServersRenewError:
			rpcErr = errors.New(c.injectLinkIntoMessage(client.DedicatedServersUpselURL, client.DedicatedServersUpselURLLogin, DedicatedServersNoServiceMessage))
		case internal.CodeDedicatedServersServiceButNoServers:
			rpcErr = errors.New(c.injectLinkIntoMessage(client.DedicatedServersSetupURL, client.DedicatedServersSetupURLLogin, DedicatedServersNoServersAvailable))
		case internal.CodeDedicatedServersNotReady:
			rpcErr = errors.New(DedicatedServersServerNotReadyMessage)
		case internal.CodeDedicatedServersNoNordlynx:
			rpcErr = errors.New(DedicatedServersNoNordlynxMessage)
		case internal.CodeVPNRunning:
			color.Yellow(client.ConnectConnected)
		case internal.CodeNothingToDo:
			color.Yellow(client.ConnectConnecting)
		case internal.CodeUFWDisabled:
			color.Yellow(client.UFWDisabledMessage)
		case internal.CodeConnecting:
			message := client.ConnectStart
			if len(out.Data) == 1 {
				message = client.ConnectStartNoHostname
			}
			color.Green(fmt.Sprintf(message, internal.StringsToInterfaces(out.Data)...))
		case internal.CodeConnected:
			message := internal.ConnectSuccess
			if len(out.Data) == 1 {
				message = internal.ConnectSuccessNoHostname
			}
			color.Green(fmt.Sprintf(message, internal.StringsToInterfaces(out.Data)...))
		}
	}

	return formatError(rpcErr)
}

func (c *cmd) ConnectAutoComplete(ctx *cli.Context) {
	args := ctx.Args()
	groupName, hasGroupFlag := getFlagValue(flagGroup, ctx)
	c.printServersForAutoComplete(args.First(), hasGroupFlag, groupName)
}
