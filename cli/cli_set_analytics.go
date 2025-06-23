package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetAnalyticsUsageText is shown next to analytics command by nordvpn set --help

const SetAnalyticsUsageText = "Help us improve by sending anonymous " +
	"aggregate data: crash reports, OS version, marketing " +
	"performance, and feature usage data – nothing that could " +
	"identify you."

func (c *cmd) setAnalyticsFlag(flag bool) error {
	resp, err := c.client.SetAnalytics(context.Background(), &pb.SetGenericRequest{Enabled: flag})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		color.Yellow(fmt.Sprintf(MsgAlreadySet, "Analytics", nstrings.GetBoolLabel(flag)))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(MsgSetSuccess, "Analytics", nstrings.GetBoolLabel(flag)))
	}

	return nil
}

func (c *cmd) setAnalyticsFlow() error {
	fmt.Printf(MsgConsentAgreement)
	flag := readForConfirmationBlockUntilValid(os.Stdin, MsgConsentAgreementPrompt)

	return c.setAnalyticsFlag(flag)
}

// SetAnalytics
func (c *cmd) SetAnalytics(ctx *cli.Context) error {
	if ctx.NArg() > 1 {
		return formatError(argsCountError(ctx))
	}

	if ctx.NArg() == 1 {
		flag, err := nstrings.BoolFromString(ctx.Args().First())
		if err != nil {
			return formatError(argsParseError(ctx))
		}
		return c.setAnalyticsFlag(flag)
	}

	return c.setAnalyticsFlow()
}
