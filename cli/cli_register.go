package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// RegisterUsageText is shown next to register command by nordvpn --help
const RegisterUsageText = "Registers a new user account"

func (c *cmd) Register(ctx *cli.Context) error {
	email, password, err := ReadCredentialsFromTerminal()
	if err != nil {
		return formatError(err)
	}

	resp, err := c.client.Register(context.Background(), &pb.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return formatError(err)
	}

	switch resp.Type {
	case internal.CodeSuccess:
		color.Green(AccountCreationSuccess)
	case internal.CodeBadRequest:
		err = errors.New(AccountInvalidData)
	case internal.CodeConflict:
		err = errors.New(AccountEmailTaken)
	case internal.CodeInternalError:
		err = errors.New(AccountInternalError)
	case internal.CodeFailure:
		err = internal.ErrUnhandled
	}

	if err != nil {
		return formatError(err)
	}

	planResp, err := c.client.Plans(context.Background(), &pb.Empty{})
	if err != nil {
		color.Red("Failed to retrieve subscription plans. Please finish the registration in NordVPN website.")
		return browse(client.SubscriptionURL, client.SubscriptionURL)
	}

	plans := planResp.GetPlans()
	// sort plans by cost in ascending order
	sort.Slice(plans, func(i, j int) bool {
		costI := plans[i].GetCost()
		costJ := plans[j].GetCost()
		// this is done to avoid string -> int conversions
		if len(costI) == len(costJ) {
			return costI < costJ
		}
		return len(costI) < len(costJ)
	})
	for i, plan := range plans {
		description := fmt.Sprintf("%s for %s %s", plan.GetTitle(), plan.GetCost(), plan.GetCurrency())
		fmt.Printf("%d) %s\n", i+1, description)
	}

	return browse(client.SubscriptionURL, client.SubscriptionURL)
}
