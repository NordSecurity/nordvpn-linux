package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Set Threat Protection Lite help text
const (
	SetThreatProtectionLiteUsageText     = "Enables or disables ThreatProtectionLite. When enabled, the ThreatProtectionLite feature will automatically block suspicious websites so that no malware or other cyber threats can infect your device. Additionally, no flashy ads will come into your sight. More information on how it works: https://nordvpn.com/features/threat-protection/?utm_medium=app&utm_source=nordvpn-linux-cli&utm_campaign=settings-explore_threat_protection&nm=app&ns=nordvpn-linux-cli&nc=settings-explore_threat_protection."
	SetThreatProtectionLiteArgsUsageText = `<enabled>|<disabled>`
	SetThreatProtectionLiteDescription   = `Use this command to enable or disable ThreatProtectionLite. When enabled, the ThreatProtectionLite feature will automatically block suspicious websites so that no malware or other cyber threats can infect your device. Additionally, no flashy ads will come into your sight. More information on how it works: https://nordvpn.com/features/threat-protection/?utm_medium=app&utm_source=nordvpn-linux-cli&utm_campaign=settings-explore_threat_protection&nm=app&ns=nordvpn-linux-cli&nc=settings-explore_threat_protection

Supported values for <disabled>: 0, false, disable, off, disabled
Example: nordvpn set threatprotectionlite off

Supported values for <enabled>: 1, true, enable, on, enabled
Example: nordvpn set threatprotectionlite on

Notes:
  Setting ThreatProtectionLite disables user defined DNS servers`
)

func setTPLErrorCodeToError(code pb.SetErrorCode, args ...any) error {
	switch code {
	case pb.SetErrorCode_FAILURE:
		return formatError(internal.ErrUnhandled)
	case pb.SetErrorCode_CONFIG_ERROR:
		return formatError(ErrConfig)
	case pb.SetErrorCode_ALREADY_SET:
		color.Yellow(fmt.Sprintf(SetThreatProtectionLiteAlreadySet, args...))
		return nil
	}
	return nil
}

func (c *cmd) SetThreatProtectionLite(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.SetThreatProtectionLite(
		context.Background(),
		&pb.SetThreatProtectionLiteRequest{
			ThreatProtectionLite: flag,
		})
	if err != nil {
		return formatError(err)
	}

	switch resp.Response.(type) {
	case *pb.SetThreatProtectionLiteResponse_ErrorCode:
		return setTPLErrorCodeToError(resp.GetErrorCode(), nstrings.GetBoolLabel(flag))
	case *pb.SetThreatProtectionLiteResponse_SetThreatProtectionLiteStatus:
		if resp.GetSetThreatProtectionLiteStatus() == pb.SetThreatProtectionLiteStatus_TPL_CONFIGURED_DNS_RESET {
			color.Yellow(SetThreatProtectionLiteDisableDNS)
		}
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(MsgSetSuccess, "Threat Protection Lite", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
