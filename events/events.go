// Package events provides publisher-subscriber interfaces.
package events

import (
	"net/http"
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
)

type TypeEventTrigger int

const (
	TriggerApp TypeEventTrigger = iota
	TriggerUser
)

type DataConnect struct {
	IsMeshnetPeer              bool
	APIHostname                string
	Auto                       bool
	ThreatProtectionLite       bool
	Protocol                   config.Protocol
	ResponseServersCount       int
	ResponseTime               int
	DurationMs                 int
	ServerFromAPI              bool
	EventStatus                TypeEventStatus
	TargetServerSelection      string
	Technology                 config.Technology
	TargetServerCity           string
	TargetServerCountry        string
	TargetServerDomain         string
	TargetServerGroup          string
	TargetServerIP             string
	TargetServerPick           string
	TargetServerPickerResponse string
	TargetServerName           string
	Error                      error
}

type DataDisconnect struct {
	Protocol              config.Protocol
	ServerFromAPI         bool
	EventStatus           TypeEventStatus
	Technology            config.Technology
	TargetServerSelection string
	ThreatProtectionLite  bool
}

type DataAuthorization struct {
	DurationMs   int
	EventTrigger TypeEventTrigger
	EventStatus  TypeEventStatus
}

type DataRequestAPI struct {
	// Note: Never use `Request.Body`, use `Request.GetBody` instead
	Request *http.Request
	// Note: In case you read `Response.Body`, make sure it is set to what it was before
	Response *http.Response
	Duration time.Duration
	Error    error
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
