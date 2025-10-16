package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// Rate help text
const (
	RateUsageText     = "Rates your last connection quality (1-5)"
	RateArgsUsageText = `[1-5]`
	RateDescription   = `Use this command to rate the connection quality of your previous session via scale of 1 (poor) through 5 (great).

Example: nordvpn rate 5

Notes:
  You can only rate a single connection once.`
)

func (c *cmd) Rate(ctx *cli.Context) error {
	var ratingInput string
	switch ctx.NArg() {
	case 0:
		fmt.Printf(RateNoArgsMessage)
		reader := bufio.NewReader(os.Stdin)
		var err error
		ratingInput, err = reader.ReadString('\n')
		if err != nil {
			return formatError(errors.New(internal.UnhandledMessage))
		}
	case 1:
		ratingInput = ctx.Args().First()
	default:
		return formatError(argsParseError(ctx))
	}
	ratingInput = strings.TrimLeft(strings.TrimSpace(ratingInput), "+")
	rating, err := strconv.ParseInt(ratingInput, 10, 0)
	if err != nil {
		return formatError(argsParseError(ctx))
	}
	if rating < 1 || rating > 5 {
		return formatError(argsParseError(ctx))
	}

	payload, err := c.client.RateConnection(context.Background(), &pb.RateRequest{Rating: rating})
	if err != nil {
		return formatError(err)
	}

	switch payload.Type {
	case internal.CodeEmptyPayloadError:
		return formatError(errors.New(RateNoConnectionMade))
	case internal.CodeNoNewDataError:
		return formatError(errors.New(RateAlreadyRated))
	case internal.CodeNothingToDo:
		color.Yellow(MsgNothingToRate)
	case internal.CodeSuccess:
		color.Green(RateSuccess)
	}

	return nil
}

func (c *cmd) RateAutoComplete(ctx *cli.Context) {
	for rate := 1; rate <= 5; rate++ {
		fmt.Println(rate)
	}
}
