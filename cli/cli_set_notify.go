package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	filesharepb "github.com/NordSecurity/nordvpn-linux/fileshare/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// SetNotifyUsageText is shown next to notify command by nordvpn set --help
const SetNotifyUsageText = "Enables or disables notifications"

func (c *cmd) SetNotify(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return formatError(argsCountError(ctx))
	}

	flag, err := nstrings.BoolFromString(ctx.Args().First())
	if err != nil {
		return formatError(argsParseError(ctx))
	}

	daemonResp, err := c.client.SetNotify(context.Background(), &pb.SetNotifyRequest{
		Uid:    int64(os.Getuid()),
		Notify: flag,
	})
	if err != nil {
		return formatError(err)
	}

	printMessage := func() {}
	defer func() {
		printMessage()
	}()

	messageNothingToSet := func() {
		color.Yellow(fmt.Sprintf(SetNotifyNothingToSet, nstrings.GetBoolLabel(flag)))
	}
	messageSuccess := func() {
		color.Green(fmt.Sprintf(SetNotifySuccess, nstrings.GetBoolLabel(flag)))
	}

	switch daemonResp.Type {
	case internal.CodeConfigError:
		return formatError(ErrConfig)
	case internal.CodeNothingToDo:
		printMessage = messageNothingToSet
	case internal.CodeSuccess:
		printMessage = messageSuccess
	}

	if c.IsFileshareDaemonReachable(ctx) != nil {
		return nil
	}

	fileshareDaemonResp, err := c.fileshareClient.SetNotifications(context.Background(),
		&filesharepb.SetNotificationsRequest{Enable: flag})

	if err != nil {
		return formatError(err)
	}

	// We configure notifications for main and fileshare daemon
	// if both notifications were configured successfully, report success to the user
	// if main daemon was already configured but fileshare daemon was not yet configured and vice versa, report success
	// if both daemons were already configured, report already configured
	switch fileshareDaemonResp.Status {
	case filesharepb.SetNotificationsStatus_NOTHING_TO_DO:
		return nil
	case filesharepb.SetNotificationsStatus_SET_SUCCESS:
		printMessage = messageSuccess
	case filesharepb.SetNotificationsStatus_SET_FAILURE:
		printMessage = func() {
			color.Red(internal.UnhandledMessage)
		}
	}

	return nil
}
