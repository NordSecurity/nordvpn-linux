package config

import "strconv"

// Technology is the VPN technology used for a connection
type Technology int32

const (
	TechnologyUnknown     Technology = 0
	TechnologyOpenVPN     Technology = 1
	TechnologyNordLynx    Technology = 2
	TechnologyNordWhisper Technology = 3
)

func (t Technology) String() string {
	switch t {
	case TechnologyUnknown:
		return "UNKNOWN_TECHNOLOGY"
	case TechnologyOpenVPN:
		return "OPENVPN"
	case TechnologyNordLynx:
		return "NORDLYNX"
	case TechnologyNordWhisper:
		return "NORDWHISPER"
	default:
		return strconv.Itoa(int(t))
	}
}

// TechNameToUpperCamelCase returns technology name as an UpperCamelCase string
func TechNameToUpperCamelCase(tech Technology) string {
	switch tech {
	case TechnologyNordLynx:
		return "NordLynx"
	case TechnologyOpenVPN:
		return "OpenVPN"
	case TechnologyNordWhisper:
		return "NordWhisper"
	case TechnologyUnknown:
		fallthrough
	default:
		return ""
	}
}
