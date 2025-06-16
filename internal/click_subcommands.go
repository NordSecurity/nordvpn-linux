package internal

import "fmt"

const (
	NordVPNScheme                 = "nordvpn"
	ClaimOnlinePurchaseSubcommand = "claim-online-purchase"
	LoginSubcommand               = "login"
	ConsentSubcommand             = "consent"
)

func SubcommandURI(subcommand string) string {
	return fmt.Sprintf("%s://%s", NordVPNScheme, subcommand)
}
