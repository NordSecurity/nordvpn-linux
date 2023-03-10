package vpn

import (
	"fmt"
)

// State type represents valid openvpn states type
type State string

// Substate type represents custom openvpn sub types
type Substate string

const (
	// UnknownState is reported when state middleware cannot parse state from string (i.e. it's undefined in list above),
	// usually that means that newer openvpn version reports something extra
	UnknownState State = "UNKNOWN"

	// ConnectingState is reported by client and server mode and is indicator of openvpn startup
	ConnectingState State = "CONNECTING"

	// WaitState is reported by client in udp mode indicating that connect request is send and response is waiting
	WaitState State = "WAIT"

	// AuthenticatingState is reported by client indicating that client is trying to authenticate itself to server
	AuthenticatingState State = "AUTH"

	// GetConfigState indicates that client is waiting for config from server (push based options)
	GetConfigState State = "GET_CONFIG"

	// AssignIPState indicates that client is trying to setup tunnel with provided ip addresses
	AssignIPState State = "ASSIGN_IP"

	// AddRoutesState indicates that client is setuping routes on tunnel
	AddRoutesState State = "ADD_ROUTES"

	// ConnectedState is reported by both client and server and means that client is successfully connected and server is ready
	// to server incoming client connect requests
	ConnectedState State = "CONNECTED"

	// ReconnectingState indicates that client lost connection and is trying to reconnect itself
	ReconnectingState State = "RECONNECTING"

	// ExitingState is reported by both client and server and means that openvpn process is exiting by any reasons (normal shutdown
	// or fatal error reported before this state)
	ExitingState State = "EXITING"

	// ExitedState fake openvpn state which indicated that openvpn has been shutdown
	ExitedState State = "EXITED"
)

const (
	UnknownSubstate      Substate = "UNKNOWN"
	AuthFlukeSubstate    Substate = "AUTH_FLUKE"
	AuthBadSubstate      Substate = "AUTH_BAD"
	TimeoutFlukeSubstate Substate = "TIMEOUT_FLUKE"
	TimeoutSubstate      Substate = "TIMEOUT"
)

var stateMap = map[State]bool{
	ConnectingState:     true,
	WaitState:           true,
	AuthenticatingState: true,
	GetConfigState:      true,
	AssignIPState:       true,
	AddRoutesState:      true,
	ConnectedState:      true,
	ReconnectingState:   true,
	ExitingState:        true,
	ExitedState:         true,
}

func StringToState(arg string) (State, error) {
	state := State(arg)
	if _, ok := stateMap[state]; ok {
		return state, nil
	}
	return UnknownState, fmt.Errorf("unknown state: '%s'", arg)
}
