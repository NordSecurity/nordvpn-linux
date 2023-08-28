package cli

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func SetLANDiscoveryErrorCodeToError(code pb.SetErrorCode, flag bool) error {
	switch code {
	case pb.SetErrorCode_FAILURE:
		return formatError(internal.ErrUnhandled)
	case pb.SetErrorCode_CONFIG_ERROR:
		return formatError(ErrConfig)
	case pb.SetErrorCode_ALREADY_SET:
		return errors.New(color.YellowString(fmt.Sprintf(SetLANDiscoveryAlreadyEnabled, nstrings.GetBoolLabel(flag))))
	}
	return nil
}

func SetLANDiscoveryStatusToMessage(code pb.SetLANDiscoveryStatus, flag bool) {
	switch code {
	case pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED_ALLOWLIST_RESET:
		color.Yellow(SetLANDiscoveryAllowlistReset)
		fallthrough
	case pb.SetLANDiscoveryStatus_DISCOVERY_CONFIGURED:
		color.Green(fmt.Sprintf(MsgSetSuccess, "LAN Discovery", nstrings.GetBoolLabel(flag)))
	}
}

func (c *cmd) SetLANDiscovery(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	arg := ctx.Args().First()
	isEnabled, err := nstrings.BoolFromString(arg)
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetLANDiscovery(context.Background(), &pb.SetLANDiscoveryRequest{
		Enabled: isEnabled,
	})

	switch resp.Response.(type) {
	case *pb.SetLANDiscoveryResponse_ErrorCode:
		return SetLANDiscoveryErrorCodeToError(resp.GetErrorCode(), isEnabled)
	case *pb.SetLANDiscoveryResponse_SetLanDiscoveryStatus:
		SetLANDiscoveryStatusToMessage(resp.GetSetLanDiscoveryStatus(), isEnabled)
	}

	if isEnabled {
		// Quick and a bit dirty fix to the issue where private subnets deleted by lan-discovery would
		// reappear. Ideally we would stop using CLI config at all, but that is too big of a task for now.
		for _, subnet := range c.config.Allowlist.Subnets.ToSlice() {
			subnetStr, ok := subnet.(string)
			if !ok {
				continue
			} else if _, subnetParsed, err := net.ParseCIDR(subnetStr); err != nil {
				continue
			} else if subnetParsed.IP.IsPrivate() || subnetParsed.IP.IsLinkLocalUnicast() {
				c.config.Allowlist.Subnets.Remove(subnet)
			}
		}
	}

	err = c.configManager.Save(c.config)
	if err != nil {
		return formatError(ErrConfig)
	}

	return nil
}
