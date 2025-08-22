package recents

import "github.com/NordSecurity/nordvpn-linux/config"

type Model struct {
	Country            string
	City               string
	SpecificServer     string
	SpecificServerName string
	Group              string
	ConnectionType     config.ServerSelectionRule
}
