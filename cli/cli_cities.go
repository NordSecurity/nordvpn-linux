package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/urfave/cli/v2"
)

// CitiesUsageText is shown next to cities command by nordvpn --help
const CitiesUsageText = "Shows a list of cities where servers are available"

// CitiesArgsUsageText is shown by nordvpn cities --help
const CitiesArgsUsageText = `[country]

Use this command to show cities where servers are available.

Example: 'nordvpn cities United_States'

Press the Tab key to see auto-suggestions for countries.`

func (c *cmd) Cities(ctx *cli.Context) error {
	args := ctx.Args()

	if args.Len() != 1 {
		return formatError(argsParseError(ctx))
	}

	resp, err := c.client.Cities(context.Background(), &pb.CitiesRequest{
		Protocol:  c.config.Protocol,
		Obfuscate: c.config.Obfuscate,
		Country:   args.First(),
	})
	if err != nil {
		return formatError(err)
	}

	if len(resp.Data) == 0 {
		return formatError(errors.New(CitiesNotFoundError))
	}

	formattedList, err := internal.Columns(resp.Data)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		fmt.Println(strings.Join(resp.Data, ", "))
	} else {
		fmt.Println(formattedList)
	}
	return nil
}

func (c *cmd) CitiesAutoComplete(ctx *cli.Context) {
	resp, err := c.client.Countries(context.Background(), &pb.CountriesRequest{
		Protocol:  c.config.Protocol,
		Obfuscate: c.config.Obfuscate,
	})
	if err != nil {
		return
	}

	for _, country := range resp.Data {
		fmt.Println(country)
	}
}
