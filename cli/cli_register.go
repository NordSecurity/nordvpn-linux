package cli

import (
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"

	"github.com/urfave/cli/v2"
)

// RegisterUsageText is shown next to register command by nordvpn --help
const RegisterUsageText = "Registers a new user account"

func (c *cmd) Register(ctx *cli.Context) error {
	return c.login(pb.LoginType_LoginType_SIGNUP)
}
