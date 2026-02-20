package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NordSecurity/nordvpn-linux/client"
	"github.com/NordSecurity/nordvpn-linux/daemon/pb"
	"github.com/NordSecurity/nordvpn-linux/internal"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const (
	// AccountUsageText is shown next to account command by nordvpn --help
	AccountUsageText    = "Shows account information"
	TermsOfServiceURL   = "https://my.nordaccount.com/legal/terms-of-service/?utm_medium=app&utm_source=nordvpn-linux-cli"
	AutoRenewalTermsURL = "https://my.nordaccount.com/legal/terms-of-service/subscription/?utm_medium=app&utm_source=nordvpn-linux-cli"
	PrivacyPolicyURL    = "https://my.nordaccount.com/legal/privacy-policy/?utm_medium=app&utm_source=nordvpn-linux-cli"
)

func formatDate(dateStr string) (string, error) {
	t, err := time.Parse(internal.ServerDateFormat, dateStr)
	if err != nil {
		return "", err
	}
	return t.Format("Jan 2, 2006"), nil
}

func displayExpiryInfo(name string, status int64, expiry string) error {
	switch status {
	case internal.CodeSuccess:
		expiryString, err := formatDate(expiry)
		if err != nil {
			return formatError(fmt.Errorf(AccountCantFetchVPNService, name))
		}
		fmt.Printf("%s: Active until %s\n", name, expiryString)
	case internal.CodeNoService:
		fmt.Printf("%s: Inactive\n", name)
	}

	return nil
}

// Account displays account information
func (c *cmd) Account(ctx *cli.Context) error {
	payload, err := c.client.AccountInfo(context.Background(), &pb.AccountRequest{Full: true})
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
	case internal.CodeExpiredAccessToken:
		fallthrough
	case internal.CodeRevokedAccessToken:
		return formatError(errors.New(client.AccessTokenExpired))
	}

	fmt.Println("Account information")
	if payload.Username != "" {
		fmt.Printf("Username: %s\n", payload.Username)
	}
	fmt.Println("Email address:", payload.Email)

	createdOnFormatted, err := formatDate(payload.CreatedOn)
	if err != nil {
		return formatError(fmt.Errorf("failed to parse account creation date"))
	}
	fmt.Println("Account created:", createdOnFormatted)

	if err := displayExpiryInfo("Subscription", payload.Type, payload.SubscriptionExpiresAt); err != nil {
		return err
	}

	if err := displayExpiryInfo("Dedicated IP",
		payload.DedicatedIpStatus,
		payload.LastDedicatedIpExpiresAt); err != nil {
		return err
	}

	var mfa string
	switch payload.MfaStatus {
	case pb.TriState_ENABLED:
		mfa = "Turned on"
	case pb.TriState_DISABLED:
		mfa = "Turned off"
	case pb.TriState_UNKNOWN:
		mfa = "unknown"
	}

	fmt.Println("Multi-factor authentication (MFA):", mfa)
	fmt.Println("\nTerms of Service -", TermsOfServiceURL)
	fmt.Println("Auto-renewal terms -", AutoRenewalTermsURL)
	fmt.Println("Privacy Policy -", PrivacyPolicyURL)

	return nil
}
