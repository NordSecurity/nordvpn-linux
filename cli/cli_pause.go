package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/pause"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

var argToPauseDurationSeconds = map[string]pb.PauseInverval{
	"5m":  pb.PauseInverval_PAUSE_5_MIN,
	"15m": pb.PauseInverval_PAUSE_15_MIN,
	"30m": pb.PauseInverval_PAUSE_30_MIN,
	"1h":  pb.PauseInverval_PAUSE_1_HOUR,
	"24h": pb.PauseInverval_PAUSE_24_HOURS}

func pauseArgToInterval(arg string) (pb.PauseInverval, error) {
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

	pauseInterval, err := pauseArgToInterval(args.First())
	if err != nil {
		return formatError(errors.New(PauseNoArgsText))
	}

	// #nosec G104 -- fire-and-forget analytics
	c.client.ReportUIEvent(context.Background(),
		&pb.UIEvent{
			FormReference: pb.UIEvent_CLI,
			ItemName:      pb.UIEvent_PAUSE,
			ItemType:      pb.UIEvent_CLICK,
			ItemValue:     pause.PauseIntervalToPauseUIEventItemValue(pauseInterval)})
	resp, err := c.client.PauseConnection(context.Background(), &pb.PauseRequest{Interval: pauseInterval})
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
