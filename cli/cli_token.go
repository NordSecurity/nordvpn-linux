package cli

import (
	"context"
	"fmt"

	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/urfave/cli/v2"
)

const (
	// TokenUsageText is shown next to login command by nordvpn --help
	TokenUsageText = "Show token information" // #nosec
)

func (c *cmd) TokenInfo(ctx *cli.Context) error {
	resp, err := c.client.IsLoggedIn(context.Background(), &pb.Empty{})
	if err != nil || !resp.GetValue() {
		return formatError(internal.ErrNotLoggedIn)
	}

	payload, err := c.client.TokenInfo(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	switch payload.Type {
	case internal.CodeSuccess:
		break
	case internal.CodeUnauthorized:
		return formatError(internal.ErrNotLoggedIn)
	}

	fmt.Println("Token Information:")
	fmt.Printf("Token: %s\n", payload.Token)
	fmt.Println("Expires at:", payload.ExpiresAt)
	fmt.Printf("Trusted Pass Owner ID: %s\n", payload.TrustedPassOwnerId)

	return nil
}
