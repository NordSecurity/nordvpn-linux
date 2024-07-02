package cli

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/urfave/cli/v2"
)

// Cities help text
const (
	CitiesUsageText     = "Shows a list of cities where servers are available"
	CitiesArgsUsageText = `<country>`
)

var CitiesDescription = fmt.Sprintf(MsgShowListOfServers, "cities") + "\n\nExample: 'nordvpn cities United_States'\n\nPress the Tab key to see auto-suggestions for countries.'"

func (c *cmd) Cities(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.Cities(context.Background(), &pb.CitiesRequest{
		Country: args.First(),
	})
	if err != nil {
		return formatError(err)
	}

	if resp.Type != internal.CodeSuccess {
		err := fmt.Errorf(MsgListIsEmpty, "cities")
		log.Println(internal.ErrorPrefix, err)
		return formatError(err)
	}

	if len(resp.Servers) == 0 {
		return formatError(errors.New(CitiesNotFoundError))
	}

	footer := footerForServerGroupsList(resp.Servers)
	formattedList, err := columns(resp.Servers,
		serverNameLen,
		formatServerName,
		footer,
	)
	if err == nil {
		fmt.Println(formattedList)
	} else {
		log.Println(internal.ErrorPrefix, err)

		columns, _ := formatTable(resp.Servers, serverNameLen, formatServerName, 1, footer)
		fmt.Println(columns)
	}
	return nil
}

func (c *cmd) CitiesAutoComplete(ctx *cli.Context) {
	resp, err := c.client.Countries(context.Background(), &pb.Empty{})
	if err != nil {
		return
	}

	for _, server := range resp.Servers {
		fmt.Println(server.Name)
	}
}
