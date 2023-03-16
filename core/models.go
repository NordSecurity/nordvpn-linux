package core

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/slices"
)

// ServerTechnology represents the nordvpn server technology
type ServerTechnology int64

const (
	// Unknown is used for invalid cases
	Unknown ServerTechnology = 0
	// OpenVPNUDP represents the OpenVPN udp technology
	OpenVPNUDP ServerTechnology = 3
	// OpenVPNTCP represents the OpenVpn tcp technolgy
	OpenVPNTCP ServerTechnology = 5
	// Socks5 represents the socks 5 technology
	Socks5 ServerTechnology = 7
	// HTTPProxy represents the http proxy technology
	HTTPProxy ServerTechnology = 9
	// PPTP represents the pptp technology
	PPTP ServerTechnology = 11
	// L2TP represents the l2tp technology
	L2TP ServerTechnology = 13
	// OpenVPNUDPObfuscated represents the openvpn udp obfuscated technology
	OpenVPNUDPObfuscated ServerTechnology = 15
	// OpenVPNTCPObfuscated represents the openvpn tcp obfuscated technology
	OpenVPNTCPObfuscated ServerTechnology = 17
	// WireguardTech represents wireguard technology
	WireguardTech ServerTechnology = 35
)

// ServerGroup represents a server group type
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

type ServerBy int

const (
	ServerByUnknown ServerBy = iota
	ServerBySpeed
	ServerByCountry
	ServerByCity
	ServerByName
)

type ServerTag struct {
	Action ServerBy
	ID     int64
}

type ServersFilter struct {
	Limit int
	Tech  ServerTechnology
	Group config.ServerGroup
	Tag   ServerTag
}

// Status is used by Server and Technology to communicate availability
type Status string

const (
	Online      = "online"
	Offline     = "offline"
	Maintenance = "maintenance"
)

type UserCreateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserCreateResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	ExpiresAt string `json:"password_expires_at"`
	CreateAt  string `json:"create_at"`
	UpdatedAt string `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type oAuth2LoginResponse struct {
	URI     string `json:"redirect_uri"`
	Attempt string `json:"attempt"`
}

type LoginResponse struct {
	UserID     int64  `json:"user_id"`
	Token      string `json:"token"`
	RenewToken string `json:"renew_token"`
	ExpiresAt  string `json:"expires_at"`
	UpdatedAt  string `json:"updated_at"`
	CreatedAt  string `json:"created_at"`
	ID         int64  `json:"id"`
}

type CredentialsResponse struct {
	ID                 int64  `json:"id"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	NordlynxPrivateKey string `json:"nordlynx_private_key"`
}

type ServicesResponse []ServiceData

type ServiceData struct {
	ID        int64   `json:"ID"`
	ExpiresAt string  `json:"expires_at"`
	Service   Service `json:"service"`
}

type Service struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CurrentUserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type TokenRenewResponse struct {
	Token      string `json:"token"`
	RenewToken string `json:"renew_token"`
	ExpiresAt  string `json:"expires_at"`
}

type Plans []Plan

type Plan struct {
	ID         int64  `json:"id"`
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Cost       string `json:"cost"`
	Currency   string `json:"currency"`
}

type Technologies []Technology

type Technology struct {
	ID       ServerTechnology `json:"id"`
	Pivot    Pivot            `json:"pivot"`
	Metadata []struct {
		Name  string      `json:"name,omitempty"`
		Value interface{} `json:"value,omitempty"`
	} `json:"metadata"`
}

func (t Technology) IsOnline() bool {
	return t.Pivot.Status == Online
}

type Servers []Server

type ServerIPRecord struct {
	ServerIP `json:"ip"`
	Type     string `json:"type"`
}

type ServerIP struct {
	IP      string `json:"ip"`
	Version uint8  `json:"version"`
}

type Server struct {
	ID                int64  `json:"id"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	Name              string `json:"name"`
	Station           string `json:"station"`
	Hostname          string `json:"hostname"`
	Load              int64  `json:"load"`
	Status            Status `json:"status"`
	Locations         `json:"locations"`
	Technologies      Technologies `json:"technologies"`
	Groups            `json:"groups"`
	Specifications    []Specification  `json:"specifications"`
	Distance          float64          `json:"distance"`
	Timestamp         int64            `json:"timestamp"`
	Penalty           float64          `json:"penalty"`
	PartialPenalty    float64          `json:"partial_penalty"`
	NordLynxPublicKey string           `json:"-"`
	Keys              []string         `json:"-"`
	IPRecords         []ServerIPRecord `json:"ips"`
}

// ServerObfuscationStatus is the return status of IsServerObfuscated
type ServerObfuscationStatus int

const (
	// ServerObfuscated status returned when server is obfuscated
	ServerObfuscated ServerObfuscationStatus = iota
	// ServerNotObfuscated status returned when server is not obfuscated
	ServerNotObfuscated
	// NotAServerName returned when server with such name has not been found
	// (there is no hostname beggining with given server tag)
	NotAServerName
)

// IsServerObfuscated returns ServerObfuscationStatus for a given server tag
func IsServerObfuscated(servers Servers, serverTag string) ServerObfuscationStatus {
	serverIndex := slices.IndexFunc(servers, func(server Server) bool {
		serverName := strings.Split(server.Hostname, ".")[0]
		return serverName == serverTag
	})

	if serverIndex == -1 {
		return NotAServerName
	}

	if IsObfuscated()(servers[serverIndex]) {
		return ServerObfuscated
	}

	return ServerNotObfuscated
}

// Predicate function used in algorithms like filter.
type Predicate func(Server) bool

// ByGroup is a Comparison function meant for use with
// github.com/NordSecurity/nordvpn-linux/slices.ContainsFunc function.
func ByGroup(s config.ServerGroup) func(Group) bool {
	return func(g Group) bool { return g.ID == s }
}

// ByTag is a Comparison function meant for use with
// github.com/NordSecurity/nordvpn-linux/slices.ContainsFunc function.
func ByTag(tag string) func(Group) bool {
	return func(g Group) bool {
		group := strings.ToLower(strings.ReplaceAll(g.Title, " ", "_"))
		return strings.EqualFold(tag, group)
	}
}

// IsOnline returns true for online servers.
func IsOnline() Predicate {
	return func(s Server) bool {
		return s.Status == Online
	}
}

// IsConnectableVia returns true if it's possible to connect to server using a given technology.
func IsConnectableVia(tech ServerTechnology) Predicate {
	return func(s Server) bool {
		for _, technology := range s.Technologies {
			if technology.ID == tech && technology.IsOnline() {
				return IsOnline()(s)
			}
		}
		return false
	}
}

// IsObfuscated returns a filter for keeping only obfuscated servers.
func IsObfuscated() Predicate {
	return func(s Server) bool {
		return IsConnectableVia(OpenVPNUDPObfuscated)(s) &&
			IsConnectableVia(OpenVPNTCPObfuscated)(s)
	}
}

// IsConnectableWithProtocol behaves like IsConnectableVia, but also includes protocol.
func IsConnectableWithProtocol(tech config.Technology, proto config.Protocol) Predicate {
	return func(s Server) bool {
		switch tech {
		case config.Technology_NORDLYNX:
			return IsConnectableVia(WireguardTech)(s)
		case config.Technology_OPENVPN:
			if proto == config.Protocol_UDP {
				return IsConnectableVia(OpenVPNUDP)(s) ||
					IsConnectableVia(OpenVPNUDPObfuscated)(s)
			}
			if proto == config.Protocol_TCP {
				return IsConnectableVia(OpenVPNTCP)(s) ||
					IsConnectableVia(OpenVPNTCPObfuscated)(s)
			}
		case config.Technology_UNKNOWN_TECHNOLOGY:
			break
		}
		return false
	}
}

func (s *Server) Version() string {
	for _, spec := range s.Specifications {
		if spec.Identifier == "version" {
			return spec.Identifier
		}
	}
	return ""
}

func (s *Server) SupportsIPv6() bool {
	for _, ip := range s.IPs() {
		if ip.Is6() {
			return true
		}
	}
	return false
}

func (s *Server) IPs() []netip.Addr {
	var serverIPs []netip.Addr
	for _, record := range s.IPRecords {
		ip, err := netip.ParseAddr(record.ServerIP.IP)
		if err == nil {
			serverIPs = append(serverIPs, ip)
		}
	}
	if serverIPs == nil {
		ip, err := s.IPv4()
		if err == nil {
			serverIPs = append(serverIPs, ip)
		}
	}
	return serverIPs
}

func (s *Server) IPv4() (netip.Addr, error) {
	return netip.ParseAddr(s.Station)
}

func (s *Server) UnmarshalJSON(b []byte) error {
	// https://stackoverflow.com/questions/52433467/how-to-call-json-unmarshal-inside-unmarshaljson-without-causing-stack-overflow
	type Hack Server
	var hack Hack

	if err := json.Unmarshal(b, &hack); err != nil {
		return err
	}

	for i, tech := range hack.Technologies {
		for _, meta := range tech.Metadata {
			var value string
			value, ok := meta.Value.(string)
			if !ok {
				continue
			}
			if meta.Name == "public_key" {
				trimmed := strings.TrimSpace(value)
				if tech.ID == WireguardTech {
					hack.NordLynxPublicKey = trimmed
				}
				break
			}
		}
		// gob ignores nil fields
		hack.Technologies[i].Metadata = nil
	}
	*s = Server(hack)

	return nil
}

type Groups []Group

type Group struct {
	ID    config.ServerGroup `json:"id"`
	Title string             `json:"title"`
}

type Specification struct {
	Identifier string `json:"identifier"`
	Values     []struct {
		Value string `json:"value"`
	} `json:"values"`
}

type Locations []Location

func (l Locations) Country() (Country, error) {
	if len(l) == 0 {
		return Country{}, fmt.Errorf("no countries in specified location")
	}
	return l[0].Country, nil
}

type Location struct {
	Country `json:"country"`
}

type Countries []Country

// Country is a weird struct in that it is defined
// in two different ways by the backend. Server
// recommendations endpoint response has city field, while
// server countries endpoint response has cities field.
// Basically, only one of the city/cities field
// exists at a given time.
type Country struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	City   `json:"city,omitempty"`
	Cities `json:"cities,omitempty"`
}

type Cities []City

type City struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	HubScore  *float64 `json:"hub_score"`
}

type Pivot struct {
	Status Status `json:"status"`
}

type Insights struct {
	CountryCode string  `json:"country_code"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
}

type NameServers struct {
	Servers []string `json:"servers"`
}
