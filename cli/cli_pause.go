package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var argToPauseDurationSeconds = map[string]uint32{"5m": 300, "15m": 900, "30m": 1800, "1h": 3600, "24h": 86400}

func pauseArgToDuration(arg string) (uint32, error) {
	if pauseDurationSeconds, ok := argToPauseDurationSeconds[arg]; ok {
		return pauseDurationSeconds, nil
	}
	return 0, fmt.Errorf("unrecognized duration")
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
		color.Green(PauseSuccess, args.First())
	}

	return nil
}

func PauseAutoComplete(ctx *cli.Context) {
	arg := ctx.Args().First()
	for duration := range argToPauseDurationSeconds {
		if strings.Contains(duration, arg) {
			fmt.Println(duration)
		}
	}
}
