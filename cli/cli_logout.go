package cli

import (
	"context"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// LogoutUsageText is shown next to logout command by nordvpn --help
const (
	flagPersistToken = "persist-token"
)

func (c *cmd) Logout(ctx *cli.Context) error {
	// #nosec G104 -- fire-and-forget analytics
	c.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_CLI,
		ItemName:      pb.UIEvent_LOGOUT,
		ItemType:      pb.UIEvent_CLICK,
	})

	persistToken := ctx.IsSet(flagPersistToken)

	payload, err := c.client.Logout(context.Background(), &pb.LogoutRequest{
		PersistToken: persistToken,
	})

	if err != nil {
		return formatError(err)
	}

	switch payload.Type {
	case internal.CodeSuccess:
		color.Green(LogoutSuccess)
		return nil
	case internal.CodeTokenInvalidated:
		color.Green(LogoutTokenSuccess)
		return nil
	default:
		return formatError(errors.New(CheckYourInternetConnMessage))
	}
}
