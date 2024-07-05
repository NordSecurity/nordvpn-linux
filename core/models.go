package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
	"strings"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/internal"
	"github.com/NordSecurity/nordvpn-linux/nstrings"
	"golang.org/x/exp/slices"
)

// ServerTechnology represents the nordvpn server technology
type ServerTechnology int64

const (
	// Unknown is used for invalid cases
	Unknown ServerTechnology = 0
	// OpenVPNUDP represents the OpenVPN udp technology
	OpenVPNUDP ServerTechnology = 3
	// OpenVPNTCP represents the OpenVpn tcp technology
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
	Online          = "online"
	Offline         = "offline"
	Maintenance     = "maintenance"
	VirtualLocation = "virtual_location"
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

type TrustedPassTokenResponse struct {
	OwnerID string `json:"owner_id"`
	Token   string `json:"token"`
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
	// (there is no hostname beginning with given server tag)
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
	version := s.getSpecificationsForIdentifier("version")

	if len(version) > 0 {
		return version[0]
	}
	return ""
}

func (s *Server) IsVirtualLocation() bool {
	virtualLocation := s.getSpecificationsForIdentifier(VirtualLocation)

	if len(virtualLocation) > 0 {
		value, err := nstrings.BoolFromString(virtualLocation[0])
		if err == nil {
			return value
		}
		log.Println(internal.DebugPrefix, "cannot convert server virtual location", s.Hostname, virtualLocation, err)
	}
	return false
}

func (s *Server) Country() *Country {
	if len(s.Locations) > 0 {
		return &s.Locations[0].Country
	}
	return nil
}

func (s *Server) getSpecificationsForIdentifier(identifier string) []string {
	for _, spec := range s.Specifications {
		if spec.Identifier == identifier {
			values := []string{}
			for _, value := range spec.Values {
				values = append(values, value.Value)
			}
			return values
		}
	}
	return nil
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
