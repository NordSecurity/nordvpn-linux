package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/urfave/cli/v2"
)

func (c *cmd) Countries(ctx *cli.Context) error {
	resp, err := c.client.Countries(context.Background(), &pb.Empty{})
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		return formatError(err)
	}

	if resp.Type != internal.CodeSuccess {
		err := fmt.Errorf(MsgListIsEmpty, "countries")
		log.Println(internal.ErrorPrefix, err)
		return formatError(err)
	}

	footer := footerForServerGroupsList(resp.Servers)
	countryList, err := columns(resp.Servers,
		serverNameLen,
		formatServerName,
		footer,
	)
	if err != nil {
		log.Println(internal.ErrorPrefix, err)
		countries, _ := formatTable(resp.Servers, serverNameLen, formatServerName, 1, footer)
		fmt.Println(countries)
	} else {
		fmt.Println(countryList)
	}
	return nil
}
