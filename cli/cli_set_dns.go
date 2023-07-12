package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetDNSUsageText is shown next to dns command by nordvpn set --help
const SetDNSUsageText = "Sets custom DNS servers"

// SetDNSArgsUsageText is shown by nordvpn set dns --help
const SetDNSArgsUsageText = `<servers>|<disabled>

Use this command to set DNS servers.

Supported values for <disabled>: 0, false, disable, off, disabled
Example: nordvpn set dns off

Arguments <servers> is a list of IP addresses separated by space
Example: nordvpn set dns 0.0.0.0 1.2.3.4

Limits:
  Can set up to 3 DNS servers

Notes:
  Setting DNS disables ThreatProtectionLite`

func setDNSCommonErrorCodeToError(code pb.SetErrorCode, args ...any) error {
	switch code {
	case pb.SetErrorCode_FAILURE:
		return formatError(internal.ErrUnhandled)
	case pb.SetErrorCode_CONFIG_ERROR:
		return formatError(ErrConfig)
	case pb.SetErrorCode_ALREADY_SET:
		return errors.New(color.YellowString(fmt.Sprintf(SetDNSAlreadySet, args...)))
	}
	return nil
}

func handleSetDNSStatus(code pb.SetDNSStatus, dns []string) error {
	switch code {
	case pb.SetDNSStatus_INVALID_DNS_ADDRESS:
		return fmt.Errorf(SetDNSInvalidAddress)
	case pb.SetDNSStatus_TOO_MANY_VALUES:
		return fmt.Errorf(SetDNSTooManyValues)
	case pb.SetDNSStatus_DNS_CONFIGURED_TPL_RESET:
		color.Yellow(SetDNSDisableThreatProtectionLite)
		fallthrough
	case pb.SetDNSStatus_DNS_CONFIGURED:
		if dns == nil {
			color.Green(fmt.Sprintf(MsgSetSuccess, "DNS", nstrings.GetBoolLabel(false)))
		} else {
			color.Green(fmt.Sprintf(MsgSetSuccess, "DNS", strings.Join(dns, ", ")))
		}
	}
	return nil
}

func (c *cmd) SetDNS(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() == 0 {
		return formatError(argsCountError(ctx))
	}

	// check if first arg is false
	var dns []string
	if args.Len() == 1 && nstrings.CanParseFalseFromString(args.First()) {
		dns = nil
	} else {
		dns = args.Slice()
	}

	resp, err := c.client.SetDNS(context.Background(), &pb.SetDNSRequest{
		Dns: dns,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Response.(type) {
	case *pb.SetDNSResponse_ErrorCode:
		if dns == nil {
			return setDNSCommonErrorCodeToError(resp.GetErrorCode(), nstrings.GetBoolLabel(false))
		} else {
			return setDNSCommonErrorCodeToError(resp.GetErrorCode(), strings.Join(dns, ", "))
		}
	case *pb.SetDNSResponse_SetDnsStatus:
		return handleSetDNSStatus(resp.GetSetDnsStatus(), dns)
	}
	return nil
}
