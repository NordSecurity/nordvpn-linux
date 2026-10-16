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
		fmt.Print(RateNoArgsMessage)
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

	// #nosec G104 -- fire-and-forget analytics
	c.client.ReportUIEvent(context.Background(), &pb.UIEvent{
		FormReference: pb.UIEvent_REF_DISCONNECT,
		ItemName:      pb.UIEvent_RATE_CONNECTION,
		ItemType:      pb.UIEvent_BUTTON,
		ItemValue:     ratingToUIEventValue(rating),
	})

	color.Green(RateSuccess)
	return nil
}

func ratingToUIEventValue(rating int64) pb.UIEvent_ItemValue {
	switch rating {
	case 1:
		return pb.UIEvent_CSAT_SCORE_1
	case 2:
		return pb.UIEvent_CSAT_SCORE_2
	case 3:
		return pb.UIEvent_CSAT_SCORE_3
	case 4:
		return pb.UIEvent_CSAT_SCORE_4
	case 5:
		return pb.UIEvent_CSAT_SCORE_5
	default:
		return pb.UIEvent_ITEM_VALUE_UNSPECIFIED
	}
}

func (c *cmd) RateAutoComplete(ctx *cli.Context) {
	for rate := 1; rate <= 5; rate++ {
		fmt.Println(rate)
	}
}
