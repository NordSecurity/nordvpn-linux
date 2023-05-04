package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
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

func (c *cmd) SetDNS(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() == 0 {
		return formatError(argsCountError(ctx))
	}

	// check if first arg is false
	var dns []string
	var threatProtectionLite bool
	if args.Len() == 1 && nstrings.CanParseFalseFromString(args.First()) {
		if len(c.config.DNS) == 0 {
			return formatError(errors.New(color.YellowString(fmt.Sprintf(MsgAlreadySet, "DNS", nstrings.GetBoolLabel(false)))))
		}
		dns = nil
	} else {
		// cannot set more than three dns
		if args.Len() > 3 {
			return formatError(argsParseError(ctx))
		}
		// check validity
		for _, arg := range args.Slice() {
			if ip := net.ParseIP(arg); ip == nil {
				// TODO: use multierror when Go 1.20 comes out
				return formatError(argsParseError(ctx))
			}
		}
		// check equality
		argsSlice := args.Slice()
		sort.Strings(c.config.DNS)
		sort.Strings(argsSlice)
		if slices.Equal(c.config.DNS, argsSlice) {
			color.Yellow(fmt.Sprintf(MsgAlreadySet, "DNS", strings.Join(argsSlice, ", ")))
			return nil
		}
		dns = argsSlice
		if c.config.ThreatProtectionLite {
			threatProtectionLite = false
		}
	}

	resp, err := c.client.SetDNS(context.Background(), &pb.SetDNSRequest{
		Dns:                  dns,
		ThreatProtectionLite: threatProtectionLite,
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
		c.config.DNS = dns
		if c.config.ThreatProtectionLite {
			color.Yellow(SetThreatProtectionLiteDisableDNS)
			c.config.ThreatProtectionLite = threatProtectionLite
		}
		err = c.configManager.Save(c.config)
		if err != nil {
			return formatError(ErrConfig)
		}
		if len(c.config.DNS) == 0 {
			color.Green(fmt.Sprintf(MsgSetSuccess, "DNS", nstrings.GetBoolLabel(false)))
		} else {
			color.Green(fmt.Sprintf(MsgSetSuccess, "DNS", strings.Join(c.config.DNS, ", ")))
		}
	}
	return nil
}
