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

// Set Real-time protection help text
const (
	SetRealTimeProtectionUsageText     = "Use this command to turn real-time protection on or off. Real-time protection blocks scam and phishing attempts and reduces intrusive ads while you’re connected to the VPN. Learn more about how it works: " + realTimeProtectionLearnMoreUrl + "."
	SetRealTimeProtectionArgsUsageText = `<enabled>|<disabled>`
	SetRealTimeProtectionDescription   = `Use this command to turn real-time protection on or off. Real-time protection blocks scam and phishing attempts and reduces intrusive ads while you’re connected to the VPN. Learn more about how it works: ` + realTimeProtectionLearnMoreUrl + "\n\n" + realTimeProtectionExamples

	realTimeProtectionLearnMoreUrl = "https://nordvpn.com/features/threat-protection/?utm_medium=app&utm_source=nordvpn-linux-cli&utm_campaign=settings-explore_threat_protection&nm=app&ns=nordvpn-linux-cli&nc=settings-explore_threat_protection"
	realTimeProtectionExamples     = `Supported values for <disabled>: 0, false, disable, off, disabled

Supported values for <on>: 1, true, enable, on, enabled
Example: nordvpn set protection on
Supported values for <off>: 0, false, disable, off, disabled
Example: nordvpn set protection off

Note: Real-time protection works with our default DNS servers only. If custom DNS is active, we’ll turn it off once you turn on real-time protection.`
)

func setTPLErrorCodeToError(code pb.SetErrorCode, args ...any) error {
	switch code {
	case pb.SetErrorCode_FAILURE:
		return formatError(internal.ErrUnhandled)
	case pb.SetErrorCode_CONFIG_ERROR:
		return formatError(ErrConfig)
	case pb.SetErrorCode_ALREADY_SET:
		color.Yellow(fmt.Sprintf(SetRealTimeProtectionAlreadySet, args...))
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
		color.Green(fmt.Sprintf(MsgSetSuccess, "Real-time protection", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
