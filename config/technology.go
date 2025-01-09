package config

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
		return ""
	}
	return ""
}
