package cli

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/urfave/cli/v2"
)

// GroupsUsageText is shown next to groups command by nordvpn --help
const GroupsUsageText = "Shows a list of available server groups"

func (c *cmd) Groups(ctx *cli.Context) error {
	resp, err := c.client.Groups(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	if resp.Type != internal.CodeSuccess {
		return formatError(fmt.Errorf(MsgListIsEmpty, "server groups"))
	}

	groupList, err := columns(resp.Data)

	if err != nil {
		log.Println(err)
		fmt.Println(strings.Join(resp.Data, ", "))
	} else {
		fmt.Println(groupList)
	}
	return nil
}
