package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	SetFirewallUsageText     = "Enables or disables use of the firewall."
	SetFirewallMarkUsageText = "Traffic control filter used in " +
		"policy-based routing. It allows classifying packets " +
		"based on a previously set fwmark by iptables."
)

func (c *cmd) SetFirewall(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetFirewall(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Firewall", nstrings.GetBoolLabel(flag)))
	case internal.CodeDependencyError:
		color.Yellow(fmt.Sprintf(MsgInUse, "Firewall", "killswitch"))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Firewall", nstrings.GetBoolLabel(flag)))
	}
	return nil
}

func (c *cmd) SetFirewallMark(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	args := ctx.Args()
	mark, err := strconv.ParseUint(strings.TrimLeft(args.First(), "0x"), 16, 64)
	if err != nil {
		return formatError(err)
	}

	resp, err := c.client.SetFirewallMark(context.Background(), &pb.SetUint32Request{Value: uint32(mark)})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Firewall Mark", args.First()))
	case internal.CodeSuccess:
		color.Yellow("Restart daemon (e.g. `sudo systemctl restart nordvpnd` on systemd distros) for this setting to take an effect.")
		color.Green(fmt.Sprintf(MsgSetSuccess, "Firewall Mark", args.First()))
	}
	return nil
}
