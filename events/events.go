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

type ContextValue struct {
	Path  string
	Value any
}

// MooseDebuggerEvent represents a debugging event to be sent to the moose library.
// It contains data and context information needed for debugging purposes.
type MooseDebuggerEvent struct {
	// JsonData contains a custom payload to be carried within a debugger event to moose.
	// This allows for attaching specialized information relevant to the specific event.
	JsonData string

	// GeneralContextPaths is a collection of existing paths to reuse in the event context.
	// These paths reference predefined context locations already established in the system.
	GeneralContextPaths []string

	// KeyBasedContextPaths allows defining custom context paths with associated values.
	// Unlike GeneralContextPaths, this field enables creating new, event-specific context
	// paths with custom values rather than reusing existing ones.
	KeyBasedContextPaths []ContextValue
}

// WithJsonData adds JSON payload to the event
func (e *MooseDebuggerEvent) WithJsonData(json string) *MooseDebuggerEvent {
	e.JsonData = json
	return e
}

// WithKeyBasedContextPaths adds arbitrary number of key-based context paths to the event
func (e *MooseDebuggerEvent) WithKeyBasedContextPaths(paths ...ContextValue) *MooseDebuggerEvent {
	e.KeyBasedContextPaths = append(e.KeyBasedContextPaths, paths...)
	return e
}

// WithGlobalContextPaths adds arbitrary number of global context paths to the event
func (e *MooseDebuggerEvent) WithGlobalContextPaths(paths ...string) *MooseDebuggerEvent {
	e.GeneralContextPaths = append(e.GeneralContextPaths, paths...)
	return e
}

// NewMooseDebuggerEvent creates and initializes a new MooseDebuggerEvent instance.
// It takes a JSON data string as input and initializes the event with empty slices
// for GeneralContextPaths and KeyBasedContextPaths.
//
// Parameters:
//   - jsonData: A string containing JSON data to be associated with the event
func NewMooseDebuggerEvent(jsonData string) *MooseDebuggerEvent {
	return &MooseDebuggerEvent{
		JsonData:             jsonData,
		GeneralContextPaths:  make([]string, 0),
		KeyBasedContextPaths: make([]ContextValue, 0),
	}
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

type ReasonCode int32

const (
	// no exceptions
	ReasonNone                         ReasonCode = 0
	ReasonCorruptedVPNCredsAuthBad     ReasonCode = 3010400
	ReasonCorruptedVPNCredsAuthMissing ReasonCode = 3020400
	ReasonAuthTokenBad                 ReasonCode = 3010000
	ReasonTokenMissing                 ReasonCode = 3020000 // Renew token endpoint returned 404
	ReasonAuthTokenInvalidated         ReasonCode = 3030100
	ReasonCorruptedVPNCreds            ReasonCode = 3000400

	ReasonTokenCorrupted ReasonCode = 3040000

	ReasonAuthTokenMissing ReasonCode = 3020100 // 401 from authorized API request

	ReasonIdempotentDataFailedAuthMissing ReasonCode = 3020005

	// Renew token endpoint returned 400, 401 response from authorized API request
	// ReasonUnauthorized ReasonCode = 3010100

	// Credential/generic errors
	ReasonResourceMissing ReasonCode = 3000004 // Could Not Find Credentials on VPN Creation [unexpected]

	// Notification
	ReasonSilentEvent ReasonCode = 3000200 // Silent in app notification [expected]

	// Endpoint/idempotency/generic flows

	ReasonKeyExpired ReasonCode = 3020002 // Idempotency Key Expired After 30 Minutes [unexpected]
	ReasonOpTimeout  ReasonCode = 3020005 // Failed to Fetch Idempotent Data [unexpected]

	// Authorization/error chains

	AuthReasonKeyExpired ReasonCode = 3020102 // Idempotency Key Expired, plus 401
	AuthReasonOpFailure  ReasonCode = 3020105 // Idempotency/authorization/data fetch [unexpected]
)

type DataAuthorization struct {
	DurationMs   int
	EventTrigger TypeEventTrigger
	EventStatus  TypeEventStatus
	EventType    TypeLoginType
	Reason       ReasonCode
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
