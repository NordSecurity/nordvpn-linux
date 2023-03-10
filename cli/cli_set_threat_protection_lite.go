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

// SetThreatProtectionLiteUsageText is shown next to threatprotectionlite command by nordvpn set --help
const SetThreatProtectionLiteUsageText = "Enables or disables ThreatProtectionLite. When enabled, the ThreatProtectionLite feature will automatically block suspicious websites so that no malware or other cyber threats can infect your device. Additionally, no flashy ads will come into your sight. More information on how it works: https://nordvpn.com/features/threat-protection/."

// SetThreatProtectionLiteArgsUsageText is shown by nordvpn set threatprotectionlite --help
const SetThreatProtectionLiteArgsUsageText = `[enabled]/[disabled]

Use this command to enable or disable ThreatProtectionLite. When enabled, the ThreatProtectionLite feature will automatically block suspicious websites so that no malware or other cyber threats can infect your device. Additionally, no flashy ads will come into your sight. More information on how it works: https://nordvpn.com/lt/features/threat-protection/

Supported values for [disabled]: 0, false, disable, off, disabled
Example: nordvpn set threatprotectionlite off

Supported values for [enabled]: 1, true, enable, on, enabled
Example: nordvpn set threatprotectionlite on

Notes:
  Setting ThreatProtectionLite disables user defined DNS servers`

func (c *cmd) SetThreatProtectionLite(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	if c.config.ThreatProtectionLite == flag {
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Threat Protection Lite", nstrings.GetBoolLabel(flag)))
		return nil
	}

	var dns = c.config.DNS
	if flag && len(c.config.DNS) > 0 {
		dns = nil
	}

	resp, err := c.client.SetThreatProtectionLite(
		context.Background(),
		&pb.SetThreatProtectionLiteRequest{
			ThreatProtectionLite: flag,
			Dns:                  dns,
		})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeFailure, internal.CodeVPNMisconfig:
		return formatError(internal.ErrUnhandled)
	case internal.CodeSuccess:
		c.config.ThreatProtectionLite = flag
		if flag && len(c.config.DNS) > 0 {
			color.Yellow(SetThreatProtectionLiteDisableDNS)
			c.config.DNS = dns
		}
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		color.Green(fmt.Sprintf(MsgSetSuccess, "Threat Protection Lite", nstrings.GetBoolLabel(flag)))
	}
	return nil
}
