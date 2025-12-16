package recents

import (
	"slices"

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
	IsVirtual          bool                       `json:"is-virtual"`
}

// IsEmpty checks whether the recent connection model is empty
func (m Model) IsEmpty() bool {
	return m.Country == "" &&
		m.City == "" &&
		m.Group == config.ServerGroup_UNDEFINED &&
		m.CountryCode == "" &&
		m.SpecificServerName == "" &&
		m.SpecificServer == "" &&
		m.ConnectionType == config.ServerSelectionRule_NONE &&
		len(m.ServerTechnologies) == 0 &&
		!m.IsVirtual
}

// Clone creates a deep copy of the recent connection model
func (m Model) Clone() Model {
	return Model{
		Country:            m.Country,
		City:               m.City,
		Group:              m.Group,
		CountryCode:        m.CountryCode,
		SpecificServerName: m.SpecificServerName,
		SpecificServer:     m.SpecificServer,
		ConnectionType:     m.ConnectionType,
		ServerTechnologies: slices.Clone(m.ServerTechnologies),
		IsVirtual:          m.IsVirtual,
	}
}

// Equals compares two models for equality
func (m Model) Equals(other Model) bool {
	return m.Country == other.Country &&
		m.City == other.City &&
		m.Group == other.Group &&
		m.CountryCode == other.CountryCode &&
		m.SpecificServerName == other.SpecificServerName &&
		m.SpecificServer == other.SpecificServer &&
		m.ConnectionType == other.ConnectionType &&
		slices.Equal(m.ServerTechnologies, other.ServerTechnologies) &&
		m.IsVirtual == other.IsVirtual
}
