// Package events provides publisher-subscriber interfaces.
package events

import (
	"net/http"
	"net/netip"
	"time"

	"github.com/NordSecurity/nordvpn-linux/config"
	"github.com/NordSecurity/nordvpn-linux/daemon/state/types"
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

type DataAllowlist struct {
	Subnets  []string
	TCPPorts []int64
	UDPPorts []int64
}

type DataDNS struct {
	Ips []string
}

type TypeEventStatus int

const (
	StatusAttempt TypeEventStatus = iota
	StatusSuccess
	StatusFailure
	StatusCanceled
)

type TypeEventTrigger int

const (
	TriggerApp TypeEventTrigger = iota
	TriggerUser
)

type TypeLoginType int

const (
	LoginLogin  TypeLoginType = iota // regular login
	LoginSignUp                      // login after signup
)

type DataConnect struct {
	IsMeshnetPeer           bool
	ThreatProtectionLite    bool
	Protocol                config.Protocol
	DurationMs              int
	ServerFromAPI           bool
	EventStatus             TypeEventStatus
	TargetServerSelection   config.ServerSelectionRule
	Technology              config.Technology
	TargetServerCity        string
	TargetServerCountry     string
	TargetServerCountryCode string
	TargetServerDomain      string
	TargetServerGroup       string
	TargetServerIP          netip.Addr
	TargetServerName        string
	Error                   error
	IsVirtualLocation       bool
	IsObfuscated            bool
	IsPostQuantum           bool
}

// DataConnectChangeNotif is used to provide notifications for internal listeners of ConnectionStatus
type DataConnectChangeNotif struct {
	Status types.ConnectionStatus
}

type DataDisconnect struct {
	Protocol              config.Protocol
	ServerFromAPI         bool
	EventStatus           TypeEventStatus
	Technology            config.Technology
	TargetServerSelection config.ServerSelectionRule
	ThreatProtectionLite  bool
	ByUser                bool
	Duration              time.Duration
	Error                 error
}

type DataAuthorization struct {
	DurationMs   int
	EventTrigger TypeEventTrigger
	EventStatus  TypeEventStatus
	EventType    TypeLoginType
}

type DataRequestAPI struct {
	// Note: Never use `Request.Body`, use `Request.GetBody` instead
	Request *http.Request
	// Note: In case you read `Response.Body`, make sure it is set to what it was before
	Response *http.Response
	Duration time.Duration
	Error    error
	// IsAttempt indicates whether the event represents an attempt
	IsAttempt bool
}

// Analytics analytics handling engine interface
type Analytics interface {
	Enable() error
	Disable() error
}

// UiItemsAction stores arguments to moose.NordvpnappSendUserInterfaceUiItemsClick
type UiItemsAction struct {
	ItemName      string
	ItemType      string
	ItemValue     string
	FormReference string
}
