package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func parsePauseArg(arg string) (uint32, error) {
	unit := arg[len(arg)-1:]
	value, err := strconv.Atoi(arg[:len(arg)-1])
	if err != nil {
		return 0, fmt.Errorf("parsing arguments: %w", err)
	}

	const minutesUnit = "m"
	const hoursUnit = "h"

	if unit == minutesUnit {
		const secondsInMinute = 60
		return uint32(value * secondsInMinute), nil
	}

	if unit == hoursUnit {
		const secondsInHour = 3600
		return uint32(value * secondsInHour), nil
	}

	return 0, fmt.Errorf("unrecognized unit")
}

func (c *cmd) Pause(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsCountError(ctx))
	}

	pauseDuration, err := parsePauseArg(args.First())
	if err != nil {
		return formatError(errors.New(ArgumentParsingError))
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
		color.Green(PauseSuccess)
	}

	return nil
}
