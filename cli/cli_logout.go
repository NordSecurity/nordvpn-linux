package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// LogoutUsageText is shown next to logout command by nordvpn --help
const (
	LogoutUsageText  = "Logs you out"
	flagPersistToken = "persist-token"
)

func (c *cmd) Logout(ctx *cli.Context) error {
	persistToken := ctx.IsSet(flagPersistToken)

	payload, err := c.client.Logout(context.Background(), &pb.LogoutRequest{
		PersistToken: persistToken,
	})

	if err != nil {
		return formatError(err)
	}

	if payload.Type != internal.CodeSuccess {
		return formatError(errors.New(CheckYourInternetConnMessage))
	}
	color.Green(fmt.Sprintf(LogoutSuccess))
	return nil
}
