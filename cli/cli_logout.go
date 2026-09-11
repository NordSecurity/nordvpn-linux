package cli

import (
	"context"
	"errors"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	flagRevokeToken = "revoke-token"
)

func (c *cmd) Logout(ctx *cli.Context) error {
	// #nosec G104 -- fire-and-forget analytics
	c.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_CLI,
		ItemName:      pb.UIEvent_LOGOUT,
		ItemType:      pb.UIEvent_CLICK,
	})

	revokeToken := ctx.IsSet(flagRevokeToken)

	payload, err := c.client.Logout(context.Background(), &pb.LogoutRequest{
		RevokeToken: revokeToken,
	})

	if err != nil {
		return formatError(err)
	}

	switch payload.Type {
	case internal.CodeSuccess:
		color.Green(LogoutSuccess)
		return nil
	case internal.CodeTokenStillValid:
		color.Green(LogoutTokenSuccess)
		return nil
	case internal.CodeTokenInvalid:
		color.Green(LogoutTokenAlreadyInvalid)
		return nil
	case internal.CodeRevokedAccessToken:
		color.Green(LogoutRevokeTokenSuccess)
		return nil
	default:
		return formatError(errors.New(CheckYourInternetConnMessage))
	}
}
