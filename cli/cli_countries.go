package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/log"

	"github.com/urfave/cli/v2"
)

func (c *cmd) Countries(ctx *cli.Context) error {
	resp, err := c.client.Countries(context.Background(), &pb.Empty{})
	if err != nil {
		log.Error(err)
		return formatError(err)
	}

	if resp.Type != internal.CodeSuccess {
		err := fmt.Errorf(MsgListIsEmpty, "countries")
		log.Error(err)
		return formatError(err)
	}

	countryList, err := columns(resp.Servers,
		serverNameLen,
		formatServerName,
	)
	if err != nil {
		log.Error(err)
		countries, _ := formatTable(resp.Servers, serverNameLen, formatServerName, 1)
		fmt.Println(countries)
	} else {
		fmt.Println(countryList)
	}
	return nil
}
