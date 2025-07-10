package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Set autoconnect help text
const (
	SetAutoconnectUsageText     = "Enables or disables auto-connect. When enabled, this feature will automatically try to connect to VPN on operating system startup."
	SetAutoConnectArgsUsageText = `<enabled>|<disabled> [<country>|<server>|<country_code>|<city>|<group>|<country> <city>]`
	SetAutoConnectDescription   = `Enables or disables auto-connect. When enabled, this feature will automatically try to connect to VPN on operating system startup.

Supported values for <disabled>: 0, false, disable, off, disabled
Example: nordvpn set autoconnect off

Supported values for <enabled>: 1, true, enable, on, enabled
Example: nordvpn set autoconnect on

Provide a <country> argument to connect to a specific country. For example: 'nordvpn set autoconnect enabled Australia'
Provide a <server> argument to connect to a specific server. For example: 'nordvpn set autoconnect enabled jp35'
Provide a <country_code> argument to connect to a specific country. For example: 'nordvpn set autoconnect enabled us'
Provide a <city> argument to connect to a specific city. For example: 'nordvpn set autoconnect enabled Budapest'
Provide a <group> argument to connect to a specific servers group. For example: 'nordvpn set autoconnect enabled Onion_Over_VPN'`
)

func (c *cmd) SetAutoConnect(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() == 0 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(args.First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	serverTag, serverGroup, err := parseConnectArgs(ctx)
	if err != nil {
		return formatError(err)
	}

	resp, err := c.client.SetAutoConnect(context.Background(), &pb.SetAutoconnectRequest{
		Enabled:     flag,
		ServerTag:   serverTag,
		ServerGroup: serverGroup,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure, internal.CodeEmptyPayloadError:
		return formatError(fmt.Errorf(client.ConnectCantConnect))
	case internal.CodeAutoConnectServerNotObfuscated:
		return formatError(errors.New(AutoConnectOnNonObfuscatedServerObfuscateOn))
	case internal.CodeAutoConnectServerObfuscated:
		return formatError(errors.New(AutoConnectOnObfuscatedServerObfuscateOff))
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Auto-connect", nstrings.GetBoolLabel(flag)))
	case internal.CodeExpiredRenewToken:
		color.Yellow(client.RelogRequest)
		err = c.Login(ctx)
		if err != nil {
			return err
		}
		return c.SetAutoConnect(ctx)
	case internal.CodeTokenRenewError:
		return formatError(errors.New(client.AccountTokenRenewError))
	case internal.CodeExpiredAccessToken:
		fallthrough
	case internal.CodeRevokedAccessToken:
		return formatError(errors.New(client.AccessTokenExpired))
	case internal.CodeDedicatedIPRenewError:
		link := client.SubscriptionDedicatedIPURL
		tokenData, err := c.getTrustedPassTokenData()
		if err == nil {
			link = fmt.Sprintf(client.SubscriptionDedicatedIPURLLogin, tokenData.token, tokenData.owner_id)
		}
		return formatError(fmt.Errorf(NoDedicatedIPMessage, link))
	case internal.CodeDedicatedIPNoServer:
		return formatError(errors.New(NoDedidcatedIPServerMessage))
	case internal.CodeDedicatedIPServiceButNoServers:
		return formatError(errors.New(NoPreferredDedicatedIPLocationSelected))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Auto-connect", nstrings.GetBoolLabel(flag)))
	}
	return nil
}

func (c *cmd) SetAutoConnectAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 0:
		booleans := nstrings.GetBools()
		sort.Strings(booleans)

		for _, v := range booleans {
			fmt.Println(v)
		}
	default:
		args := ctx.Args()
		if args.Len() > 0 {
			//check first arg
			if !nstrings.CanParseTrueFromString(args.First()) {
				return
			}

			groupName, hasGroupFlag := getFlagValue(flagGroup, ctx)

			if !hasGroupFlag && strings.HasPrefix(args.Get(args.Len()-1), "-") {
				// if the group flag is not set, but the last argument starts with "-" then give as suggestions --group
				fmt.Println("--" + flagGroup)
				return
			}
			c.printServersForAutoComplete(args.Get(1), hasGroupFlag, groupName)
		}
	}
}
