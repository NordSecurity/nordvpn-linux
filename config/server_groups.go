package config

// GroupMap maps group titles to IDs
var GroupMap = map[string]ServerGroup{
	"double_vpn":                       ServerGroup_DOUBLE_VPN,
	"onion_over_vpn":                   ServerGroup_ONION_OVER_VPN,
	"dedicated_ip":                     ServerGroup_DEDICATED_IP,
	"standard_vpn_servers":             ServerGroup_STANDARD_VPN_SERVERS,
	"p2p":                              ServerGroup_P2P,
	"europe":                           ServerGroup_EUROPE,
	"the_americas":                     ServerGroup_THE_AMERICAS,
	"asia_pacific":                     ServerGroup_ASIA_PACIFIC,
	"africa_the_middle_east_and_india": ServerGroup_AFRICA_THE_MIDDLE_EAST_AND_INDIA,
	"obfuscated_servers":               ServerGroup_OBFUSCATED,
}
