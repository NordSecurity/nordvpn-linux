package recents

import (
	"github.com/NordSecurity/nordvpn-linux/config"
)

type Model struct {
	Country            string                     `json:"country,omitempty"`
	City               string                     `json:"city,omitempty"`
	Group              config.ServerGroup         `json:"group,omitempty"`
	CountryCode        string                     `json:"country-code,omitempty"`
	SpecificServerName string                     `json:"specific-server-name,omitempty"`
	SpecificServer     string                     `json:"specific-server,omitempty"`
	ConnectionType     config.ServerSelectionRule `json:"connection-type,omitempty"`
	IsVirtual          bool                       `json:"is-virtual,omitempty"`
	ConnectionTech     config.Technology          `json:"connection-tech,omitempty"`
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
		IsVirtual:          m.IsVirtual,
		ConnectionTech:     m.ConnectionTech,
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
		m.IsVirtual == other.IsVirtual &&
		m.AreTechCompatible(other.ConnectionTech)
}

// AreTechCompatible - compares the model ConnectionTech with the given technology parameter.
// The function returns true if both are NordWhisper or both are not NordWhisper, not necessary same technology.
// This is because recent connections having NordWhisper technology are marked obfuscated, while all the others are not.
func (m Model) AreTechCompatible(tech config.Technology) bool {
	// only NordWhisper tech is treated differently. All the others are considered compatible
	return (tech == config.Technology_NORDWHISPER) == (m.ConnectionTech == config.Technology_NORDWHISPER)
}
