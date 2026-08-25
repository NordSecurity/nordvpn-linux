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
	SetRealTimeProtectionUsageText     = "Turns real-time protection on or off. Real-time protection blocks scam and phishing attempts and reduces intrusive ads while you’re connected to the VPN. Learn more about how it works: " + realTimeProtectionLearnMoreUrl + "."
	SetRealTimeProtectionArgsUsageText = `<enabled>|<disabled>`
	SetRealTimeProtectionDescription   = `Use this command to enable or disable Real-time protection. When enabled, the Real-time protection feature will automatically block suspicious websites so that no malware or other cyber threats can infect your device. Additionally, no flashy ads will come into your sight. More information on how it works: ` + realTimeProtectionLearnMoreUrl + "\n\n" + realTimeProtectionExamples

	realTimeProtectionLearnMoreUrl = "https://nordvpn.com/features/threat-protection/?utm_medium=app&utm_source=nordvpn-linux-cli&utm_campaign=settings-explore_threat_protection&nm=app&ns=nordvpn-linux-cli&nc=settings-explore_threat_protection"
	realTimeProtectionExamples     = `Supported values for <disabled>: 0, false, disable, off, disabled
Example: nordvpn set protection off

Supported values for <enabled>: 1, true, enable, on, enabled
Example: nordvpn set protection on

Notes:
  Real-time protection isn’t compatible with custom DNS. Activating one turns the other off.`
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
