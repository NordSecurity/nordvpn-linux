package config

import "strconv"

// Technology is the VPN technology used for a connection
type Technology int32

const (
	Technology_UNKNOWN_TECHNOLOGY Technology = 0
	Technology_OPENVPN            Technology = 1
	Technology_NORDLYNX           Technology = 2
	Technology_NORDWHISPER        Technology = 3
)

func (t Technology) String() string {
	switch t {
	case Technology_UNKNOWN_TECHNOLOGY:
		return "UNKNOWN_TECHNOLOGY"
	case Technology_OPENVPN:
		return "OPENVPN"
	case Technology_NORDLYNX:
		return "NORDLYNX"
	case Technology_NORDWHISPER:
		return "NORDWHISPER"
	default:
		return strconv.Itoa(int(t))
	}
}

// TechNameToUpperCamelCase returns technology name as an UpperCamelCase string
func TechNameToUpperCamelCase(tech Technology) string {
	switch tech {
	case Technology_NORDLYNX:
		return "NordLynx"
	case Technology_OPENVPN:
		return "OpenVPN"
	case Technology_NORDWHISPER:
		return "NordWhisper"
	case Technology_UNKNOWN_TECHNOLOGY:
		fallthrough
	default:
		return ""
	}
}
