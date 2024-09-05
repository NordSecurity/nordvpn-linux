package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

// AccountUsageText is shown next to account command by nordvpn --help
const AccountUsageText = "Shows account information"

func displayServiceStatus(serviceName string, serviceStatus int64, expiry string) error {
	switch serviceStatus {
	case internal.CodeSuccess:
		expiryTime, err := time.Parse(internal.ServerDateFormat, expiry)
		if err != nil {
			return formatError(fmt.Errorf(AccountCantFetchVPNService, serviceName))
		}

		expiryString := fmt.Sprintf("%s %s, %d",
			expiryTime.Month().String()[0:3], ordinal(expiryTime.Day()), expiryTime.Year())
		fmt.Printf("%s Service: Active (Expires on %s)\n", serviceName, expiryString)
	case internal.CodeNoService:
		fmt.Printf("%s: Inactive\n", serviceName)
	}

	return nil
}

func (c *cmd) Account(ctx *cli.Context) error {
	payload, err := c.client.AccountInfo(context.Background(), &pb.Empty{})
	if err != nil {
		return formatError(err)
	}

	switch payload.Type {
	case internal.CodeUnauthorized:
		return formatError(errors.New(AccountTokenUnauthorizedError))
	case internal.CodeExpiredRenewToken:
		color.Yellow(client.RelogRequest)
		err = c.Login(ctx)
		if err != nil {
			return formatError(err)
		}
		return c.Account(ctx)
	case internal.CodeTokenRenewError:
		return formatError(errors.New(client.AccountTokenRenewError))
	}

	fmt.Println("Account Information:")
	if payload.Username != "" {
		fmt.Printf("Username: %s\n", payload.Username)
	}
	fmt.Println("Email Address:", payload.Email)

	if err := displayServiceStatus("VPN", payload.Type, payload.ExpiresAt); err != nil {
		return err
	}

	if err := displayServiceStatus("Dedicated IP",
		payload.DedicatedIpStatus,
		payload.LastDedicatedIpExpiresAt); err != nil {
		return err
	}

	mfa := "enabled"
	if payload.MfaStatus == pb.TriState_DISABLED {
		mfa = "disabled"
	} else if payload.MfaStatus == pb.TriState_UNKNOWN {
		mfa = "unknown"
	}

	fmt.Println("Multi-factor Authentication (MFA):", mfa)

	return nil
}

func ordinal(day int) string {
	switch day {
	case 1, 21, 31:
		return strconv.Itoa(day) + "st"
	case 2, 22:
		return strconv.Itoa(day) + "nd"
	case 3, 23:
		return strconv.Itoa(day) + "rd"
	default:
		return strconv.Itoa(day) + "th"
	}
}

func activeBoolToString(isActive bool) string {
	if isActive {
		return "Active"
	}
	return "Inactive"
}
