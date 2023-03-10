// Package events provides publisher-subscriber interfaces.
package events

import (
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
)

// Handler is used to process messages.
type Handler[T any] func(T) error

// Publisher allows publishing messages of type T.
type Publisher[T any] interface {
	Publish(message T)
}

// Subscriber listens to messages of type T.
type Subscriber[T any] interface {
	// Subscribe allows registering multiple handlers for the same message.
	Subscribe(Handler[T])
}

// PublishSubcriber allows both publishing and subscribing to messages of type T.
type PublishSubcriber[T any] interface {
	Publisher[T]
	Subscriber[T]
}

type DataWhitelist struct {
	Subnets  int
	TCPPorts int
	UDPPorts int
}

type DataDNS struct {
	Enabled bool
	Ips     []string
}

type TypeConnect int

const (
	ConnectAttempt TypeConnect = iota
	ConnectSuccess
	ConnectFailure
)

type TypeDisconnect int

const (
	DisconnectAttempt TypeDisconnect = iota
	DisconnectSuccess
	DisconnectFailure
)

type DataConnect struct {
	APIHostname                string
	Auto                       bool
	ThreatProtectionLite       bool
	DNSResolutionTime          time.Duration
	Protocol                   config.Protocol
	ResponseServersCount       int
	ResponseTime               int
	ServerFromAPI              bool
	Type                       TypeConnect
	TargetServerSelection      string
	Technology                 config.Technology
	TargetServerCity           string
	TargetServerCountry        string
	TargetServerDomain         string
	TargetServerGroup          string
	TargetServerIP             string
	TargetServerPick           string
	TargetServerPickerResponse string
}

type DataDisconnect struct {
	Protocol              config.Protocol
	ServerFromAPI         bool
	Type                  TypeDisconnect
	Technology            config.Technology
	TargetServerSelection string
	ThreatProtectionLite  bool
}

type DataRequestAPI struct {
	Endpoint         string
	Hostname         string
	DNSResolution    time.Duration
	EventDuration    time.Duration
	IsSuccessful     bool
	RequestLimits    string
	RequestOffset    string
	RequestFields    string
	RequestFilters   string
	ResponseCode     int
	ResponseSummary  string
	TransferProtocol string
}

// Analytics analytics handling engine interface
type Analytics interface {
	Enable() error
	Disable() error
}

// ServerRating last used server rating info
type ServerRating struct {
	Rate   int
	Server string
}
