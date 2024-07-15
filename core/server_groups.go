package core

// ServerGroup represents a server group type
//
// This should be in the core package, but cannot be there due to import cycles
// because of dimensions package.
type ServerGroup int64

const (
	// UndefinedGroup represents non existing server group
	UndefinedGroup ServerGroup = 0
	// DoubleVPN represents the double vpn server group
	DoubleVPN ServerGroup = 1
	// OnionOverVPN represents a OnionOverVPN server group
	OnionOverVPN ServerGroup = 3
	// UltraFastTV represents a UltraFastTV server group
	UltraFastTV ServerGroup = 5
	// AntiDDoS represents an AntiDDoS server group
	AntiDDoS ServerGroup = 7
	// DedicatedIP servers represents the Dedicated IP servers
	DedicatedIP ServerGroup = 9
	// StandardVPNServers represents a StandardVPNServers group
	StandardVPNServers ServerGroup = 11
	// NetflixUSA represents a NetflixUSA server group
	NetflixUSA ServerGroup = 13
	// P2P represents a P2P server group
	P2P ServerGroup = 15
	// Obfuscated represents an Obfuscated server group
	Obfuscated ServerGroup = 17
	// Europe servers represents the European servers
	Europe ServerGroup = 19
	// TheAmericas represents TheAmericas servers
	TheAmericas ServerGroup = 21
	// AsiaPacific represents a AsiaPacific server group
	AsiaPacific ServerGroup = 23
	// AfricaMiddleEastIndia represents a Africa, the Middle East and India server group
	AfricaMiddleEastIndia ServerGroup = 25
)

// GroupMap maps group titles to IDs
var GroupMap = map[string]ServerGroup{
	"double_vpn":                       DoubleVPN,
	"onion_over_vpn":                   OnionOverVPN,
	"dedicated_ip":                     DedicatedIP,
	"standard_vpn_servers":             StandardVPNServers,
	"p2p":                              P2P,
	"europe":                           Europe,
	"the_americas":                     TheAmericas,
	"asia_pacific":                     AsiaPacific,
	"africa_the_middle_east_and_india": AfricaMiddleEastIndia,
	"obfuscated_servers":               Obfuscated,
}
