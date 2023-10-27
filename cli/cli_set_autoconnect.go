package cli

import (
	"context"
	"errors"
	"fmt"
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

	// generate server tag from given args
	var serverTag string
	if args.Len() > 1 {
		serverTag = strings.Join(args.Slice()[1:], "")
		serverTag = strings.Trim(serverTag, " ")
		serverTag = strings.ToLower(serverTag)
	}

	settings, err := c.getSettings()
	if err != nil {
		return formatError(err)
	}
	allowlist := settings.GetAllowlist()

	resp, err := c.client.SetAutoConnect(context.Background(), &pb.SetAutoconnectRequest{
		ServerTag:   serverTag,
		Obfuscate:   c.config.Obfuscate,
		AutoConnect: flag,
		Allowlist:   allowlist,
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
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Auto-connect", nstrings.GetBoolLabel(flag)))
	}
	return nil
}

func (c *cmd) SetAutoConnectAutoComplete(ctx *cli.Context) {
	switch ctx.NArg() {
	case 0:
		for _, v := range nstrings.GetBools() {
			fmt.Println(v)
		}
	default:
		if ctx.NArg() > 0 {
			//check first arg
			if !nstrings.CanParseTrueFromString(ctx.Args().First()) {
				return
			}

			// create config after auth
			args := ctx.Args()
			resp, err := func(args []string) (*pb.Payload, error) {
				switch len(args) {
				case 1:
					return c.client.Countries(context.Background(), &pb.CountriesRequest{
						Obfuscate: c.config.Obfuscate,
					})
				case 2:
					return c.client.Cities(context.Background(), &pb.CitiesRequest{
						Obfuscate: c.config.Obfuscate,
						Country:   args[1],
					})
				}
				return nil, errors.New("bad args")
			}(args.Slice())
			if err != nil {
				return
			}

			for _, item := range resp.Data {
				fmt.Println(item)
			}
		}
	}
}
