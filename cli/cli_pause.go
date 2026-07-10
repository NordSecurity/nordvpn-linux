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

func pauseArgToDuration(arg string) (uint32, error) {
	switch arg {
	case "5m":
		return 300, nil
	case "15m":
		return 900, nil
	case "30m":
		return 1800, nil
	case "1h":
		return 3600, nil
	case "24h":
		return 86400, nil
	default:
		return 0, fmt.Errorf("unrecognized duration")
	}
}

func (c *cmd) Pause(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(errors.New(PauseNoArgsText))
	}

	pauseDuration, err := pauseArgToDuration(args.First())
	if err != nil {
		return formatError(errors.New(PauseNoArgsText))
	}

	resp, err := c.client.PauseConnection(context.Background(), &pb.PauseRequest{Seconds: pauseDuration})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeNothingToDo:
		return formatError(errors.New(PauseNothingToDo))
	case internal.CodePauseAttemptWhenConnectedToMeshPeer:
		return formatError(errors.New(PauseWhenMeshnetOn))
	case internal.CodeFailure:
		return formatError(errors.New(internal.UnhandledMessage))
	case internal.CodeSuccess:
		color.Green(fmt.Sprintf(PauseSuccess, args.First()))
	}

	return nil
}
