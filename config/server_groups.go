package config

// GroupMap maps group titles to IDs
var GroupMap = map[string]ServerGroup{
	"double_vpn":           ServerGroup_DOUBLE_VPN,
	"onion_over_vpn":       ServerGroup_ONION_OVER_VPN,
	"dedicated_ip":         ServerGroup_DEDICATED_IP,
	"standard_vpn_servers": ServerGroup_STANDARD_VPN_SERVERS,
	"p2p":                  ServerGroup_P2P,
	"obfuscated_servers":   ServerGroup_OBFUSCATED,
}

// IsRegionalGroup reports whether g is a deprecated regional group; uses raw ints since the named constants are removed.
func IsRegionalGroup(g ServerGroup) bool {
	//exhaustive:ignore [only regional group IDs are relevant]
	switch g {
	case 19, 21, 23, 25: // EUROPE, THE_AMERICAS, ASIA_PACIFIC, AFRICA_THE_MIDDLE_EAST_AND_INDIA
		return true
	}
	return false
}

// GroupTitleForId converts group ID to group lowercase title
func GroupTitleForId(group ServerGroup) string {
	for k, v := range GroupMap {
		if v == group {
			return k
		}
	}

	return ""
}
