package recents

import (
	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/core"
)

type Model struct {
	Country            string                     `json:"country"`
	City               string                     `json:"city"`
	Group              config.ServerGroup         `json:"group"`
	CountryCode        string                     `json:"country-code"`
	SpecificServerName string                     `json:"specific-server-name"`
	SpecificServer     string                     `json:"specific-server"`
	ConnectionType     config.ServerSelectionRule `json:"connection-type"`
	ServerTechnologies []core.ServerTechnology    `json:"server-technologies"`
}
