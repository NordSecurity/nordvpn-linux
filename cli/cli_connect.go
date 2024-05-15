package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

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

func (c *cmd) browseToSubscriptionPage(url string, loginURL string) {
	resp, err := c.client.TokenInfo(context.Background(), &pb.Empty{})
	if err != nil {
		browse(url)
		return
	}

	browse(fmt.Sprintf(loginURL, resp.Token, resp.Id))
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
		return fmt.Errorf("flag provided but not defined: %s", args.Get(1))
	}

	// generate server tag from given args
	serverTag := strings.Join(args.Slice(), " ")
	serverTag = strings.ToLower(serverTag)
	serverGroup := ctx.String(flagGroup)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	defer close(ch)
	go func(ch chan os.Signal) {
		for range ch {
			// #nosec G104 -- LVPN-2090
			c.client.Disconnect(context.Background(), &pb.Empty{})
		}
	}(ch)

	resp, err := c.client.Connect(context.Background(), &pb.ConnectRequest{
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
			return formatError(err)
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
		case internal.CodeAccountExpired:
			c.browseToSubscriptionPage(client.SubscriptionURL, client.SubscriptionDedicatedIPURLLogin)
			rpcErr = ErrAccountExpired
		case internal.CodeDedicatedIPRenewError:
			// #nosec G104 -- the user gets URL in case of failure
			c.browseToSubscriptionPage(client.SubscriptionDedicatedIPURL, client.SubscriptionDedicatedIPURLLogin)
			rpcErr = ErrAccountExpired
		case internal.CodeDisconnected:
			rpcErr = errors.New(internal.DisconnectSuccess)
		case internal.CodeTagNonexisting:
			rpcErr = errors.New(internal.TagNonexistentErrorMessage)
		case internal.CodeGroupNonexisting:
			rpcErr = errors.New(internal.GroupNonexistentErrorMessage)
		case internal.CodeServerUnavailable:
			rpcErr = errors.New(internal.ServerUnavailableErrorMessage)
		case internal.CodeDoubleGroupError:
			rpcErr = errors.New(internal.DoubleGroupErrorMessage)
		case internal.CodeVPNRunning:
			color.Yellow(client.ConnectConnected)
		case internal.CodeUFWDisabled:
			color.Yellow(client.UFWDisabledMessage)
		case internal.CodeConnecting:
			color.Green(fmt.Sprintf(client.ConnectStart, internal.StringsToInterfaces(out.Data)...))
		case internal.CodeConnected:
			color.Green(fmt.Sprintf(internal.ConnectSuccess, internal.StringsToInterfaces(out.Data)...))
		}
	}

	return formatError(rpcErr)
}

func (c *cmd) ConnectAutoComplete(ctx *cli.Context) {
	args := ctx.Args()
	if args.Len() == 0 {
		resp, err := c.client.Groups(context.Background(), &pb.Empty{})
		if err != nil {
			return
		}
		groupList, err := internal.Columns(resp.Data)
		if err != nil {
			log.Println(err)
		}
		resp, err = c.client.Countries(context.Background(), &pb.Empty{})
		if err != nil {
			return
		}
		countryList, err := internal.Columns(resp.Data)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(countryList + " " + groupList)
	} else if args.Len() == 1 {
		resp, err := c.client.Cities(context.Background(), &pb.CitiesRequest{
			Country: ctx.Args().First(),
		})
		if err != nil {
			return
		}
		cityList, err := internal.Columns(resp.Data)
		if err != nil {
			log.Println(err)
		}
		fmt.Println(cityList)
	}
}
